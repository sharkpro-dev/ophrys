package engine

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

const (
	TICKERS = "tickers"
	DEPTHS  = "depths"

	WORKER_TICKER_HANDLER = "tickerHandler"
	WORKER_DEPTH_HANDLER  = "depthHandler"
	WORKER_TICKER_STORAGE = "tickerStorage"
)

var MARKET_DATA_CATEGORIES = []string{
	TICKERS,
	DEPTHS,
}

type OphrysTicker struct {
	Time               int64
	Symbol             string
	PriceChange        float32
	PriceChangePercent float32
	Vwap               float64
	LastPrice          float64
	LastQuantity       float64
	OpeningPrice       float64
	HighPrice          float64
	LowPrice           float64
	TradeVolume        float64
	NumberOfTrades     int64
}

type OphrysDepth struct {
	symbol string
	bids   []float64
	asks   []float64
}

type Storage interface {
	Id() string
	Open(ctx context.Context) error
	C() chan interface{}
	Store(i interface{})
	Close()
}

type Provider interface {
	Id() string
	Provide(e *Engine)
	Subscribe(streamId string) chan interface{}
	Unsubscribe(streamId string) chan interface{}
	SubscriptionsList() chan interface{}
}

type API interface {
	Id() string
	Engage(*Engine) error
}

type Engine struct {
	providers    map[string]*Provider
	storage      *Storage
	apis         map[string]*API
	workers      map[uuid.UUID]*Worker
	dataChannels map[string]chan interface{}
	wg           sync.WaitGroup
	ctx          context.Context
	cancelFunc   context.CancelFunc
	depths       map[string]interface{}
}

func NewEngine(storage *Storage) *Engine {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	dataChannels := make(map[string]chan interface{})

	for _, category := range MARKET_DATA_CATEGORIES {
		dataChannels[category] = make(chan interface{})
	}

	return &Engine{
		providers:    make(map[string]*Provider),
		storage:      storage,
		apis:         make(map[string]*API),
		workers:      make(map[uuid.UUID]*Worker),
		dataChannels: dataChannels,
		wg:           sync.WaitGroup{},
		ctx:          ctx,
		cancelFunc:   cancelFunc,
	}
}

func (e *Engine) EngageAPI(api *API) {
	e.apis[(*api).Id()] = api
}

func (e *Engine) EngageProvider(provider *Provider) {
	e.providers[(*provider).Id()] = provider
}

func (e *Engine) TurnOn() {
	go func() {
		for _, provider := range e.providers {
			(*provider).Provide(e)
		}
	}()

	go func() {
		(*e.storage).Open(e.ctx)
	}()

	go func() {
		for _, api := range e.apis {
			go (*api).Engage(e)
		}
	}()

	e.newWorkers(6, WORKER_TICKER_HANDLER, handleTickers, e.tickersChannel())
	e.newWorkers(3, WORKER_DEPTH_HANDLER, handleDepths, e.depthsChannel())
	e.newWorkers(3, WORKER_TICKER_STORAGE, storeMarketData, (*e.storage).C())
}

func (e *Engine) newWorkers(n int, name string, f func(*Worker, interface{}), c chan interface{}) {
	for i := 0; i < n; i++ {
		e.newWorker(name, f, c)
	}
}

func (e *Engine) AcceptDepth(symbol string, bids []float64, asks []float64) {
	depth := &OphrysDepth{
		symbol: symbol,
		bids:   bids,
		asks:   asks,
	}

	e.depthsChannel() <- depth
}

func (e *Engine) AcceptTicker(time int64, symbol string, priceChange float32, priceChangePercent float32, vwap float64, lastPrice float64, lastQuantity float64, openingPrice float64, highPrice float64, lowPrice float64, tradeVolume float64, numberOfTrades int64) {
	ticker := &OphrysTicker{
		Time:               time,
		Symbol:             symbol,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		Vwap:               vwap,
		LastPrice:          lastPrice,
		LastQuantity:       lastQuantity,
		OpeningPrice:       openingPrice,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
		TradeVolume:        tradeVolume,
		NumberOfTrades:     numberOfTrades,
	}

	e.tickersChannel() <- ticker
}

func (e *Engine) newWorker(name string, f func(*Worker, interface{}), c chan interface{}) {
	id := uuid.New()
	w := newWorker(id, name, e, f, c)
	e.workers[id] = w
	w.Start()
}

func (e *Engine) GetProvider(id string) *Provider {
	return e.providers[id]
}

func (e *Engine) Context() context.Context {
	return e.ctx
}

func (e *Engine) tickersChannel() chan interface{} {
	return e.dataChannels[TICKERS]
}

func (e *Engine) depthsChannel() chan interface{} {
	return e.dataChannels[DEPTHS]
}

func handleTickers(w *Worker, i interface{}) {
	ophrysTicker := i.(*OphrysTicker)
	(*w.engine.storage).C() <- ophrysTicker
}

func handleDepths(w *Worker, i interface{}) {
	ophrysDepth := i.(*OphrysDepth)
	w.engine.depths[ophrysDepth.symbol] = ophrysDepth
}

func storeMarketData(w *Worker, md interface{}) {
	(*w.engine.storage).Store(md)
}

func (e *Engine) Workers() map[uuid.UUID]*Worker {
	return e.workers
}

func (e *Engine) TurnOff() {
	e.cancelFunc()
	for _, channel := range e.dataChannels {
		close(channel)
	}
	e.wg.Wait()
}
