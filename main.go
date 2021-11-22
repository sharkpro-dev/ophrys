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

	var s engine.Storage = storage.NewTStorage("./data")
	p := provider.NewBinanceProvider("stream.binance.com", 9443)
	var bp engine.Provider = p

	var api engine.API = api.NewHttpAPI(9000)

	e := engine.NewEngine(&bp, &s, &api)
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
