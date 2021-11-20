package engine

import (
	"context"
	"fmt"
	"log"
	pr "ophrys/pkg/provider"
	st "ophrys/pkg/storage"
	"sync"
)

type Engine struct {
	provider   *pr.Provider
	storage    *st.Storage
	marketData chan map[string]interface{}
	Done       chan struct{}
	waitGroup  sync.WaitGroup
	Ctx        context.Context
}

func NewEngine(provider *pr.Provider, storage *st.Storage) *Engine {
	return &Engine{provider: provider, marketData: make(chan map[string]interface{}), storage: storage, Ctx: context.Background()}
}

func (e *Engine) TurnOn() {
	(*e.storage).Open()
	(*e.provider).Provide(e.marketData)

	for i := 1; i < 4; i++ {
		go func(n int) {
			e.waitGroup.Add(1)
			defer e.waitGroup.Done()
			for {
				md := <-e.marketData
				go (*e.storage).Store(md)
				log.Printf(fmt.Sprintf("%d", n))
			}
		}(i)
	}
}

func (e *Engine) TurnOff() {
	log.Printf("Attemting to : TurnOff Done")
	(*e.provider).Abort()
	(*e.storage).Close()
	log.Printf("TurnOff Done")
	e.Done <- struct{}{}
}
