package utils

import "sync"

// All ConcurrentMaps should follow the Set / Get / Remove interface

// ConcurrentStringMap ...
type ConcurrentStringMap struct {
	data map[string]string
	sync.RWMutex
}

// NewConcurrentStringMap returns a new thread safe map. This should be used
// when using go routines.
func NewConcurrentStringMap() *ConcurrentStringMap {
	sm := &ConcurrentStringMap{
		data: make(map[string]string),
	}
	return sm
}

// Set sets the value ...
func (csm *ConcurrentStringMap) Set(key, value string) {
	csm.Lock()
	defer csm.Unlock()
	csm.data[key] = value
}

// Get returns the value ...
func (csm *ConcurrentStringMap) Get(key string) (string, bool) {
	csm.Lock()
	defer csm.Unlock()
	data, exists := csm.data[key]
	return data, exists
}

// ConcurrentStringIntMap ...
type ConcurrentStringIntMap struct {
	data map[string]int
	sync.RWMutex
}

// NewConcurrentStringIntMap returns a new thread safe map. This should be used
// when using go routines.
func NewConcurrentStringIntMap() *ConcurrentStringIntMap {
	sm := &ConcurrentStringIntMap{
		data: make(map[string]int),
	}
	return sm
}

// Set sets the value ...
func (csm *ConcurrentStringIntMap) Set(key string, value int) {
	csm.Lock()
	defer csm.Unlock()
	csm.data[key] = value
}

// Get returns the value ...
func (csm *ConcurrentStringIntMap) Get(key string) (int, bool) {
	csm.Lock()
	defer csm.Unlock()
	data, exists := csm.data[key]
	return data, exists
}
