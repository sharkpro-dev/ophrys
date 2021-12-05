package storage

import (
	"context"
	"fmt"
	"log"
	"ophrys/pkg/engine"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/nakabonne/tstorage"
)

type TStorage struct {
	id         string
	datapath   string
	connection tstorage.Storage
	ctx        context.Context
	c          chan interface{}
}

func NewTStorage(datapath string) *TStorage {
	return &TStorage{id: "TStorage", datapath: datapath, c: make(chan interface{})}
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

func (ts *TStorage) C() chan interface{} {
	return ts.c
}

const (
	TICKER_INSERT = `INSERT INTO tickers (
		time_,
		symbol,
		price_change,
		price_change_percent,
		vwap,
		last_price,
		last_quantity,
		opening_price,
		high_price,
		low_price,
		trade_volume,
		number_of_trades 
	) VALUES (to_timestamp($1),$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
)

type PostgresStorage struct {
	id       string
	host     string
	port     int16
	user     string
	password string
	dbname   string
	ctx      context.Context
	db       *sqlx.DB
	c        chan interface{}
}

func NewPostgresStorage(host string, port int16, user string, password string, dbname string) *PostgresStorage {
	return &PostgresStorage{
		id:       "Postgres",
		host:     host,
		port:     port,
		user:     user,
		password: password,
		dbname:   dbname,
		c:        make(chan interface{}),
	}
}

func (p *PostgresStorage) Open(ctx context.Context) error {
	p.ctx = ctx

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", p.user, p.password, p.host, p.port, p.dbname))
	if err != nil {
		log.Fatalln(err)
	}

	p.db = db

	return nil
}

func (p *PostgresStorage) Store(i interface{}) {
	ticker := i.(*engine.OphrysTicker)

	p.db.MustExec(TICKER_INSERT,
		ticker.Time,
		ticker.Symbol,
		ticker.PriceChange,
		ticker.PriceChangePercent,
		ticker.Vwap,
		ticker.LastPrice,
		ticker.LastQuantity,
		ticker.OpeningPrice,
		ticker.HighPrice,
		ticker.LowPrice,
		ticker.TradeVolume,
		ticker.NumberOfTrades,
	)
}
func (p *PostgresStorage) Close() {
	p.db.Close()
}

func (p *PostgresStorage) Id() string {
	return p.id
}

func (p *PostgresStorage) C() chan interface{} {
	return p.c
}
