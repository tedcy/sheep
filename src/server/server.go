package server

import (
	"coding.net/tedcy/sheep/src/watcher"
	"coding.net/tedcy/sheep/src/limiter"
	"coding.net/tedcy/sheep/src/server/limiter_wrapper"
	"coding.net/tedcy/sheep/src/server/real_server"
	"coding.net/tedcy/sheep/src/server/real_server/common"
	"golang.org/x/net/context"
	"net"
	"time"
	"strings"
)

type ServerConfig struct {
	Addr					string
	Type					string
	WatcherAddrs			string
	WatcherTimeout			time.Duration
	WatcherPath				string
	LimiterType				limiter.LimiterType
	Limit					int64
	Interceptors			[]common.ServerInterceptor
	Opt						interface{}
}

func New(ctx context.Context, config *ServerConfig) (s *Server,err error) {
	s = &Server{}	
	s.limiterWrapper = limiter_wrapper.New(ctx, config.LimiterType, config.Limit)
	//limiter should insert first
	var interceptors []common.ServerInterceptor
	if s.limiterWrapper != nil {
		interceptors = append([]common.ServerInterceptor{}, s.limiterWrapper.ServerInterceptor)
	}
	interceptors = append(interceptors, config.Interceptors...)
	s.server, err = real_server.New(config.Type, ctx, common.MergeInterceptor(interceptors), config.Opt)
	if err != nil {
		return
	}
	if config.WatcherAddrs != "" {
		s.watcher, err = watcher.New(ctx, &watcher.Config{
			Target:		config.WatcherAddrs,
			Timeout:	config.WatcherTimeout,
		})
		if err != nil {
			return
		}
	}
	//todo add etcd
	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return
	}
	s.lis = lis
	if config.WatcherAddrs != "" {
		config.Addr = checkConfigAddr(s.watcher, config.Addr)
		err = s.watcher.CreateEphemeral(config.WatcherPath + "/" + config.Addr, nil)
	}
	return
}

func checkConfigAddr(watcher watcher.WatcherI, addr string) string{
	if watcher == nil {
		return addr
	}
	ss := strings.SplitN(addr, ":", 2)
	if ss[0] == "" || ss[0] == "0.0.0.0"{
		ss[0] = watcher.GetLocalIp()
	}
	return ss[0] + ":" + ss[1]
}

type Server struct {
	server					real_server.RealServerI
	limiterWrapper			limiter_wrapper.LimiterWrapper
	lis						net.Listener
	watcher					watcher.WatcherI
}

func (this *Server) Serve() error {
	return this.server.Serve(this.lis)
}

func (this *Server) Register(protoDesc, imp interface{}) error{
	return this.server.Register(protoDesc, imp)
}

func (this *Server) GetRegisterHandler() interface{} {
	return this.server.GetRegisterHandler()
}

func (this *Server) Close() (err error){
	err = this.server.Stop()
	if this.limiterWrapper != nil {
		_ = this.limiterWrapper.Close()
	}
	if this.watcher != nil {
		_ = this.watcher.Close()
	}
	return
}
