package limiter_wrapper

import (
	"github.com/tedcy/sheep/limiter"
	"github.com/tedcy/sheep/server/real_server/common"
	"golang.org/x/net/context"
)

type LimiterWrapper interface {
	ServerInterceptor (ctx context.Context, req interface{}, handler common.ServerHandler) (resp interface{}, err error)
	Close() error
}

func New(ctx context.Context, t limiter.LimiterType, limit int64) LimiterWrapper{
	l := &limiterWrapper{}
	switch t {
	case limiter.QueueLengthLimiterType:
		l.limiter = limiter.New(ctx, t, limit)
	case limiter.InvokeTimeLimiterType:
		l.limiter = limiter.New(ctx, t, limit)
	default:
		return nil
	}
	l.ctx, l.cancel = context.WithCancel(ctx)
	return l
}

type limiterWrapper struct {
	ctx			context.Context	
	cancel		context.CancelFunc
	limiter		limiter.LimiterI
}

func (this *limiterWrapper) ServerInterceptor(ctx context.Context, req interface{}, handler common.ServerHandler) (resp interface{}, err error) {
	resp, err = this.limiter.Execute(func() (interface{}, error) {
		return handler(ctx, req)
	}, nil)
	return
}

func (this *limiterWrapper) Close() error {
	this.cancel()
	return nil
}
