package breaker_wrapper

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/sony/gobreaker"
	"coding.net/tedcy/sheep/src/common"
	"sync"
)

type BreakerI interface{
	GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	NotifyOpen() <-chan string
	NotifyHalfOpen() <-chan string
	NotifyClose() <-chan string
}

func New() BreakerI{
	b := &breaker_wrapper{}
	b.notifyOpen = make(chan string)
	b.notifyHalfOpen = make(chan string)
	b.notifyClose = make(chan string)
	return b
}

type breaker_wrapper struct{
	breakers		sync.Map
	notifyOpen		chan string
	notifyHalfOpen	chan string
	notifyClose		chan string
}

func (this *breaker_wrapper) getBreaker(addr string) (b *gobreaker.CircuitBreaker){
	bI, ok := this.breakers.Load(addr)
	if !ok {
		st := gobreaker.Settings{}
		st.Name = addr
		st.OnStateChange = this.newStateChangeCb()
		b = gobreaker.NewCircuitBreaker(st)
		this.breakers.Store(addr, b)
	}
	b = bI.(*gobreaker.CircuitBreaker)
	return
}

func (this *breaker_wrapper) newStateChangeCb() func(name string, from, to gobreaker.State) {
	return func(name string, from, to gobreaker.State) {
		switch from {
		case gobreaker.StateOpen:
			switch to {
			case gobreaker.StateClosed:
				this.notifyClose <- name
			default:
				panic("wtf")
			}
		case gobreaker.StateHalfOpen:
			switch to {
			case gobreaker.StateOpen:
				this.notifyOpen <- name
			case gobreaker.StateClosed:
				this.notifyClose <- name
			default:
				panic("wtf")
			}
		case gobreaker.StateClosed:
			switch to {
			case gobreaker.StateHalfOpen:
				this.notifyHalfOpen <- name
			default:
				panic("wtf")
			}
		}
	}
}

func (this *breaker_wrapper) GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	addr, err := common.GetClietIP(ctx)
	if err != nil {
		return err
	}
	b := this.getBreaker(addr)
	_, err = b.Execute(func() (interface{}, error) {
		return nil, invoker(ctx, method, req, reply, cc, opts...) 
	})
	return err
}

func (this *breaker_wrapper) NotifyOpen() <-chan string {
	return this.notifyOpen
}

func (this *breaker_wrapper) NotifyHalfOpen() <-chan string {
	return this.notifyHalfOpen
}

func (this *breaker_wrapper) NotifyClose() <-chan string {
	return this.notifyClose
}
