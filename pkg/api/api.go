package api

import (
	"net/http"
	"ophrys/pkg/engine"
)

type API interface {
}

type HttpAPI struct {
}

func (api *HttpAPI) Engage(e *engine.Engine) error {

	//_ := http.Server{Addr: "9000"}

	return nil
}

func streamSubscribe(w http.ResponseWriter, req *http.Request) {
	req.Context()
}
