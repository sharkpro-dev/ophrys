# ophrys

Framework that provides tools to do an efficient algorithmic trading.


### How it works

Ophrys aims to help the trader to elaborate algorithms to buy and sell financial assets. It provides tools to analyze the market, and to operate.
It has an engine that can have attachments to work as needed.


### Markets

There is a developed attachment to start trading in Binance. It also can be extended to operate in different markets.


### GUI

It also counts with a Front End that is currently under development.

https://github.com/gonzaloea/ophrys-gui

### Roadmap

- [x] Connection to Binance
- [x] Exposed REST API
- [x] TimeSeries data storage adapter.
- [x] Postgresql price data storage adapter. 
- [x] Asset price analysis tools
- [ ] Notifications
- [ ] Some algorithms

### How to contribute?

Fork the repo. Make a PR and wait for the review!

### Initial SQL script to use with postgres attachment.

```sql
CREATE TABLE public.tickers
(
    time_ timestamp with time zone,
    symbol character varying COLLATE pg_catalog."default",
    price_change numeric,
    price_change_percent numeric,
    vwap numeric,
    last_price numeric,
    last_quantity numeric,
    opening_price numeric,
    high_price numeric,
    low_price numeric,
    trade_volume numeric,
    number_of_trades numeric
)
```
