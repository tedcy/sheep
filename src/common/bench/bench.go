package bench

import (
	"sync/atomic"
	"sync"
	"time"
	"fmt"
)

type Bench struct {
	name		string
	goroutines	[]int
	time		time.Duration
	benchFunc	func(interface{})
	initFunc	func() (interface{}, []chan<- struct{})
	accurate	bool
	data		interface{}
}

type BenchConfig struct {
	Name		string
	Goroutines	[]int
	Time		time.Duration
	BenchFunc	func(interface{})
	InitFunc	func() (interface{}, []chan<- struct{})
	Accurate	bool
}

func New(c *BenchConfig) *Bench{
	b := &Bench{}
	b.name = c.Name
	b.goroutines = c.Goroutines
	b.time = c.Time
	b.benchFunc = c.BenchFunc
	b.initFunc = c.InitFunc
	b.accurate = c.Accurate
	if b.name == "" {
		b.name = "Unknown"
	}
	if b.goroutines == nil {
		b.goroutines = []int{1,10,100,1000,5000,10000,100000}
	}
	if b.time == 0 {
		b.time = time.Second * 5
	}
	if b.benchFunc == nil {
		b.benchFunc = func(interface{}){}
	}
	if b.initFunc == nil {
		b.initFunc = func()(interface{},[]chan<-struct{}){return nil, nil}
	}
	return b
}

func (this *Bench) bench(gocount int) (uint32, time.Duration){
	var count uint32
	var sumT int64
	wg := &sync.WaitGroup{}
	after := time.Now().Add(this.time)
	for i := 0;i < gocount;i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var now time.Time
			for ;; {
				now = time.Now()
				this.benchFunc(this.data)
				atomic.AddInt64(&sumT, int64(time.Now().Sub(now)))
				atomic.AddUint32(&count, 1)
				if after.Before(now) {
					break
				}
			}
		}()
	}
	wg.Wait()
	qps := count / uint32(this.time.Seconds())
	var delay time.Duration
	if count != 0 {
		delay = time.Duration(sumT) / time.Duration(count)
	}
	//fmt.Printf("qps: %d delay: %s\n", qps, delay)
	return qps, delay
}

func (this *Bench) Run() {
	var cs []chan<-struct{}
	this.data, cs = this.initFunc()
	for _, gocount := range this.goroutines {
		if this.accurate {
			baseQps, baseDelay := this.bench(gocount)
			qps, delay := this.bench(gocount)
			fmt.Printf("name:%s c:%d qps:%d delay:%s\n", 
			this.name,
			gocount,
			int(1/(1/float64(qps) - 1/float64(baseQps))),
			delay - baseDelay)
		}else {
			qps, delay := this.bench(gocount)
			fmt.Printf("name:%s c:%d qps:%d delay:%s\n", 
			this.name,
			gocount,
			qps,
			delay)
		}
	}
	for _, c := range cs {
		close(c)
	}
}
