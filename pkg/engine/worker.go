package engine

import "github.com/google/uuid"

type Worker struct {
	uuid   uuid.UUID
	name   string
	engine *Engine
	task   func(w *Worker)
	loop   bool
}

func newWorker(uuid uuid.UUID, name string, engine *Engine, task func(e *Worker)) (w *Worker) {
	return &Worker{uuid: uuid, name: name, engine: engine, task: task}
}

func (w *Worker) Start() {
	go func() {
		w.task(w)
	}()
}
