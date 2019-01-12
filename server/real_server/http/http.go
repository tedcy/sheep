package http

import (
	"github.com/tedcy/sheep/server/real_server/common"
	"golang.org/x/net/context"
	"fmt"
	"net/http"
	"net"
	"strings"
)

//TODO responseWriter代理，两次writerHeader去重
type HttpHandlerI interface{
	Handler(ctx context.Context, req interface{}) (resp interface{}, err error)
	Decode(httpReq *HttpReq) (req interface{},err error)
	Encode(resp interface{}, outputErr error, rw ResponseWriter) (err error)
}

type HttpServer struct {
	server				*http.Server
	interceptor         common.ServerInterceptor
	protoImps			map[string]HttpHandlerI
	protoSlowImps		map[string]HttpHandlerI
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
	s.protoSlowImps = make(map[string]HttpHandlerI)
	s.server = &http.Server{}
	s.interceptor = interceptor
	s.server.Handler = s
	return
}

//Register GET:/path
//Register GET:/path*
func (this *HttpServer) Register(protoDesc interface{}, imp interface{}) error{
	s, ok := protoDesc.(string)
	if !ok {
		return fmt.Errorf("invalid http protoDesc register")
	}
	handler, ok := imp.(HttpHandlerI)
	if !ok {
		return fmt.Errorf("invalid http imp")
	}
	if strings.HasSuffix(s, "*") {
		s = strings.TrimSuffix(s, "*")
		this.protoSlowImps[s] = handler
		return nil
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
		for path, impsHandler := range this.protoSlowImps {
			if strings.HasPrefix(methodPath, path) {
				ok = true
				handler = impsHandler
			}
		}
		if !ok {
			rw.WriteHeader(404)
			return
		}
	}
	realHttpReq := newHttpReq(httpReq)
	defer realHttpReq.Close()
	realHttpRw := newHttpResp(rw)
	defer realHttpRw.Close()

	var resp interface{}
	var err error
	//拦截器报错也会又Encode接口来处理
	defer func() {
		err = handler.Encode(resp, err, realHttpRw)
		if err != nil {
			realHttpRw.WriteHeader(501)
			return
		}
	}()
	req, err := handler.Decode(realHttpReq)
	if err != nil {
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "serviceName", methodPath)
	//拦截器的统计会受到框架报错的影响
	if this.interceptor != nil {
		resp, err = this.interceptor(ctx, req, handler.Handler)
	}else {
		resp, err = handler.Handler(ctx, req)
	}
	return
}

func (this *HttpServer) Stop() error {
	return this.server.Close()	
}
