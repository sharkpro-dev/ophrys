package engine

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Storage interface {
	Id() string
	Open(ctx context.Context) error
	C() chan map[string]interface{}
	Store(i map[string]interface{})
	Close()
}

type Provider interface {
	Id() string
	Provide(c chan map[string]interface{}, ctx context.Context)
	Subscribe(streamId string)
	//Abort()
}

type API interface {
	Id() string
	Engage(*Engine) error
}

type Engine struct {
	providers  map[string]*Provider
	storage    *Storage
	apis       map[string]*API
	workers    map[uuid.UUID]*Worker
	marketData chan map[string]interface{}
	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewEngine(storage *Storage) *Engine {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	return &Engine{
		providers:  make(map[string]*Provider),
		storage:    storage,
		apis:       make(map[string]*API),
		workers:    make(map[uuid.UUID]*Worker),
		marketData: make(chan map[string]interface{}),
		wg:         sync.WaitGroup{},
		ctx:        ctx,
		cancelFunc: cancelFunc,
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
			(*provider).Provide(e.marketData, e.ctx)
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

	e.newWorkers(6, "marketDataHandler", handleMarketData, e.marketData)
	e.newWorkers(3, "storeMarketData", storeMarketData, (*e.storage).C())
}

func (e *Engine) newWorkers(n int, name string, f func(*Worker, map[string]interface{}), c chan map[string]interface{}) {
	for i := 0; i < n; i++ {
		e.newWorker(name, f, c)
	}
}

func (e *Engine) newWorker(name string, f func(*Worker, map[string]interface{}), c chan map[string]interface{}) {
	id := uuid.New()
	w := newWorker(id, name, e, f, c)
	e.workers[id] = w
	w.Start()
}

func (e *Engine) GetProvider(id string) *Provider {
	return e.providers[id]
}

func handleMarketData(w *Worker, md map[string]interface{}) {
	(*w.engine.storage).C() <- md
}

func storeMarketData(w *Worker, md map[string]interface{}) {
	(*w.engine.storage).Store(md)
}

func (e *Engine) Workers() map[uuid.UUID]*Worker {
	return e.workers
}

func (e *Engine) TurnOff() {
	e.cancelFunc()
	close(e.marketData)
	e.wg.Wait()
}
