package main

import (
	"log"
	"ophrys/pkg/api"
	"ophrys/pkg/engine"
	"ophrys/pkg/provider"
	"ophrys/pkg/storage"
	"os"
	"os/signal"
)

func main() {
	var tstorage engine.Storage = storage.NewTStorage("./data")
	var binanceProvider engine.Provider = provider.NewBinanceProvider("stream.binance.com", 9443)
	var httpApi engine.API = api.NewHttpAPI(9000)

	e := engine.NewEngine()
	e.EngageAPI(&httpApi)
	e.EngageProvider(&binanceProvider)
	e.EngageStorage(&tstorage)

	e.TurnOn()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		log.Println("interrupted")
		e.TurnOff()
		select {
		case <-e.Done:
			log.Println("interrupttt -> Done")
			break
		}

		return
	}

}
