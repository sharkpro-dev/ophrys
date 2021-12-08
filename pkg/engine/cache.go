package engine

import "sync"

type Cache struct {
	depthsLock  sync.RWMutex
	tickersLock sync.RWMutex
	LastDepths  map[string]*OphrysDepth
	LastTickers map[string]*OphrysTicker
}

func NewCache() *Cache {
	return &Cache{
		LastDepths:  make(map[string]*OphrysDepth),
		LastTickers: make(map[string]*OphrysTicker),
	}
}

func (c *Cache) UpdateLastDepth(lastDepth *OphrysDepth) {
	c.depthsLock.Lock()
	c.LastDepths[lastDepth.Symbol] = lastDepth
	c.depthsLock.Unlock()
}

func (c *Cache) UpdateLastTicker(lastTicker *OphrysTicker) {
	c.tickersLock.Lock()
	c.LastTickers[lastTicker.Symbol] = lastTicker
	c.tickersLock.Unlock()
}

func (c *Cache) GetLastTicker(symbol string) *OphrysTicker {
	c.tickersLock.RLock()
	lastTicker := c.LastTickers[symbol]
	c.tickersLock.RUnlock()

	return lastTicker
}

func (c *Cache) GetLastDepth(symbol string) *OphrysDepth {
	c.depthsLock.RLock()
	lastDepth := c.LastDepths[symbol]
	c.depthsLock.RUnlock()

	return lastDepth
}
