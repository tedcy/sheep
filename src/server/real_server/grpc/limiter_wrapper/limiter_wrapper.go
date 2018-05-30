package limiter_wrapper

import (
	"coding.net/tedcy/sheep/src/limiter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type LimiterWrapper interface {
	UnaryServerInterceptor (ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	Close() error
}

func New(ctx context.Context, t limiter.LimiterType, limit int64) *limiterWrapper{
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

func (this *limiterWrapper) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = this.limiter.Execute(func() (interface{}, error) {
		return handler(ctx, req)
	}, nil)
	return
}

func (this *limiterWrapper) Close() error {
	this.cancel()
	return nil
}
