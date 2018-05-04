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
	"golang.org/x/net/context"
)

var ErrDefaultFallBack error = errors.New("default fallback")
var ErrInvokerIsNil error = errors.New("invoker is nil")

type LimiterI interface {
	Execute(invoker,fallback func()(interface{}, error)) (interface{}, error)
	Close() error
}

type LimiterType int

const (
	QueueLengthLimiterType		LimiterType = iota
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

func NewQueueLengthLimiter(ctx context.Context, limit int32) (*QueueLengthLimiter){
	l := &QueueLengthLimiter{}
	l.limit = limit	
	l.ctx, l.cancel = context.WithCancel(ctx)
	go l.timeLooper()
	return l
}

type QueueLengthLimiter struct{
	queueLength		int32	
	limit			int32
	ctx				context.Context
	cancel			context.CancelFunc
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

func (this *QueueLengthLimiter) timeLooper() {
	for ;; {
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

func NewInvokeTimeLimter(ctx context.Context, limit int64) (*InvokeTimeLimiter) {
	l := &InvokeTimeLimiter{}
	l.limit = limit
	l.ctx, l.cancel = context.WithCancel(ctx)
	l.queue = common.NewSimpleQueue()
	go l.timeLooper()
	return l
}

type InvokeTimeLimiter struct {
	queue			*common.SimpleQueue
	limit			int64
	t				int64
	count			int64
	queueLength		int32
	ctx				context.Context
	cancel			context.CancelFunc
}

func (this *InvokeTimeLimiter) Execute(invoker,fallback func()(interface{}, error)) (resp interface{},err error) {
	defer atomic.AddInt32(&this.queueLength, -1)
	if atomic.AddInt32(&this.queueLength, 1) > int32(this.limit) {
	//if this.queue.GetAverage(5) > this.limit {
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

func (this *InvokeTimeLimiter) Close() error {
	if this.cancel != nil {
		this.cancel()
	}
	return nil
}

func (this *InvokeTimeLimiter) timeLooper() {
	go func() {
		for ;; {
			select {
				case <-time.After(time.Second):
				case <-this.ctx.Done():
				return
			}
			println(atomic.LoadInt32(&this.queueLength))
			println(time.Duration(this.queue.GetAverage(10)).String())
		}
	}()
	for ;; {
		select {
		case <-time.After(time.Millisecond * 100):
		case <-this.ctx.Done():
			return
		}
		t := atomic.SwapInt64(&this.t, 0)
		count := atomic.SwapInt64(&this.count, 0)
		if count == 0 {
			this.queue.Insert(0)
		}else {
			this.queue.Insert(t/count)
		}
	}
}
