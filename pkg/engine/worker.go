package engine

import "github.com/google/uuid"

const (
	WAITING = "WAITING"
	RUNNING = "RUNNING"
	STOPPED = "STOPPED"
)

type Worker struct {
	Uuid   uuid.UUID `json:"uuid"`
	Name   string    `json:"name"`
	engine *Engine
	task   func(w *Worker)
	Status string `json:"status"`
}

func newWorker(uuid uuid.UUID, name string, engine *Engine, task func(e *Worker)) (w *Worker) {
	return &Worker{Uuid: uuid, Name: name, engine: engine, task: task, Status: WAITING}
}

func (w *Worker) Start() {
	go func() {
		w.engine.wg.Add(1)
		w.Status = RUNNING
		w.task(w)
		w.Status = STOPPED
		w.engine.wg.Done()
	}()
}
