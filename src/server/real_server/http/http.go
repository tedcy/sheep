package http

import (
	"coding.net/tedcy/sheep/src/server/real_server/common"
	"golang.org/x/net/context"
	"fmt"
	"net/http"
	"net"
	"io"
)

type HttpHandlerI interface{
	Handler(ctx context.Context, req interface{}) (resp interface{}, err error)
	Decode(r io.Reader) (req interface{},err error)
	Encode(resp interface{}, rw io.Writer) (err error)
}

type HttpServer struct {
	server				*http.Server
	interceptor         common.ServerInterceptor
	protoImps			map[string]HttpHandlerI
}

type HttpServerOpt struct {
}

func New(ctx context.Context, interceptor common.ServerInterceptor, opt interface{}) (s *HttpServer, err error){
	/*o, ok := opt.(*HttpServerOpt)
	if !ok {
		err = fmt.Errorf("invalid http opt type")
		return
	}*/
	s = &HttpServer{}
	s.protoImps = make(map[string]HttpHandlerI)
	s.server = &http.Server{}
	s.interceptor = interceptor
	s.server.Handler = s
	return
}

//Register GET:path
func (this *HttpServer) Register(protoDesc interface{}, imp interface{}) error{
	s, ok := protoDesc.(string)
	if !ok {
		return fmt.Errorf("invalid http protoDesc register")
	}
	handler, ok := imp.(HttpHandlerI)
	if !ok {
		return fmt.Errorf("invalid http imp")
	}
	this.protoImps[s] = handler
	return nil
}

func (this *HttpServer) GetRegisterHandler() interface{} {
	return nil
}

func (this *HttpServer) Serve(lis net.Listener) error{
	return this.server.Serve(lis)
}

func (this *HttpServer) ServeHTTP(rw http.ResponseWriter, httpReq *http.Request) {
	methodPath := httpReq.Method + ":" + httpReq.URL.Path
	handler, ok := this.protoImps[methodPath]
	if !ok {
		rw.WriteHeader(404)
		return
	}
	if httpReq.Body == nil {
		rw.WriteHeader(501)
		return
	}
	defer httpReq.Body.Close()
	req, err := handler.Decode(httpReq.Body)
	if err != nil {
		rw.WriteHeader(501)
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "serviceName", methodPath)
	var resp interface{}
	if this.interceptor != nil {
		resp, err = this.interceptor(ctx, req, handler.Handler)
	}else {
		resp, err = handler.Handler(ctx, req)
	}
	//如果rw没写入header，这里补上
	if err != nil {
		rw.WriteHeader(501)
	}
	err = handler.Encode(resp, rw)
	if err != nil {
		rw.WriteHeader(501)
	}
	return
}

func (this *HttpServer) Stop() error {
	return this.server.Close()	
}
