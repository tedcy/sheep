package client

import (
	"coding.net/tedcy/sheep/src/client/balancer"
	"coding.net/tedcy/sheep/src/client/breaker_notify"
	"coding.net/tedcy/sheep/src/client/weighter_notify"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type BalancerType int

const (
	DefaultBalancer		BalancerType = iota
	RespTimeBalancer
)

type DialConfig struct {
	EnableBreak		bool
	BalancerType	BalancerType	
	Timeout			time.Duration
	//etcd://172.16.176.38:2379,ip:port
	Target			string
	Path			string
}

func DialContext(ctx context.Context, config *DialConfig, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error){
	c := &client{}
	c.balancer, err = balancer.New(config.Path, config.Timeout)
	if err != nil {
		return
	}
	if config.EnableBreak {
		c.enableBreak()
	}
	c.withBalanceType(config.BalancerType)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBalancer(c.balancer))
	opts = append(opts, grpc.WithUnaryInterceptor(c.clientIntercept()))
	conn, err = grpc.DialContext(ctx, config.Target, opts...)
	if err != nil {
		return
	}
	return
}

type client struct {
	breaker  breaker_notify.BreakerNotifyI
	weighter weighter_notify.WeighterNotifyI
	balancer *balancer.Balancer
	//打开熔断器时插入aop
	//使用balance返回aop不为空就插入
	intercepts []grpc.UnaryClientInterceptor
}

func (this *client) enableBreak() {
	this.breaker = breaker_notify.New()
	this.balancer.SetNotifyOpen(this.breaker.NotifyOpen())
	this.balancer.SetNotifyClose(this.breaker.NotifyClose())
	this.balancer.SetNotifyHalfOpen(this.breaker.NotifyHalfOpen())
	this.intercepts = append(this.intercepts, this.breaker.GrpcUnaryClientInterceptor)
}

func (this *client) withBalanceType(t BalancerType) {
	switch t {
	case RespTimeBalancer:
		this.weighter = weighter_notify.New(weighter_notify.RespTimeWeighter)
	}
	this.balancer.SetNotifyWeighterChange(this.weighter.NotifyWeighterChange())
	this.intercepts = append(this.intercepts, this.weighter.GrpcUnaryClientInterceptor)
}

func (this *client) clientIntercept() grpc.UnaryClientInterceptor {
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
