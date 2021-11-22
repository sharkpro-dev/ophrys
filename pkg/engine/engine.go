package engine

import (
	"context"
	"fmt"
	"sync"
)

type Storage interface {
	Open(ctx context.Context) error
	Store(i map[string]interface{})
	Close()
}

type Provider interface {
	Provide(c chan map[string]interface{}, ctx context.Context)
	Subscribe(streamId string)
	Abort()
}

type API interface {
	Engage(*Engine) error
}

type Engine struct {
	Provider   *Provider
	storage    *Storage
	marketData chan map[string]interface{}
	Done       chan struct{}
	waitGroup  sync.WaitGroup
	ctx        context.Context
	api        *API
	cancelFunc context.CancelFunc
}

func NewEngine(provider *Provider, storage *Storage, api *API) *Engine {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	return &Engine{Provider: provider, marketData: make(chan map[string]interface{}), storage: storage, ctx: ctx, cancelFunc: cancelFunc, api: api}
}

func (e *Engine) TurnOn() {
	go (*e.api).Engage(e)
	(*e.storage).Open(e.ctx)
	(*e.Provider).Provide(e.marketData, e.ctx)
	e.newWorkers(6, handleMarketData)
}

func (e *Engine) Wait() {
	e.waitGroup.Wait()
}

func (e *Engine) newWorkers(n int, f func(*Worker)) {
	for i := 0; i < n; i++ {
		e.newWorker(fmt.Sprintf("%d", i), f)
	}
}

func (e *Engine) newWorker(uuid string, f func(*Worker)) {
	w := newWorker(uuid, e, f)
	w.Start()
}

func handleMarketData(w *Worker) {
	for {
		md := <-w.engine.marketData
		go (*w.engine.storage).Store(md)
		//log.Printf(w.uuid)
	}
}

func (e *Engine) TurnOff() {
	close(e.marketData)
	e.cancelFunc()
	e.Done <- struct{}{}
}
