package breaker

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type BreakerI interface{
	GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	NotifyOpen() <-chan []string
	NotifyHalfOpen() <-chan []string
	NotifyClose() <-chan []string
}

func New() BreakerI{
	return &breaker{}
}

type breaker struct{

}

func (this *breaker) GrpcUnaryClientInterceptor (ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}

func (this *breaker) NotifyOpen() <-chan []string {
	return nil
}

func (this *breaker) NotifyHalfOpen() <-chan []string {
	return nil
}

func (this *breaker) NotifyClose() <-chan []string {
	return nil
}
