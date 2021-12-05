package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ophrys/pkg/engine"
	"strings"

	"github.com/gorilla/mux"
)

const (
	ASSET_KEY       = "asset"
	PROVIDER_ID_KEY = "providerId"

	HEADER_CONTENT_TYPE                 = "Content-Type"
	HEADER_ACCESS_CONTROL_ALLOW_ORIGIN  = "Access-Control-Allow-Origin"
	HEADER_ACCESS_CONTROL_ALLOW_HEADERS = "Access-Control-Allow-Headers"

	MIME_APPLICATION_JSON = "application/json"

	ROUTE_ASSET_SUBSCRIBE    = "/asset/subscribe"
	ROUTE_ASSET_UNSUBSCRIBE  = "/asset/unsubscribe"
	ROUTE_ASSET_DEPTH        = "/asset/depth"
	ROUTE_SUBSCRIPTIONS_LIST = "/subscriptions"
	ROUTE_WORKERS            = "/workers"
)

type OphrysEngineHandler struct {
	e *engine.Engine
	f func(e *engine.Engine, w http.ResponseWriter, r *http.Request) error
}

func (oeh *OphrysEngineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(HEADER_ACCESS_CONTROL_ALLOW_ORIGIN, "*")
	w.Header().Set(HEADER_ACCESS_CONTROL_ALLOW_HEADERS, HEADER_CONTENT_TYPE)
	err := oeh.f(oeh.e, w, r)
	if err != nil {
		log.Printf(err.Error())
	}
}

type HttpAPI struct {
	port int
	id   string
}

func (api *HttpAPI) Id() string {
	return api.id
}

func NewHttpAPI(port int) *HttpAPI {
	return &HttpAPI{port: port}
}

func (api *HttpAPI) Engage(e *engine.Engine) error {
	r := mux.NewRouter()

	r.Handle(ROUTE_ASSET_SUBSCRIBE, &OphrysEngineHandler{e: e, f: subscribeAsset}).Methods(http.MethodPost, http.MethodOptions)
	r.Handle(ROUTE_ASSET_UNSUBSCRIBE, &OphrysEngineHandler{e: e, f: unsubscribeAsset}).Methods(http.MethodPost, http.MethodOptions)
	r.Handle(ROUTE_ASSET_DEPTH, &OphrysEngineHandler{e: e, f: assetDepth}).Methods(http.MethodGet, http.MethodOptions)
	r.Handle(ROUTE_SUBSCRIPTIONS_LIST, &OphrysEngineHandler{e: e, f: subscriptionsList}).Methods(http.MethodGet, http.MethodOptions)
	r.Handle(ROUTE_WORKERS, &OphrysEngineHandler{e: e, f: workersList}).Methods(http.MethodGet, http.MethodOptions)

	return http.ListenAndServe(fmt.Sprintf(":%d", api.port), r)
}

func subscribeAsset(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodPost {
		var p map[string]interface{}

		json.NewDecoder(r.Body).Decode(&p)

		var asset string = p[ASSET_KEY].(string)
		var providerId string = p[PROVIDER_ID_KEY].(string)

		response := (*e.GetProvider(providerId)).Subscribe(asset)

		responseJSON, err := json.Marshal(response)
		if err != nil {
			return err
		}
		w.Header().Set(HEADER_CONTENT_TYPE, MIME_APPLICATION_JSON)
		_, err = w.Write(responseJSON)

		if err != nil {
			return err
		}
	}

	return nil

}

func unsubscribeAsset(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodPost {
		var p map[string]interface{}

		json.NewDecoder(r.Body).Decode(&p)

		var asset string = p[ASSET_KEY].(string)
		var providerId string = p[PROVIDER_ID_KEY].(string)

		response := (*e.GetProvider(providerId)).Unsubscribe(asset)

		responseJSON, err := json.Marshal(response)
		if err != nil {
			return err
		}
		w.Header().Set(HEADER_CONTENT_TYPE, MIME_APPLICATION_JSON)
		_, err = w.Write(responseJSON)

		if err != nil {
			return err
		}
	}

	return nil

}

func subscriptionsList(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		providerId := r.URL.Query().Get(PROVIDER_ID_KEY)

		subscriptionList := <-(*e.GetProvider(providerId)).SubscriptionsList()

		subscriptionListJSON, err := json.Marshal(subscriptionList)
		if err != nil {
			return err
		}
		w.Header().Set(HEADER_CONTENT_TYPE, MIME_APPLICATION_JSON)
		_, err = w.Write(subscriptionListJSON)

		if err != nil {
			return err
		}
	}

	return nil
}

func assetDepth(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		asset := r.URL.Query().Get(ASSET_KEY)

		depth := e.Depths[strings.ToUpper(asset)]

		depthJSON, err := json.Marshal(depth)
		if err != nil {
			return err
		}
		w.Header().Set(HEADER_CONTENT_TYPE, MIME_APPLICATION_JSON)
		_, err = w.Write(depthJSON)

		if err != nil {
			return err
		}
	}

	return nil
}

func workersList(e *engine.Engine, w http.ResponseWriter, r *http.Request) error {
	workersJSON, err := json.Marshal(e.Workers())
	if err != nil {
		return err
	}
	w.Header().Set(HEADER_CONTENT_TYPE, MIME_APPLICATION_JSON)
	_, err = w.Write(workersJSON)

	if err != nil {
		return err
	}

	return nil
}
