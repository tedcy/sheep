package weighter_notify

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"coding.net/tedcy/sheep/src/common"
	"time"
)

type WeighterType int
const (
	Default WeighterType			= iota
	RespTimeWeighter
)

type WeighterNotifyI interface{
	GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	NotifyWeighterChange() <-chan []*common.KV
}

func New(t WeighterType) WeighterNotifyI {
	w := &weighter_notify{}
	w.respTime = common.NewAdder()
	w.counts = common.NewAdder()
	w.kvs = make(chan []*common.KV)
	go w.respTimeLooper()
	return w
}

type weighter_notify struct {
	respTime	*common.Adder
	counts		*common.Adder
	kvs			chan []*common.KV
}

func (this *weighter_notify) GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	t := time.Now()
	var p peer.Peer
	opts = append(opts, grpc.Peer(&p))
	err = invoker(ctx, method, req, reply, cc, opts...) 
	delta := time.Now().Sub(t)
	addr := p.Addr.String()
	this.respTime.Add(addr, uint64(delta / time.Millisecond))
	this.counts.Add(addr, 1)
	return
}

//由于weighter_notify不知道watcher信息，所以节点数是可能大于真实节点:有节点下线
//可能小于真实节点数，有节点新上线还没有被调度访问过
func (this *weighter_notify) NotifyWeighterChange() <-chan []*common.KV {
	return this.kvs
}

func (this *weighter_notify) respTimeLooper() {
	for ;; {
		time.Sleep(time.Second * 30)
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
		avgRespTime = this.respTime.Get(key) / this.counts.Get(key)
		sumRespTime += avgRespTime
		ss[key] = avgRespTime
	}
	this.respTime.Clean()
	this.counts.Clean()
	var kvs []*common.KV
	for _, key := range keys {
		kv := &common.KV{}
		kv.Key = key
		kv.Weight = int(sumRespTime - avgRespTime)
		kvs = append(kvs, kv)
	}
	this.kvs <- kvs
}
