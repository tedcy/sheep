package flow

import (
	"golang.org/x/net/context"
	sheep_server "github.com/tedcy/sheep/src/server"
)

type Server struct {
	server		*sheep_server.Server
}

//grpc多个服务可以共用同一个Server，通过pb.RegisterGreeterServer(grpcServer)
//http多个服务可以公用同一个
func NewServer(config *sheep_server.ServerConfig) (server *Server, err error) {
	server = &Server{}
	server.server, err = sheep_server.New(context.Background(), config)
	if err != nil {
		return
	}
	return
}

func (this *Server) NewFlow(implement interface{}) (flow FlowI, err error){
	flow = NewBaseFlow()
	err = setInterface(implement, flow, "FlowI")
	if err != nil {
		return
	}
	i, ok := implement.(_init)
	if ok {
		i.Init()
	}
	
	return
}

func (this *Server) NewOverLappedFlow(implement interface{}, flow FlowI) (err error) {
	err = setInterface(implement, flow, "FlowI")
	if err != nil {
		return
	}
	i, ok := implement.(_init)
	if ok {
		i.Init()
	}
	return
}

func (this *Server) GetRegisterServer() *sheep_server.Server {
	return this.server
}

func (this *Server) Serve() (err error){
	return this.server.Serve()
}
