package common

import (
	"sync/atomic"
	"sync"
)

type Adder struct {
	data		*sync.Map
}

func NewAdder() *Adder {
	adder := &Adder{}
	adder.data = &sync.Map{}
	return adder
}

func (this *Adder) Add(key string, addValue uint64) {
	valuePtr, ok := this.data.Load(key)
	if !ok {
		this.data.Store(key, &addValue)
		return
	}
	valueIntPtr, ok := valuePtr.(*uint64)
	if ok {
		atomic.AddUint64(valueIntPtr, addValue)
		return
	}
	panic("invalid ptr")
	return
}

func (this *Adder) Get(key string) uint64 {
	valuePtr, ok := this.data.Load(key)
	if !ok {
		return 0
	}
	valueIntPtr, ok := valuePtr.(*uint64)
	if ok {
		return *valueIntPtr
	}
	panic("invalid ptr")
	return 0
}

func (this *Adder) List() (keys []string) {
	this.data.Range(func (key,value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return
}

func (this *Adder) Clean() {
	this.data = &sync.Map{}
}
