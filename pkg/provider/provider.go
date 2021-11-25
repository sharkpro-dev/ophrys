package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type BinanceProvider struct {
	id         string
	host       string
	port       int32
	connection *websocket.Conn
	messages   chan *BinanceMessage
	ctx        context.Context
	done       chan struct{}
	finished   bool
}

type BinanceMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int64    `json:"id"`
}

func NewBinanceProvider(host string, port int32) *BinanceProvider {
	return &BinanceProvider{id: "BinanceProvider", host: host, port: port, done: make(chan struct{}), messages: make(chan *BinanceMessage, 30), finished: false}
}

func (bp *BinanceProvider) Provide(c chan map[string]interface{}, ctx context.Context) {
	bp.ctx = ctx

	binanceUrl := fmt.Sprintf("%s:%d", bp.host, bp.port)

	u := url.URL{Scheme: "wss", Host: binanceUrl, Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	connection, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		log.Fatal("FATAL dial:", err)
	}

	bp.connection = connection

	bp.connection.SetCloseHandler(func(code int, text string) error {
		log.Printf("CloseHandle: %d |  %s", code, text)
		close(bp.done)
		return nil
	})

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Printf("ReadMessage: %s", err.Error())
				return
			}
			log.Printf("%s", message)

			var result map[string]interface{}
			json.Unmarshal(message, &result)

			if !bp.finished {
				c <- result
			}
		}
	}()

	go func() {
		for {
			select {
			case <-bp.done:
				log.Println("Connection close done")
				return
			case <-bp.ctx.Done():
				log.Println("provider interrupted by context")
				bp.finished = true

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close", err)
					return
				}

				log.Println("WriteMessage of closing")
				t := time.NewTicker(time.Second * 2)
				select {
				case <-bp.done:
					log.Println("Cleanly closed the connection")
					return
				case <-t.C:
					log.Println("Timeout")
					close(bp.done)
					return

				}
			case message := <-bp.messages:
				err := bp.connection.WriteJSON(message)
				if err != nil {
					log.Fatal("FATAL: read message:", err)
				}
			}
		}
	}()
}

func (bp *BinanceProvider) Id() string {
	return bp.id
}

func (bp *BinanceProvider) Subscribe(path string) {
	bp.messages <- &BinanceMessage{Method: "SUBSCRIBE", Params: []string{path}, Id: 1}

}
