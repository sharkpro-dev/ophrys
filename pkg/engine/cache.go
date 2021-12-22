package engine

import (
	"ophrys/pkg/adt"
)

type Cache struct {
	LastDepths  *adt.ConcurrentMap
	LastTickers *adt.ConcurrentMap
}

func NewCache() *Cache {
	return &Cache{
		LastDepths:  adt.NewConcurrentMap(),
		LastTickers: adt.NewConcurrentMap(),
	}
}

func (c *Cache) UpdateLastDepth(lastDepth *OphrysDepth) {
	c.LastDepths.Put(lastDepth.Symbol, lastDepth)
}

func (c *Cache) UpdateLastTicker(lastTicker *OphrysTicker) {
	c.LastTickers.Put(lastTicker.Symbol, lastTicker)
}

func (c *Cache) GetLastTicker(symbol string) *OphrysTicker {
	ticker, ok := c.LastTickers.Get(symbol)

	if !ok {
		return nil
	}

	return ticker.(*OphrysTicker)
}

func (c *Cache) GetLastDepth(symbol string) *OphrysDepth {
	depth, ok := c.LastDepths.Get(symbol)

	if !ok {
		return nil
	}

	return depth.(*OphrysDepth)
}
