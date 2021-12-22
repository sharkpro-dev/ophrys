package adt

import (
	"sync"
)

type ConcurrentCircularQueue struct {
	data   []interface{}
	lock   sync.RWMutex
	length int
}

func NewConcurrentCircularQueue(length int) *ConcurrentCircularQueue {
	return &ConcurrentCircularQueue{data: make([]interface{}, length), lock: sync.RWMutex{}, length: length}
}

func (q *ConcurrentCircularQueue) Enqueue(d interface{}) interface{} {
	var dequeued interface{}

	q.lock.Lock()
	if len(q.data) == q.length {
		dequeued, q.data = q.data[0], q.data[1:]
	}

	q.data = append(q.data, d)
	q.lock.Unlock()

	return dequeued
}

func (q *ConcurrentCircularQueue) Dequeue() interface{} {
	var dequeued interface{}
	q.lock.Lock()
	dequeued, q.data = q.data[0], q.data[1:]
	q.lock.Unlock()
	return dequeued
}

func (q *ConcurrentCircularQueue) Each(f func(interface{}, *interface{}), shared *interface{}) {
	q.lock.RLock()
	for _, d := range q.data {
		f(d, shared)
	}
	q.lock.RUnlock()
}

func (q *ConcurrentCircularQueue) Map(f func(d interface{}) interface{}) []interface{} {
	q.lock.RLock()
	var r = make([]interface{}, 0)
	for _, d := range q.data {
		if d != nil {
			r = append(r, f(d))
		}
	}
	q.lock.RUnlock()

	return r
}

type ConcurrentMap struct {
	lock sync.RWMutex
	m    map[interface{}]interface{}
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{lock: sync.RWMutex{}, m: make(map[interface{}]interface{})}
}

func (cmap *ConcurrentMap) Put(key interface{}, value interface{}) {
	cmap.lock.Lock()
	cmap.m[key] = value
	cmap.lock.Unlock()
}

func (cmap *ConcurrentMap) Get(key interface{}) (interface{}, bool) {
	cmap.lock.RLock()
	r, hasKey := cmap.m[key]
	cmap.lock.RUnlock()
	return r, hasKey
}

func (cmap *ConcurrentMap) Has(key interface{}) bool {
	_, hasKey := cmap.Get(key)
	return hasKey
}

func (cmap *ConcurrentMap) Each(f func(key interface{}, value interface{})) {
	cmap.lock.RLock()
	for key, value := range cmap.m {
		f(key, value)
	}
	cmap.lock.RUnlock()
}
