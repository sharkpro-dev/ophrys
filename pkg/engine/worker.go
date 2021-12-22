package engine

import (
	"github.com/google/uuid"
)

const (
	WAITING = "WAITING"
	RUNNING = "RUNNING"
	STOPPED = "STOPPED"
)

type Worker struct {
	Uuid   uuid.UUID `json:"uuid"`
	Name   string    `json:"name"`
	engine *Engine
	task   func(w *Worker, data interface{})
	Status string `json:"status"`
	c      chan interface{}
}

func newWorker(uuid uuid.UUID, name string, engine *Engine, task func(w *Worker, data interface{}), c chan interface{}) *Worker {
	return &Worker{Uuid: uuid, Name: name, engine: engine, task: task, Status: WAITING, c: c}
}

func (w *Worker) Start() {
	go func() {
		w.engine.wg.Add(1)
		defer w.finish()

		w.Status = RUNNING
		for {
			select {
			case data := <-w.c:
				w.task(w, data)
			case <-w.engine.ctx.Done():
				return
			}
		}

	}()
}

func (w *Worker) finish() {
	w.engine.wg.Done()
	w.Status = STOPPED
}
