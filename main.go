package main

import (
	"flag"
	"log"
	"ophrys/pkg/engine"
	"ophrys/pkg/provider"
	"ophrys/pkg/storage"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "stream.binance.com:9443", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	var s storage.Storage = storage.NewTStorage("./data")
	var p provider.Provider = provider.NewBinanceProvider("stream.binance.com", 9443)
	e := engine.NewEngine(&p, &s)

	e.TurnOn()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {

		select {
		case <-e.Done:
			log.Println("Doneee")
			return
		case <-interrupt:
			log.Println("interrupttt")
			e.TurnOff()
			select {
			case <-e.Done:
				return
			}

		}

	}

	// storage, _ := tstorage.NewStorage(
	// 	tstorage.WithDataPath("./data"),
	// )
	// defer storage.Close()

	// webhook, err := disgohook.NewWebhookClientByToken(nil, nil, "906972566755885127/v2KF4TUjFtgNdIkX_qXJYkttyaDM6FFAAPnlVHv7d9qUghemzfarcVbmQ1QbEPqdSgGq")

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	// u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws/adausdt@aggTrade"}
	// log.Printf("connecting to %s", u.String())

	// c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// if err != nil {
	// 	log.Fatal("dial:", err)
	// }
	// defer c.Close()

	// ticker := time.NewTicker(10 * time.Minute)
	// quit := make(chan struct{})
	// go func() {
	// 	defer close(quit)
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			points, err := storage.Select("aggTrade", []tstorage.Label{
	// 				{Name: "ADAUSDT"},
	// 			}, time.Now().UnixMilli()-3000, time.Now().UnixMilli()+5000)

	// 			if err == nil && len(points) > 0 {
	// 				log.Println("Emitting to discord: ADAUSDT: ", points[len(points)-1].Value)
	// 				_, err = webhook.SendContent(fmt.Sprintf("ADAUSDT: %f", points[len(points)-1].Value))
	// 			}

	// 		case <-quit:
	// 			ticker.Stop()
	// 			return
	// 		}
	// 	}
	// }()

	// done := make(chan struct{})

	// go func() {
	// 	defer close(done)
	// 	for {
	// 		_, message, err := c.ReadMessage()
	// 		if err != nil {
	// 			log.Fatal("read:", err)
	// 			return
	// 		}
	// 		//log.Printf("recv: %s", message)

	// 		var result map[string]interface{}
	// 		json.Unmarshal(message, &result)
	// 		log.Printf("%s", message)

	// 		price, err := strconv.ParseFloat(result["p"].(string), 64)
	// 		if err != nil {
	// 			log.Fatal("price:", err)
	// 			return
	// 		}

	// 		_ = storage.InsertRows([]tstorage.Row{
	// 			{
	// 				Metric: result["e"].(string),
	// 				Labels: []tstorage.Label{
	// 					{Name: result["s"].(string)},
	// 				},
	// 				DataPoint: tstorage.DataPoint{Timestamp: int64(result["T"].(float64)), Value: price},
	// 			},
	// 		})
	// 	}
	// }()

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	case <-interrupt:
	// 		log.Println("interrupt")

	// 		// Cleanly close the connection by sending a close message and then
	// 		// waiting (with timeout) for the server to close the connection.
	// 		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		if err != nil {
	// 			log.Println("write close:", err)
	// 			return
	// 		}
	// 		select {
	// 		case <-done:
	// 			return
	// 		}
	// 	}
	// }

}
