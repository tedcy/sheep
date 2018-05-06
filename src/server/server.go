package server

import (
	"coding.net/tedcy/sheep/src/server/limiter_wrapper"
	"coding.net/tedcy/sheep/src/watcher"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerConfig struct {
	LimiterWrapperType		limiter_wrapper.LimiterWrapperType
	Addr					string
	Limit					int64
	GrpcOpts				[]grpc.ServerOption
	WatcherAddrs			string
	WatcherTimeout			time.Duration
	WatcherPath				string
}

func New(ctx context.Context, config *ServerConfig) (s *Server,err error) {
	s = &Server{}	
	s.watcher, err = watcher.New(ctx, &watcher.Config{
		Target:		config.WatcherAddrs,
		Timeout:	config.WatcherTimeout,
	})
	if err != nil {
		return
	}
	s.limiterWrapper = limiter_wrapper.New(ctx, config.LimiterWrapperType, config.Limit)
	if s.limiterWrapper != nil {
		config.GrpcOpts = append(config.GrpcOpts, grpc.UnaryInterceptor(s.limiterWrapper.UnaryServerInterceptor))
	}
	s.Server = grpc.NewServer(config.GrpcOpts...)
	//todo add etcd
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return
	}
	s.lis = lis
	err = s.watcher.CreateEphemeral(config.WatcherPath + "/" + config.Addr, nil)
	return
}

type Server struct {
	limiterWrapper			limiter_wrapper.LimiterWrapper
	Server					*grpc.Server
	lis						net.Listener
	watcher					watcher.WatcherI
}

func (this *Server) Serve() error {
	return this.Server.Serve(this.lis)
}

func (this *Server) Close() error{
	this.Server.Stop()
	if this.watcher != nil {
		_ = this.watcher.Close()
	}
	return this.limiterWrapper.Close()
}
