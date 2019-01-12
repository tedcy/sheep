package http

import (
	"net/http"
)

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
}

type httpResponse struct{
	realRw http.ResponseWriter
	ifHeaderSent bool
}

func newHttpResp(realRw http.ResponseWriter) *httpResponse {
	return &httpResponse{
		realRw: realRw,
	}
}

func (this *httpResponse) Header() http.Header {
	return this.realRw.Header()
}
func (this *httpResponse) Write(bs []byte) (int, error) {
	if !this.ifHeaderSent {
		this.ifHeaderSent = true
	}
	return this.realRw.Write(bs)
}
//如果已经发过状态码就不再发了
func (this *httpResponse) WriteHeader(statusCode int) {
	if !this.ifHeaderSent {
		this.ifHeaderSent = true
		this.realRw.WriteHeader(statusCode)
	}
}
//如果没发过就随便发一个
func (this *httpResponse) Close() {
	if !this.ifHeaderSent {
		this.realRw.WriteHeader(200)
	}
}
