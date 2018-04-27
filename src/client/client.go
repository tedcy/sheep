package client

import (
	"coding.net/tedcy/sheep/src/balancer"
	"coding.net/tedcy/sheep/src/breaker"
	"coding.net/tedcy/sheep/src/weighter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type BalancerType int

const (
	Default BalancerType = iota
	RespTimeBalancer
)

type Client struct {
	breaker  breaker.BreakerI
	balancer *balancer.Balancer
	weighter weighter.WeighterI
	//打开熔断器时插入aop
	//使用balance返回aop不为空就插入
	intercepts []grpc.UnaryClientInterceptor
}

func New() (*Client, error) {
	return &Client{}, nil
}

func (this *Client) EnableBreak() {
	this.breaker = breaker.New()
	this.balancer.SetNotifyOpen(this.breaker.NotifyOpen())
	this.balancer.SetNotifyClose(this.breaker.NotifyClose())
	this.balancer.SetNotifyHalfOpen(this.breaker.NotifyHalfOpen())
	this.intercepts = append(this.intercepts, this.breaker.GrpcUnaryClientInterceptor)
}

func (this *Client) WithBalanceType(t BalancerType) {
	switch t {
	case RespTimeBalancer:
		this.weighter = weighter.New(weighter.RespTimeWeighter)
	}
	this.balancer.SetNotifyWeighterChange(this.weighter.NotifyWeighterChange())
	this.intercepts = append(this.intercepts, this.weighter.GrpcUnaryClientInterceptor)
}

func (this *Client) Dial(address string) {

}

func (this *Client) ClientIntercept() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, handler grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var index int
		index = -1
		var i grpc.UnaryInvoker
		i = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			index++
			if index >= len(this.intercepts) {
				return handler(ctx, method, req, reply, cc, opts...)
			} else {
				return this.intercepts[index](ctx, method, req, reply, cc, i, opts...)
			}
		}
		return i(ctx, method, req, reply, cc, opts...)
	}
}
