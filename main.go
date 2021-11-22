package main

import (
	"flag"
	"log"
	"ophrys/pkg/api"
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

	var tstorage engine.Storage = storage.NewTStorage("./data")
	var binanceProvider engine.Provider = provider.NewBinanceProvider("stream.binance.com", 9443)
	var httpApi engine.API = api.NewHttpAPI(9000)

	e := engine.NewEngine()
	e.EngageAPI(&httpApi)
	e.EngageProvider(&binanceProvider)
	e.EngageStorage(&tstorage)
	//p.Subscribe("btcusdt@aggTrade")
	//p.Subscribe("adausdt@aggTrade")

	e.TurnOn()

	e.Wait()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		log.Println("interrupttt")
		e.TurnOff()
		return
		// select {
		// case <-e.Done:
		// 	log.Println("interrupttt -> Done")
		// 	break
		// }

	}

}
