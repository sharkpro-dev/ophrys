package engine

import (
	"log"
	"ophrys/pkg/adt"
)

type StatisticsCalculationManager struct {
	c                     chan interface{}
	statisticsCalculators *adt.ConcurrentMap
	calculus              map[string]func([]interface{}) float64
	bucketSizes           []int
}

func NewStatisticsCalculationManager() *StatisticsCalculationManager {
	return &StatisticsCalculationManager{
		c:                     make(chan interface{}),
		statisticsCalculators: adt.NewConcurrentMap(),
		calculus:              make(map[string]func([]interface{}) float64),
		bucketSizes:           make([]int, 0),
	}
}

func (scm *StatisticsCalculationManager) C() chan interface{} {
	return scm.c
}

func (scm *StatisticsCalculationManager) addCalculus(name string, c func([]interface{}) float64) {
	scm.calculus[name] = c
}

func (scm *StatisticsCalculationManager) addBucketSize(n int) {
	scm.bucketSizes = append(scm.bucketSizes, n)
	scm.statisticsCalculators.Each(func(key interface{}, value interface{}) {
		_, exists := value.(*StatisticsCalculator).tickerBuckets[n]
		if !exists {
			value.(*StatisticsCalculator).tickerBuckets[n] = adt.NewConcurrentCircularQueue(n)
		}
	})
}

func handleTicker(w *Worker, t interface{}) {
	ophrysTicker := t.(*OphrysTicker)
	calculator, ok := w.engine.stats.statisticsCalculators.Get(ophrysTicker.Symbol)

	if !ok {

		calculator = NewStatisticsCalculator(w.engine.stats.bucketSizes)
		w.engine.stats.statisticsCalculators.Put(ophrysTicker.Symbol, calculator)
	}

	calculator.(*StatisticsCalculator).Accept(ophrysTicker, w.engine.stats.calculus)
}

type StatisticsCalculator struct {
	tickerBuckets map[int]*adt.ConcurrentCircularQueue
	calculations  map[int]map[string]float64
}

func NewStatisticsCalculator(bucketSizes []int) *StatisticsCalculator {
	tickerBuckets := make(map[int]*adt.ConcurrentCircularQueue)

	for _, n := range bucketSizes {
		tickerBuckets[n] = adt.NewConcurrentCircularQueue(n)
	}

	return &StatisticsCalculator{
		tickerBuckets: tickerBuckets,
		calculations:  make(map[int]map[string]float64),
	}
}

func (sc *StatisticsCalculator) Accept(ticker *OphrysTicker, calcs map[string]func([]interface{}) float64) {
	for bucket, queue := range sc.tickerBuckets {
		queue.Enqueue(ticker)

		prices := queue.Map(func(o interface{}) interface{} {
			return o.(*OphrysTicker)
		})

		for name, calc := range calcs {
			_, ok := sc.calculations[bucket]
			if !ok {
				sc.calculations[bucket] = make(map[string]float64)
			}

			sc.calculations[bucket][name] = calc(prices)
		}
	}
	log.Print(sc.calculations)
}
