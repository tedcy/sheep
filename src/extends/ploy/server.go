package ploy

import (
	"fmt"
	"strings"
	"sync"
)

type Server struct {
	sUrl		string
	sDriverUrl	string
	servers		[]ketty.Server
	lock		sync.Mutex
}

//grpc多个服务可以共用同一个Server，通过pb.RegisterGreeterServer(grpcServer)
//http多个服务也可以公用同一个
func NewServer(sUrl, sDriverUrl string) (server *Server, err error) {
	server = &Server{}
	server.sUrl = sUrl
	server.sDriverUrl = sDriverUrl
	return
}

func (this *Server) NewFlow(router string, implement interface{}) (flow FlowI, err error){
	this.lock.Lock()
	defer this.lock.Unlock()
	var server ketty.Server
	server, err = ketty.Listen(this.sUrl + router, this.sDriverUrl)
	if err != nil {
		return
	}
	
	err = server.RegisterMethod(h.GetHandle(), implement)
	if err != nil {
		return
	}
	flow = NewBaseFlow()
	err = setInterface(implement, flow, "FlowI")
	if err != nil {
		return
	}
	i, ok := implement.(_init)
	if ok {
		i.Init()
	}
	
	this.servers = append(this.servers, server)
	return
}

func (this *Server) NewOverLappedFlow(router string, implement interface{}, flow FlowI) (err error) {
	h, ok := implement.(getHandle)
	if !ok {
		err = fmt.Errorf("not implement getHandle")
		return
	}
	server, err := ketty.Listen(this.sUrl + router, this.sDriverUrl)
	if err != nil {
		return
	}
	err = server.RegisterMethod(h.GetHandle(), implement)
	if err != nil {
		return
	}
	err = setInterface(implement, flow, "FlowI")
	if err != nil {
		return
	}

	this.servers = append(this.servers, server)
	return
}

func (this *Server) Serve() (err error){
	for i, s := range this.servers {
		err = s.Serve()
		if err != nil {
			return
		}
	}
	return
}
