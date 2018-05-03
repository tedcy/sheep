package limiter
//限流器和熔断器的区别
//限流器是对单个请求来做限流的，没有状态转换，越在上层拦截越好
//熔断器可以不在上层，因为只需要拥有统计功能，对外汇报状态就可以让其他层进行熔断，是可以在后层的
//所以限流器更适合服务端，而熔断器更适合客户端

import (
	"errors"
	"time"
	"coding.net/tedcy/sheep/src/common"
	"sync/atomic"
)

var ErrDefaultFallBack error = errors.New("default fallback")
var ErrInvokerIsNil error = errors.New("invoker is nil")

type LimiterI interface {
	Execute(invoker,fallback func()(interface{}, error)) (interface{}, error)
}

type LimiterType int

const (
	QueueLengthLimiterType		LimiterType = iota
	InvokeTimeLimiterType
)

func New(t LimiterType, limit int64) LimiterI {
	switch t {
	case QueueLengthLimiterType:
		return NewQueueLengthLimiter(int32(limit))
	case InvokeTimeLimiterType:
		return NewInvokeTimeLimter(limit)
	}
	return nil
}

func NewQueueLengthLimiter(limit int32) (*QueueLengthLimiter){
	l := &QueueLengthLimiter{}
	l.limit = limit	
	return l
}

type QueueLengthLimiter struct{
	queueLength		int32	
	limit			int32
}

func (this *QueueLengthLimiter) Execute(invoker,fallback func()(interface{}, error)) (resp interface{},err error) {
	defer atomic.AddInt32(&this.queueLength, -1)
	if atomic.AddInt32(&this.queueLength, 1) > this.limit {
		if fallback != nil {
			resp, err = fallback()
		}else {
			err = ErrDefaultFallBack
		}
	}else {
		if invoker != nil {
			resp, err = invoker()
		}else {
			err = ErrInvokerIsNil
		}
	}
	
	return
}

func NewInvokeTimeLimter(limit int64) (*InvokeTimeLimiter) {
	l := &InvokeTimeLimiter{}
	l.limit = limit
	return l
}

type InvokeTimeLimiter struct {
	queue			*common.SimpleQueue
	limit			int64
	t				int64
	count			int64
}

func (this *InvokeTimeLimiter) Execute(invoker,fallback func()(interface{}, error)) (resp interface{},err error) {
	if this.queue.GetAverage(20) > this.limit {
		if fallback != nil {
			resp, err = fallback()
		}else {
			err = ErrDefaultFallBack
		}
	}else {
		if invoker != nil {
			t := time.Now()
			resp, err = invoker()	
			delta := time.Now().Sub(t)
			atomic.AddInt64(&this.t, int64(delta))
			atomic.AddInt64(&this.count, 1)
		}else {
			err = ErrInvokerIsNil
		}
	}
	
	return
}

func (this *InvokeTimeLimiter) timeLooper() {
	for ;; {
		time.Sleep(time.Millisecond * 100)
		t := atomic.SwapInt64(&this.t, 0)
		count := atomic.SwapInt64(&this.count, 0)
		if count == 0 {
			continue
		}
		this.queue.Insert(t/count)
	}
}
