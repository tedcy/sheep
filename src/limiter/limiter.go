package limiter

//限流器和熔断器的区别
//限流器是对单个请求来做限流的，没有状态转换，越在上层拦截越好
//熔断器可以不在上层，因为只需要拥有统计功能，对外汇报状态就可以让其他层进行熔断，是可以在后层的
//所以限流器更适合服务端，而熔断器更适合客户端

import (
	"coding.net/tedcy/sheep/src/common"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"sync/atomic"
	"time"
)

var ErrDefaultFallBack error = errors.New("default fallback")
var ErrInvokerIsNil error = errors.New("invoker is nil")

type LimiterI interface {
	Execute(invoker, fallback func() (interface{}, error)) (interface{}, error)
	Close() error
}

type LimiterType int

const (
	NoneLimiterType LimiterType = iota
	QueueLengthLimiterType
	InvokeTimeLimiterType
)

func New(ctx context.Context, t LimiterType, limit int64) LimiterI {
	switch t {
	case QueueLengthLimiterType:
		return NewQueueLengthLimiter(ctx, int32(limit))
	case InvokeTimeLimiterType:
		return NewInvokeTimeLimter(ctx, limit)
	}
	return nil
}

func NewQueueLengthLimiter(ctx context.Context, limit int32) *QueueLengthLimiter {
	l := &QueueLengthLimiter{}
	l.limit = limit
	l.ctx, l.cancel = context.WithCancel(ctx)
	go l.timeLooper()
	return l
}

type QueueLengthLimiter struct {
	queueLength int32
	limit       int32
	ctx         context.Context
	cancel      context.CancelFunc
}

func (this *QueueLengthLimiter) Execute(invoker, fallback func() (interface{}, error)) (resp interface{}, err error) {
	defer atomic.AddInt32(&this.queueLength, -1)
	if atomic.AddInt32(&this.queueLength, 1) > this.limit {
		if fallback != nil {
			resp, err = fallback()
		} else {
			err = ErrDefaultFallBack
		}
	} else {
		if invoker != nil {
			resp, err = invoker()
		} else {
			err = ErrInvokerIsNil
		}
	}

	return
}

func (this *QueueLengthLimiter) timeLooper() {
	for {
		select {
		case <-time.After(time.Second):
		case <-this.ctx.Done():
			return
		}
		println(atomic.LoadInt32(&this.queueLength))
	}
}

func (this *QueueLengthLimiter) Close() error {
	if this.cancel != nil {
		this.cancel()
	}
	return nil
}

//自动限流原理
//响应时间和任务队列成正比
//如果目前响应时间大于限定响应时间，那么根据比例推断限制多少任务队列可能达到预期的响应时间
//只计算了invoker而没有计算fallback的响应时间
//计算响应时间和任务队列做了一些去噪处理，取可信的一些数据
//太过激烈的波动限额值会引起更大的拥塞问题，所以每次步进delta值的1/2
func NewInvokeTimeLimter(ctx context.Context, limit int64) *InvokeTimeLimiter {
	l := &InvokeTimeLimiter{}
	l.limit = limit * int64(time.Millisecond)
	l.ctx, l.cancel = context.WithCancel(ctx)
	l.tQueue = common.NewSimpleQueue()
	l.lQueue = common.NewSimpleQueue()
	go l.timeLooper()
	return l
}

type InvokeTimeLimiter struct {
	tQueue      *common.SimpleQueue
	lQueue      *common.SimpleQueue
	limit       int64
	lengthLimit int64
	t           int64
	count       int64
	queueLength int64
	ctx         context.Context
	cancel      context.CancelFunc
}

func (this *InvokeTimeLimiter) Execute(invoker, fallback func() (interface{}, error)) (resp interface{}, err error) {
	defer atomic.AddInt64(&this.queueLength, -1)
	lengthLimit := atomic.LoadInt64(&this.lengthLimit)
	nowLength := atomic.AddInt64(&this.queueLength, 1)
	if lengthLimit != 0 && nowLength > lengthLimit {
		if fallback != nil {
			resp, err = fallback()
		} else {
			err = ErrDefaultFallBack
		}
	} else {
		if invoker != nil {
			t := time.Now()
			resp, err = invoker()
			delta := time.Now().Sub(t)
			atomic.AddInt64(&this.t, int64(delta))
			atomic.AddInt64(&this.count, 1)
		} else {
			err = ErrInvokerIsNil
		}
	}

	return
}

func (this *InvokeTimeLimiter) Close() error {
	if this.cancel != nil {
		this.cancel()
	}
	return nil
}

func (this *InvokeTimeLimiter) timeLooper() {
	/*go func() {
		var delta time.Duration = time.Second * 2
		for {
			select {
			case <-time.After(delta):
			case <-this.ctx.Done():
				return
			}
			t := time.Now().Add(-delta)
			avgCost := this.tQueue.GetAverage(t)
			avgQueue := this.lQueue.GetAverage(t)
			fmt.Printf("cost: %d,queue: %d\n", avgCost/int64(time.Millisecond), avgQueue)
		}	
	}()*/
	go func() {
		var delta time.Duration = time.Second * 10
		for {
			select {
			case <-time.After(delta):
			case <-this.ctx.Done():
				return
			}
			var queueLength int64
			var nowQueueLength int64
			t := time.Now().Add(-delta)
			mostCost := this.tQueue.GetMost(t)
			if mostCost > this.limit {
				queueLength = this.lQueue.GetMost(t)
				queueLength = queueLength / (mostCost / this.limit)
				nowQueueLength = atomic.LoadInt64(&this.lengthLimit)
				if nowQueueLength != 0 {
					queueLength = nowQueueLength + ((queueLength - nowQueueLength) * 1 / 2)
				}
				fmt.Printf("- - - %d %d %d\n",
					mostCost/int64(time.Millisecond),
					this.limit/int64(time.Millisecond),
					queueLength)
				atomic.StoreInt64(&this.lengthLimit, queueLength)
			}
		}

	}()
	for {
		select {
		case <-time.After(time.Millisecond * 10):
		case <-this.ctx.Done():
			return
		}
		t := atomic.SwapInt64(&this.t, 0)
		count := atomic.SwapInt64(&this.count, 0)
		queueLength := atomic.LoadInt64(&this.queueLength)
		if count != 0 {
			this.tQueue.Insert(t / count)
			this.lQueue.Insert(queueLength)
		}
	}
}
