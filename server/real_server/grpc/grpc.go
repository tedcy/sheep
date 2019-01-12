package grpc

import (
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/tedcy/sheep/server/real_server/common"
	"fmt"
	"net"
)

type GrpcServer struct {
	server					*grpc.Server
	interceptor				common.ServerInterceptor
}

type GrpcServerOpt struct {
	GrpcOpts				[]grpc.ServerOption
}

func New(ctx context.Context, interceptor common.ServerInterceptor, opt interface{}) (s *GrpcServer, err error){
	opts := []grpc.ServerOption{}
	if opt != nil {
		o, ok := opt.(*GrpcServerOpt)
		if !ok {
			err = fmt.Errorf("invalid grpc opt type")
			return
		}
		opts = o.GrpcOpts
	}
	
	s = &GrpcServer{}
	opts = append(opts, grpc.UnaryInterceptor(s.ServeGrpc))
	s.server = grpc.NewServer(opts...)
	s.interceptor = interceptor
	return
}

//todo 函数类型怎么直接转换？
func (this *GrpcServer) ServeGrpc(ctx context.Context, req interface{}, 
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ctx = context.WithValue(ctx, "serviceName", info.FullMethod)
	if this.interceptor != nil {
		return this.interceptor(ctx, req, 
			func(ctx context.Context, req interface{}) (resp interface{}, err error) {
				return handler(ctx, req)
			})
	}
	return handler(ctx, req)
}

func (this *GrpcServer) Register(protoDesc interface{}, imp interface{}) error{
	//not realize
	return nil
}

func (this *GrpcServer) GetRegisterHandler() interface{} {
	return this.server
}

func (this *GrpcServer) Serve(lis net.Listener) error{
	return this.server.Serve(lis)
}

func (this *GrpcServer) Stop() (err error) {
	this.server.Stop()
	return
}
