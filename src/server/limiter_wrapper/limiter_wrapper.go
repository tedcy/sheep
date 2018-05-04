package limiter_wrapper

import (
	"coding.net/tedcy/sheep/src/limiter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type LimiterWrapperType int

const (
	None								LimiterWrapperType = iota
	QueueLengthLimiterWrapperType		
	InvokeTimeLimiterWrapperType
)

type LimiterWrapper interface {
	UnaryServerInterceptor (ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	Close() error
}

func New(ctx context.Context, t LimiterWrapperType, limit int64) *limiterWrapper{
	l := &limiterWrapper{}
	switch t {
	case QueueLengthLimiterWrapperType:
		l.limiter = limiter.New(ctx, limiter.QueueLengthLimiterType, limit)
	case InvokeTimeLimiterWrapperType:
		l.limiter = limiter.New(ctx, limiter.InvokeTimeLimiterType, limit)
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
