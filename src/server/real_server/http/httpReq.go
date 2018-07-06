package http

import (
	"net/http"
	"io"
)

type HttpMap interface{
	Get(key string) string
}

type HttpReq struct {
	req *http.Request
	Headers HttpMap
	QueryStrs HttpMap 
	Body io.Reader
}

func newHttpReq(req *http.Request) (httpReq *HttpReq){
	httpReq = &HttpReq{}
	httpReq.req = req
	httpReq.Body = req.Body
	httpReq.Headers = req.Header
	httpReq.QueryStrs = req.URL.Query()
	return httpReq
}

func (this *HttpReq) Close() {
	if this.req.Body != nil {
		this.req.Body.Close()	
	}
}
