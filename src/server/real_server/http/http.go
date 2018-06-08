package http

import (
	"coding.net/tedcy/sheep/src/limiter"
	"golang.org/x/net/context"
	"fmt"
	"net/http"
	"net"
	"io"
)

type HttpHandlerI interface{
	Handler(ctx context.Context, req io.Reader, resp io.Writer) (err error)
}

type HttpServer struct {
	limiter				limiter.LimiterI
	server				*http.Server
	protoImps			map[string]HttpHandlerI
}

type HttpServerOpt struct {
	LimiterType				limiter.LimiterType
	Limit					int64
}

func New(ctx context.Context, opt interface{}) (s *HttpServer, err error){
	o, ok := opt.(*HttpServerOpt)
	if !ok {
		err = fmt.Errorf("invalid http opt type")
		return
	}
	s = &HttpServer{}
	s.protoImps = make(map[string]HttpHandlerI)
	s.limiter = limiter.New(ctx, o.LimiterType, o.Limit)
	s.server = &http.Server{}
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

func (this *HttpServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path
	handler, ok := this.protoImps[method + ":" + path]
	if !ok {
		rw.WriteHeader(404)
		return
	}
	defer req.Body.Close()
	if this.limiter != nil {
		this.limiter.Execute(func()(interface{}, error){
			err := handler.Handler(context.TODO(), req.Body, rw)
			//如果rw没写入header，这里补上
			if err != nil {
				rw.WriteHeader(501)
			}
			return nil, err
		}, nil)
	}else {
		err := handler.Handler(context.TODO(), req.Body, rw)
		//如果rw没写入header，这里补上
		if err != nil {
			rw.WriteHeader(501)
		}
	}
}

func (this *HttpServer) Stop() error {
	return this.server.Close()	
}
