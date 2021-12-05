package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"ophrys/pkg/engine"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	LIST_SUBSCRIPTIONS = "LIST_SUBSCRIPTIONS"
	SUBSCRIBE          = "SUBSCRIBE"
	UNSUBSCRIBE        = "UNSUBSCRIBE"

	MESSAGE_DEPTH_UPDATE = "depthUpdate"
	MESSAGE_24HR_TICKER  = "24hrTicker"
)

type BinanceProvider struct {
	id         string
	host       string
	port       int32
	connection *websocket.Conn
	messages   chan *BinanceOutgoingMessage
	ctx        context.Context
	done       chan struct{}
	responses  map[int]chan interface{}
}

type BinanceOutgoingMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

func NewBinanceProvider(host string, port int32) *BinanceProvider {
	return &BinanceProvider{id: "BinanceProvider", host: host, port: port, done: make(chan struct{}), messages: make(chan *BinanceOutgoingMessage, 30),
		responses: make(map[int]chan interface{})}
}

func (bp *BinanceProvider) Provide(e *engine.Engine) {
	bp.ctx = e.Context()

	binanceUrl := fmt.Sprintf("%s:%d", bp.host, bp.port)

	u := url.URL{Scheme: "wss", Host: binanceUrl, Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	connection, _, err := websocket.DefaultDialer.DialContext(e.Context(), u.String(), nil)
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
			type_, message, err := connection.ReadMessage()
			if err != nil {
				log.Printf("ReadMessage: %s", err.Error())
				return
			}

			log.Printf("%d, %s", type_, message)

			var result map[string]interface{}
			json.Unmarshal(message, &result)

			var responseChannel chan interface{}
			id, ok := result["id"]

			if ok {
				responseChannel, ok = bp.responses[int(id.(float64))]
			}

			if ok {
				responseChannel <- result
				delete(bp.responses, int(id.(float64)))
			} else {
				switch result["e"] {
				case MESSAGE_DEPTH_UPDATE:
					e.AcceptDepth(
						result["s"].(string),
						interfaceSliceToStringSlice(result["b"].([]interface{})),
						interfaceSliceToStringSlice(result["a"].([]interface{})),
					)
				case MESSAGE_24HR_TICKER:

					e.AcceptTicker(
						int64(result["E"].(float64))/1000,
						result["s"].(string),
						float32(stringToFloat(result["p"].(string))),
						float32(stringToFloat(result["P"].(string))),
						stringToFloat(result["w"].(string)),
						stringToFloat(result["c"].(string)),
						stringToFloat(result["Q"].(string)),
						stringToFloat(result["o"].(string)),
						stringToFloat(result["h"].(string)),
						stringToFloat(result["l"].(string)),
						stringToFloat(result["v"].(string)),
						int64(result["n"].(float64)),
					)
				}
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

func (bp *BinanceProvider) subscribe(path string) chan interface{} {
	id := rand.Intn(1000)
	bp.responses[id] = make(chan interface{})

	bp.messages <- &BinanceOutgoingMessage{Method: SUBSCRIBE, Params: []string{path}, Id: id}

	return bp.responses[id]
}

func (bp *BinanceProvider) Subscribe(asset string) interface{} {
	response := make([]interface{}, 2)

	response[0] = <-bp.subscribe(fmt.Sprintf("%s@ticker", asset))
	response[1] = <-bp.subscribe(fmt.Sprintf("%s@depth", asset))

	return response
}

func (bp *BinanceProvider) unsubscribe(path string) chan interface{} {
	id := rand.Intn(1000)
	bp.responses[id] = make(chan interface{})
	bp.messages <- &BinanceOutgoingMessage{Method: UNSUBSCRIBE, Params: []string{path}, Id: id}

	return bp.responses[id]
}

func (bp *BinanceProvider) Unsubscribe(asset string) interface{} {
	response := make([]interface{}, 2)

	response[0] = <-bp.unsubscribe(fmt.Sprintf("%s@ticker", asset))
	response[1] = <-bp.unsubscribe(fmt.Sprintf("%s@depth", asset))

	return response
}

func (bp *BinanceProvider) SubscriptionsList() chan interface{} {
	id := rand.Intn(1000)
	bp.responses[id] = make(chan interface{})
	bp.messages <- &BinanceOutgoingMessage{Method: LIST_SUBSCRIPTIONS, Id: id}

	return bp.responses[id]
}

func stringToFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 16)

	if err != nil {
		log.Fatal(err)
	}

	return value
}

func interfaceSliceToStringSlice(slice []interface{}) [][]string {
	var floatSlice [][]string = make([][]string, len(slice))

	for i, d := range slice {
		floatSlice[i] = make([]string, len(d.([]interface{})))
		for j, e := range d.([]interface{}) {
			floatSlice[i][j] = e.(string)
		}
	}

	return floatSlice
}
