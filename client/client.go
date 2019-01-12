package client

import (
	"github.com/tedcy/sheep/client/balancer"
	"github.com/tedcy/sheep/client/breaker_notify"
	"github.com/tedcy/sheep/client/weighter_notify"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
	"strings"
)

type BalancerType int

const (
	DefaultBalancer BalancerType = iota
	RespTimeBalancer
)

type DialConfig struct {
	EnableBreak  bool
	BalancerType BalancerType
	Timeout      time.Duration
	//etcd://172.16.176.38:2379,ip:port/path
	TargetPath string
}

func splitTargetPath(targetPath string) (target, path string) {
	index := strings.LastIndex(targetPath, "/")
	target = targetPath[:index]
	path = targetPath[index:]
	return
}

func DialContext(ctx context.Context, config *DialConfig, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	target, path := splitTargetPath(config.TargetPath)
	c := &client{}
	c.ctx = ctx
	c.Balancer, err = balancer.New(ctx, path, config.Timeout)
	if err != nil {
		return
	}
	if config.EnableBreak {
		c.enableBreak()
	}
	c.withBalanceType(config.BalancerType)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBalancer(c))
	opts = append(opts, grpc.WithUnaryInterceptor(c.clientIntercept()))
	conn, err = grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return
	}
	return
}

type client struct {
	*balancer.Balancer
	breaker  breaker_notify.BreakerNotifyI
	weighter weighter_notify.WeighterNotifyI
	//打开熔断器时插入aop
	//使用balance返回aop不为空就插入
	intercepts []grpc.UnaryClientInterceptor
	ctx        context.Context
}

func (this *client) enableBreak() {
	this.breaker = breaker_notify.New()
	this.Balancer.SetNotifyOpen(this.breaker.NotifyOpen())
	this.Balancer.SetNotifyClose(this.breaker.NotifyClose())
	this.Balancer.SetNotifyHalfOpen(this.breaker.NotifyHalfOpen())
	this.intercepts = append(this.intercepts, this.breaker.GrpcUnaryClientInterceptor)
}

func (this *client) withBalanceType(t BalancerType) {
	switch t {
	case RespTimeBalancer:
		this.weighter = weighter_notify.New(this.ctx, weighter_notify.RespTimeWeighter)
	case DefaultBalancer:
		this.weighter = weighter_notify.New(this.ctx, weighter_notify.DefaultWeighter)
	}
	if this.weighter != nil {
		this.Balancer.SetNotifyWeighterChange(this.weighter.NotifyWeighterChange())
		this.intercepts = append(this.intercepts, this.weighter.GrpcUnaryClientInterceptor)
	}
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

//will call by grpc.ClientConn Close()
func (this *client) Close() (err error) {
	if this.breaker != nil {
		err = this.breaker.Close()
		if err != nil {
			println(err)
		}
	}
	if this.weighter != nil {
		err = this.weighter.Close()
		if err != nil {
			println(err)
		}
	}
	err = this.Balancer.Close()
	if err != nil {
		println(err.Error())
	}
	return nil
}
