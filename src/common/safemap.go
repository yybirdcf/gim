package common

import (
	"sync"
)

type SafeMap struct {
	lock *sync.RWMutex
	sm   map[interface{}]interface{}
}

func NewSafeMap(sm map[interface{}]interface{}) *SafeMap {
	return &SafeMap{
		lock: new(sync.RWMutex),
		sm:   sm,
	}
}

func (self *SafeMap) Get(k interface{}) interface{} {
	self.lock.Lock()
	defer self.lock.Unlock()

	if val, ok := self.sm[k]; ok {
		return val
	}

	return nil
}

func (self *SafeMap) Set(k interface{}, v interface{}) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	if val, ok := self.sm[k]; !ok {
		self.sm[k] = v
	} else if val != v {
		self.sm[k] = v
	} else {
		return false
	}

	return true
}

func (self *SafeMap) Check(k interface{}) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.sm[k]; !ok {
		return false
	}

	return true
}

func (self *SafeMap) Delete(k interface{}) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.sm, k)
}
