package server

import (
	"coding.net/tedcy/sheep/src/server/limiter_wrapper"
	//"coding.net/tedcy/sheep/src/watcher"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
)

type ServerConfig struct {
	LimiterWrapperType		limiter_wrapper.LimiterWrapperType
	Addr					string
	Limit					int64
}

func New(ctx context.Context, config *ServerConfig) (*Server, error) {
	s := &Server{}	
	s.limiterWrapper = limiter_wrapper.New(ctx, config.LimiterWrapperType, config.Limit)
	if s.limiterWrapper != nil {
		s.Server = grpc.NewServer(grpc.UnaryInterceptor(s.limiterWrapper.UnaryServerInterceptor))
	}
	//todo add etcd
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}
	s.lis = lis
	return s, nil
}

type Server struct {
	limiterWrapper			limiter_wrapper.LimiterWrapper
	Server					*grpc.Server
	lis						net.Listener
}

func (this *Server) Serve() error {
	return this.Server.Serve(this.lis)
}

func (this *Server) Close() error{
	this.Server.Stop()
	return this.limiterWrapper.Close()
}
