# ophrys

Framework that provides tools to do an efficient algorithmic trading.


### How it works

Ophrys aims to help the trader to elaborate algorithms to buy and sell financial assets. It provides tools to analyze the market, and to operate.
It has an engine that can have attachments to work as needed.


### Markets

There is a developed attachment to start trading in Binance. It also can be extended to operate in different markets.


### GUI

It also counts with a Front End that is currently under development.

https://github.com/sharkpro-dev/ophrys-gui

### Roadmap

- [x] Connection to Binance
- [x] Exposed REST API
- [x] TimeSeries data storage adapter.
- [x] Postgresql price data storage adapter. 
- [x] Asset price analysis tools
- [ ] Notifications
- [ ] Some algorithms

### How to contribute?

Fork the repo. Make a PR and wait for the review!

### Example

```go
package main

import (
	"flag"
	"fmt"
	"ophrys/pkg/api"
	"ophrys/pkg/engine"
	"ophrys/pkg/market"
	"ophrys/pkg/storage"
	"os"
	"os/signal"

	"gonum.org/v1/gonum/stat"
)

func main() {
	secretKeyPtr := flag.String("secretKey", "", "Market Client Secret Key.")
	apiKeyPtr := flag.String("apiKey", "", "Market Client API Key")
	flag.Parse()
	fmt.Printf("secretKeyPtr: %s, apiKeyPtr: %s\n", *secretKeyPtr, *apiKeyPtr)

	var binanceClient engine.MarketClient = market.NewBinanceClient("https://api.binance.com", *apiKeyPtr, *secretKeyPtr)
	var tstorage engine.Storage = storage.NewPostgresStorage("localhost", 5432, "ophrys", "ophrys", "ophrys")
	var binanceProvider engine.Provider = market.NewBinanceProvider("stream.binance.com", 9443)
	var httpApi engine.API = api.NewHttpAPI(9000)

	e := engine.NewEngine(&tstorage)
	e.EngageAPI(&httpApi)
	e.EngageProvider(&binanceProvider)
	e.EngageMarketClient(&binanceClient)

	e.AddCalculationBuckets(10, 100, 1000)

	e.AddCalculation("lastPriceMean", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.Mean(lastPrices, nil)
	})

	e.AddCalculation("lastPriceStdDev", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.StdDev(lastPrices, nil)
	})

	e.TurnOn()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	e.TurnOff()

}
```

### Initial SQL script to use with postgres attachment.

```sql
CREATE TABLE public.tickers
(
    time_ timestamp with time zone,
    symbol character varying COLLATE pg_catalog."default",
    price_change numeric,
    price_change_percent numeric,
    vwap numeric,
    last_price numeric,
    last_quantity numeric,
    opening_price numeric,
    high_price numeric,
    low_price numeric,
    trade_volume numeric,
    number_of_trades numeric
)
```
