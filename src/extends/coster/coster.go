package coster

import (
	"github.com/tedcy/sheep/src/common"
	//"github.com/tedcy/sheep/src/extends/log"
	"sync"
	"time"
	"sync/atomic"
)

//TODO this can be a base logic instead of extends
type CosterManager struct {
	costers		map[interface{}]*Coster
	lock		sync.Mutex
}

var g_costerManager *CosterManager;

func GetInstance() *CosterManager{
	if g_costerManager == nil {
		g_costerManager = new(CosterManager)
		g_costerManager.costers = make(map[interface{}]*Coster)
	}
	return g_costerManager;
}

func (this *CosterManager) GetCoster(name interface{}) *Coster{
	this.lock.Lock()
	defer this.lock.Unlock()
	coster, ok := this.costers[name]
	if (!ok) {
		coster = new(Coster)
		coster.Init()
		this.costers[name] = coster
	}
	return coster
}

func (this *CosterManager) Start() *CosterOnce{
	return &CosterOnce{
		start:		time.Now(),
	}
}

type Coster struct{
	common.SimpleQueue
	t		int64
	count	int64
	closed	chan struct{}
}

func (this *Coster) Init() {
	this.closed = make(chan struct{})
	go this.costerStatisticsLooper()
}

func (this *Coster) costerStatisticsLooper() {
	for {
		select {
		case <-time.After(time.Second):
			t := atomic.SwapInt64(&this.t, 0)
			count := atomic.SwapInt64(&this.count, 0)
			if count != 0 {
				this.SimpleQueue.Insert(t / count)
			}
		case <-this.closed:
			return
		}
	}
}

type CosterOnce struct {
	start time.Time
}

func (this *Coster) End(once *CosterOnce) {
	delta := time.Now().Sub(once.start)
	atomic.AddInt64(&this.t, int64(delta))
	atomic.AddInt64(&this.count, 1)
}
