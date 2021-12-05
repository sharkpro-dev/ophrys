# ophrys



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