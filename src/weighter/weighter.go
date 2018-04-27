package weighter

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"coding.net/tedcy/sheep/src/common"
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
	return &weighter{}
}

type weighter struct {

}

func (this *weighter) GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}

func (this *weighter) NotifyWeighterChange() <-chan []*common.KV {
	return nil
}
