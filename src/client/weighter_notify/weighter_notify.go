package weighter_notify

import (
	"coding.net/tedcy/sheep/src/common"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"time"
)

type WeighterType int

const (
	Default WeighterType = iota
	RespTimeWeighter
)

type WeighterNotifyI interface {
	GrpcUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	NotifyWeighterChange() <-chan []*common.KV
	Close() error
}

func New(ctx context.Context, t WeighterType) WeighterNotifyI {
	w := &weighter_notify{}
	w.respTime = common.NewAdder()
	w.counts = common.NewAdder()
	w.kvs = make(chan []*common.KV)
	w.ctx, w.cancel = context.WithCancel(ctx)
	go w.respTimeLooper()
	return w
}

type weighter_notify struct {
	respTime *common.Adder
	counts   *common.Adder
	kvs      chan []*common.KV
	ctx      context.Context
	cancel   context.CancelFunc
}

func (this *weighter_notify) GrpcUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	t := time.Now()
	var p peer.Peer
	opts = append(opts, grpc.Peer(&p))
	err = invoker(ctx, method, req, reply, cc, opts...)
	delta := time.Now().Sub(t)
	if err == nil {
		//不报错的节点才进行统计调用时间权值
		addr := p.Addr.String()
		this.respTime.Add(addr, uint64(delta/time.Microsecond))
		this.counts.Add(addr, 1)
		return
	}
	return
}

//由于weighter_notify不知道watcher信息，所以节点数是可能大于真实节点:有节点下线
//可能小于真实节点数，有节点新上线还没有被调度访问过
func (this *weighter_notify) NotifyWeighterChange() <-chan []*common.KV {
	return this.kvs
}

func (this *weighter_notify) respTimeLooper() {
	for {
		select {
		case <-this.ctx.Done():
			return
		case <-time.NewTimer(time.Second * 30).C:
		}
		this.scanServers()
	}
}

//see
//https://github.com/Netflix/ribbon/blob/master/ribbon-loadbalancer/src/main/java/com/netflix/loadbalancer/ResponseTimeWeightedRule.java
func (this *weighter_notify) scanServers() {
	keys := this.counts.List()
	var avgRespTime uint64
	var sumRespTime uint64
	ss := make(map[string]uint64)
	for _, key := range keys {
		respTime := this.respTime.Get(key)
		count := this.counts.Get(key)
		avgRespTime = respTime / count
		sumRespTime += avgRespTime
		ss[key] = avgRespTime
	}
	if sumRespTime == 0 {
		//所有节点都被断流时等下一轮统计
		println("wait next time")
		return
	}
	//防止只有一个key时使得weight归零
	sumRespTime++
	this.respTime.Clean()
	this.counts.Clean()
	var kvs []*common.KV
	for _, key := range keys {
		kv := &common.KV{}
		kv.Key = key
		kv.Weight = int(sumRespTime - ss[key])
		kvs = append(kvs, kv)
	}
	this.kvs <- kvs
}

func (this *weighter_notify) Close() error {
	this.cancel()
	close(this.kvs)
	this.kvs = nil
	return nil
}
