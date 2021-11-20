package storage

import (
	"log"
	"strconv"

	"github.com/nakabonne/tstorage"
)

type Storage interface {
	Open() error
	Store(i map[string]interface{})
	Close()
}

type TStorage struct {
	datapath   string
	connection tstorage.Storage
}

func NewTStorage(datapath string) *TStorage {
	return &TStorage{datapath: datapath}
}

func (ts *TStorage) Open() error {
	storage, error := tstorage.NewStorage(
		tstorage.WithDataPath(ts.datapath),
	)

	ts.connection = storage

	if error != nil {
		return error
	}

	return nil
}

func (ts *TStorage) Store(i map[string]interface{}) {
	if i["result"] == nil {
		return
	}

	price, err := strconv.ParseFloat(i["p"].(string), 64)
	if err != nil {
		log.Fatal("price:", err)
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
