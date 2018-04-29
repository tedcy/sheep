package breaker_notify

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"github.com/sony/gobreaker"
	"sync"
)

type BreakerNotifyI interface{
	GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	NotifyOpen() <-chan string
	NotifyHalfOpen() <-chan string
	NotifyClose() <-chan string
}

func New() BreakerNotifyI{
	b := &breaker_notify{}
	b.notifyOpen = make(chan string)
	b.notifyHalfOpen = make(chan string)
	b.notifyClose = make(chan string)
	return b
}

type breaker_notify struct{
	breakers		sync.Map
	notifyOpen		chan string
	notifyHalfOpen	chan string
	notifyClose		chan string
}

func (this *breaker_notify) getBreaker(addr string) (b *gobreaker.CircuitBreaker){
	bI, ok := this.breakers.Load(addr)
	if !ok {
		st := gobreaker.Settings{}
		st.Name = addr
		st.OnStateChange = this.newStateChangeCb()
		b = gobreaker.NewCircuitBreaker(st)
		this.breakers.Store(addr, b)
	}else {
		b = bI.(*gobreaker.CircuitBreaker)
	}
	return
}

func (this *breaker_notify) newStateChangeCb() func(name string, from, to gobreaker.State) {
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

func (this *breaker_notify) GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	var p peer.Peer
	opts = append(opts, grpc.Peer(&p))
	realErr := invoker(ctx, method, req, reply, cc, opts...) 
	addr := p.Addr.String()
	b := this.getBreaker(addr)
	_, err = b.Execute(func() (interface{}, error) {return nil, realErr})
	return
}

func (this *breaker_notify) NotifyOpen() <-chan string {
	return this.notifyOpen
}

func (this *breaker_notify) NotifyHalfOpen() <-chan string {
	return this.notifyHalfOpen
}

func (this *breaker_notify) NotifyClose() <-chan string {
	return this.notifyClose
}
