package grpc

import (
	"coding.net/tedcy/sheep/src/limiter"
	"coding.net/tedcy/sheep/src/server/real_server/grpc/limiter_wrapper"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"fmt"
	"net"
)

type GrpcServer struct {
	limiterWrapper			limiter_wrapper.LimiterWrapper
	server					*grpc.Server
}

type GrpcServerOpt struct {
	LimiterType				limiter.LimiterType
	Limit					int64
	GrpcOpts				[]grpc.ServerOption
}

func New(ctx context.Context, opt interface{}) (s *GrpcServer, err error){
	o, ok := opt.(*GrpcServerOpt)
	if !ok {
		err = fmt.Errorf("invalid grpc opt type")
		return
	}
	s = &GrpcServer{}
	s.limiterWrapper = limiter_wrapper.New(ctx, o.LimiterType, o.Limit)
	if s.limiterWrapper != nil {
		o.GrpcOpts = append(o.GrpcOpts, grpc.UnaryInterceptor(s.limiterWrapper.UnaryServerInterceptor))
	}
	s.server = grpc.NewServer(o.GrpcOpts...)
	return
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
	if this.limiterWrapper != nil {
		this.limiterWrapper.Close()
	}
	return
}
