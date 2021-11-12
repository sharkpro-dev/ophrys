package provider

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type Provider interface {
	Provide(c chan map[string]interface{})
	Abort()
}

type BinanceProvider struct {
	host       string
	port       int32
	connection *websocket.Conn
	done       chan struct{}
	interrupt  chan struct{}
}

func NewBinanceProvider(host string, port int32) *BinanceProvider {
	return &BinanceProvider{host: host, port: port, done: make(chan struct{}), interrupt: make(chan struct{})}
}

func (bp *BinanceProvider) Provide(c chan map[string]interface{}) {
	binanceUrl := fmt.Sprintf("%s:%d", bp.host, bp.port)

	u := url.URL{Scheme: "wss", Host: binanceUrl, Path: "/ws/adausdt@aggTrade"}
	log.Printf("connecting to %s", u.String())

	connection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	go func() {
		for {
			defer close(bp.done)
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Fatal("read:", err)
				return
			}
			log.Printf("recv: %s", message)

			var result map[string]interface{}
			json.Unmarshal(message, &result)

			c <- result

			// price, err := strconv.ParseFloat(result["p"].(string), 64)
			// if err != nil {
			// 	log.Fatal("price:", err)
			// 	return
			// }

			// _ = storage.InsertRows([]tstorage.Row{
			// 	{
			// 		Metric: result["e"].(string),
			// 		Labels: []tstorage.Label{
			// 			{Name: result["s"].(string)},
			// 		},
			// 		DataPoint: tstorage.DataPoint{Timestamp: int64(result["T"].(float64)), Value: price},
			// 	},
			// })
		}
	}()

	go func() {
		for {
			select {
			case <-bp.done:
				return
			case <-bp.interrupt:
				log.Println("interrupted by abort")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-bp.done:
					log.Println("Cleanly closed the connection")
					return
				}
			}
		}
	}()
}

func (bp *BinanceProvider) Abort() {
	close(bp.interrupt)
	bp.connection.Close()
}
