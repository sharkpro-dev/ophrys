package storage

import (
	"context"
	"log"
	"strconv"

	"github.com/nakabonne/tstorage"
)

type TStorage struct {
	id         string
	datapath   string
	connection tstorage.Storage
	ctx        context.Context
	c          chan map[string]interface{}
}

func NewTStorage(datapath string) *TStorage {
	return &TStorage{id: "TStorage", datapath: datapath, c: make(chan map[string]interface{})}
}

func (ts *TStorage) Open(ctx context.Context) error {
	ts.ctx = ctx
	storage, err := tstorage.NewStorage(
		tstorage.WithDataPath(ts.datapath),
	)

	if err != nil {
		return err
	}

	ts.connection = storage

	return nil
}

func (ts *TStorage) Id() string {
	return ts.id
}

func (ts *TStorage) Store(i map[string]interface{}) {
	if i["result"] == nil {
		return
	}

	price, err := strconv.ParseFloat(i["p"].(string), 64)
	if err != nil {
		log.Fatal("FATAL: price:", err)
		return
	}

	_ = ts.connection.InsertRows([]tstorage.Row{
		{
			Metric: i["e"].(string),
			Labels: []tstorage.Label{
				{Name: i["s"].(string)},
			},
			DataPoint: tstorage.DataPoint{Timestamp: int64(i["T"].(float64)), Value: price},
		},
	})
}
func (ts *TStorage) Close() {
	ts.connection.Close()
}

func (ts *TStorage) C() chan map[string]interface{} {
	return ts.c
}
