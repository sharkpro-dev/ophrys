package engine

type Worker struct {
	uuid   string
	engine *Engine
	task   func(w *Worker)
	loop   bool
}

func newWorker(uuid string, engine *Engine, task func(e *Worker)) (w *Worker) {
	return &Worker{uuid: uuid, engine: engine, task: task}
}

func (w *Worker) Start() {
	go func() {
		w.task(w)
	}()
}
