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
	done       chan struct{}
	done2      chan struct{}
	interrupt  chan struct{}
	messages   chan *BinanceMessage
}

type BinanceMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int64    `json:"id"`
}

func NewBinanceProvider(host string, port int32) *BinanceProvider {
	return &BinanceProvider{id: "BinanceProvider", host: host, port: port, done: make(chan struct{}), done2: make(chan struct{}), interrupt: make(chan struct{}), messages: make(chan *BinanceMessage, 30)}
}

func (bp *BinanceProvider) Provide(c chan map[string]interface{}, ctx context.Context) {
	binanceUrl := fmt.Sprintf("%s:%d", bp.host, bp.port)

	u := url.URL{Scheme: "wss", Host: binanceUrl, Path: "/ws"}
	log.Printf("[BinanceProvider] Connecting to %s", u.String())

	connection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("[BinanceProvider] dial:", err)
	}

	bp.connection = connection

	bp.connection.SetCloseHandler(func(code int, text string) error {
		log.Printf("[BinanceProvider] CloseHandle: %d |  %s", code, text)
		bp.done <- struct{}{}
		return nil
	})

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Fatal("[BinanceProvider] ReadMessage: ", err)
				return
			}
			log.Printf("%s", message)

			var result map[string]interface{}
			json.Unmarshal(message, &result)

			c <- result
		}
	}()

	go func() {
		for {
			select {
			case <-bp.done:
				log.Println("[BinanceProvider] Connection close done")
				bp.done2 <- struct{}{}
				return
			case <-bp.interrupt:
				log.Println("[BinanceProvider] provider interrupted by abort")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("[BinanceProvider] write close", err)
					return
				}
				log.Println("[BinanceProvider] WriteMessage of closing")
				t := time.NewTicker(time.Second * 2)
				select {
				case <-bp.done:
					log.Println("[BinanceProvider] Cleanly closed the connection")
					bp.done2 <- struct{}{}
					return
				case <-t.C:
					log.Println("[BinanceProvider] Timeout")
					bp.done2 <- struct{}{}
					return

				}
			case message := <-bp.messages:
				err := bp.connection.WriteJSON(message)
				if err != nil {
					log.Fatal("[BinanceProvider] read  message:", err)
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

func (bp *BinanceProvider) Abort() {
	log.Printf("[BinanceProvider] Interrupting provider")
	close(bp.interrupt)

	log.Printf("[BinanceProvider] Interruption sent to provider.")
	<-bp.done2
}
