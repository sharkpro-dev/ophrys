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
	p := provider.NewBinanceProvider("stream.binance.com", 9443)
	var bp provider.Provider = p

	e := engine.NewEngine(&bp, &s)
	//p.EnqueueSubscription("btcusdt@aggTrade")
	p.EnqueueSubscription("adausdt@aggTrade")

	e.TurnOn()

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
