package engine

import (
	"context"

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
	Abort()
}

type API interface {
	Id() string
	Engage(*Engine) error
}

type Engine struct {
	providers  map[string]*Provider
	storages   map[string]*Storage
	apis       map[string]*API
	workers    map[uuid.UUID]*Worker
	marketData chan map[string]interface{}
	Done       chan struct{}
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewEngine() *Engine {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	return &Engine{
		providers:  make(map[string]*Provider),
		storages:   make(map[string]*Storage),
		apis:       make(map[string]*API),
		workers:    make(map[uuid.UUID]*Worker),
		marketData: make(chan map[string]interface{}),
		Done:       make(chan struct{}),
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
}

func (e *Engine) EngageAPI(api *API) {
	e.apis[(*api).Id()] = api
}

func (e *Engine) EngageStorage(storage *Storage) {
	e.storages[(*storage).Id()] = storage
}

func (e *Engine) EngageProvider(provider *Provider) {
	e.providers[(*provider).Id()] = provider
}

func (e *Engine) TurnOn() {
	e.newWorker("startingProviders", func(*Worker) {
		for _, provider := range e.providers {
			(*provider).Provide(e.marketData, e.ctx)
		}
	})

	e.newWorker("startingStorages", func(*Worker) {
		for _, storage := range e.storages {
			(*storage).Open(e.ctx)
		}
	})

	e.newWorker("startingAPIs", func(*Worker) {
		for _, api := range e.apis {
			go (*api).Engage(e)
		}
	})

	e.newWorkers(6, "handleMarketData", handleMarketData)
}

func (e *Engine) newWorkers(n int, name string, f func(*Worker)) {
	for i := 0; i < n; i++ {
		e.newWorker(name, f)
	}
}

func (e *Engine) newWorker(name string, f func(*Worker)) {
	id := uuid.New()
	w := newWorker(id, name, e, f)
	e.workers[id] = w
	w.Start()
}

func (e *Engine) GetProvider(id string) *Provider {
	return e.providers[id]
}

func handleMarketData(w *Worker) {
	for {
		md := <-w.engine.marketData
		go broadcast(md, w.engine.storages)
	}
}

func broadcast(md map[string]interface{}, storages map[string]*Storage) {
	for _, storage := range storages {
		go (*storage).Store(md)
	}
}

func (e *Engine) Workers() map[uuid.UUID]*Worker {
	return e.workers
}

func (e *Engine) TurnOff() {
	close(e.marketData)
	e.cancelFunc()
	e.Done <- struct{}{}
}
