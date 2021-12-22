package main

import (
	"log"
	"ophrys/pkg/api"
	"ophrys/pkg/engine"
	"ophrys/pkg/market"
	"ophrys/pkg/storage"
	"os"
	"os/signal"

	"gonum.org/v1/gonum/stat"
)

func main() {
	var tstorage engine.Storage = storage.NewPostgresStorage("localhost", 5432, "ophrys", "ophrys", "ophrys")
	var binanceProvider engine.Provider = market.NewBinanceProvider("stream.binance.com", 9443)
	var httpApi engine.API = api.NewHttpAPI(9000)

	e := engine.NewEngine(&tstorage)
	e.EngageAPI(&httpApi)
	e.EngageProvider(&binanceProvider)

	e.AddCalculationBuckets(10, 100, 1000)

	e.AddCalculation("lastPriceMean", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.Mean(lastPrices, nil)
	})

	e.AddCalculation("lastPrice1Moment", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.Moment(1, lastPrices, nil)
	})

	e.AddCalculation("lastPriceEntropy", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.Entropy(lastPrices)
	})

	e.TurnOn()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Println("interrupted")
	e.TurnOff()

}
