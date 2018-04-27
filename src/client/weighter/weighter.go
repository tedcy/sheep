package weighter

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"coding.net/tedcy/sheep/src/common"
	"time"
)

type WeighterType int
const (
	Default WeighterType			= iota
	RespTimeWeighter
)

type WeighterI interface{
	GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	NotifyWeighterChange() <-chan []*common.KV
}

func New(t WeighterType) WeighterI {
	w := &weighter{}
	w.respTime = common.NewAdder()
	w.counts = common.NewAdder()
	w.respTimeLooper()
	return w
}

type weighter struct {
	respTime	*common.Adder
	counts		*common.Adder
	kvs			chan []*common.KV
}

func (this *weighter) GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	addr, err := common.GetClietIP(ctx)
	if err != nil {
		return err
	}
	t := time.Now()
	err = invoker(ctx, method, req, reply, cc, opts...) 
	delta := time.Now().Sub(t)
	this.respTime.Add(addr, uint64(delta / time.Millisecond))
	this.counts.Add(addr, 1)
	return err
}

//由于weighter不知道watcher信息，所以节点数是可能大于真实节点:有节点下线
//可能小于真实节点数，有节点新上线还没有被调度访问过
func (this *weighter) NotifyWeighterChange() <-chan []*common.KV {
	return this.kvs
}

func (this *weighter) respTimeLooper() {
	for ;; {
		time.Sleep(time.Second * 30)
		this.scanServers()
	}
}

//see
//https://github.com/Netflix/ribbon/blob/master/ribbon-loadbalancer/src/main/java/com/netflix/loadbalancer/ResponseTimeWeightedRule.java
func (this *weighter) scanServers() {
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
