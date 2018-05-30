package server

import (
	"coding.net/tedcy/sheep/src/watcher"
	"coding.net/tedcy/sheep/src/server/real_server"
	"golang.org/x/net/context"
	"net"
	"time"
)

type ServerConfig struct {
	Addr					string
	Type					string
	WatcherAddrs			string
	WatcherTimeout			time.Duration
	WatcherPath				string
	Opt						interface{}
}

func New(ctx context.Context, config *ServerConfig) (s *Server,err error) {
	s = &Server{}	
	s.server, err = real_server.New(config.Type, ctx, config.Opt)
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
		err = s.watcher.CreateEphemeral(config.WatcherPath + "/" + config.Addr, nil)
	}
	return
}

type Server struct {
	server					real_server.RealServerI
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
	if this.watcher != nil {
		_ = this.watcher.Close()
	}
	return
}
