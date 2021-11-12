package engine

import (
	"log"
	pr "ophrys/pkg/provider"
	st "ophrys/pkg/storage"
)

type Engine struct {
	provider   *pr.Provider
	storage    *st.Storage
	marketData chan map[string]interface{}

	Done chan struct{}
}

func NewEngine(provider *pr.Provider, storage *st.Storage) *Engine {
	return &Engine{provider: provider, marketData: make(chan map[string]interface{}), storage: storage}
}

func (e *Engine) TurnOn() {
	(*e.storage).Open()
	(*e.provider).Provide(e.marketData)

	for {
		md := <-e.marketData
		(*e.storage).Store(md)
		log.Printf("%s", md)

	}
}

func (e *Engine) TurnOff() {
	(*e.provider).Abort()
	(*e.storage).Close()
	e.Done <- struct{}{}
}
