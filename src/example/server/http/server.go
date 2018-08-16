package main

import (
	"golang.org/x/net/context"
	sheep_server "coding.net/tedcy/sheep/src/server"
	"coding.net/tedcy/sheep/src/server/real_server/http"
	"coding.net/tedcy/sheep/src/limiter"
	"io/ioutil"
)

type server struct {
}

func (s *server) Decode(httpReq *http.HttpReq) (req interface{},err error) {
	b, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		return
	}
	req = b
	return
}

func (s *server) Encode(resp interface{}, outputErr error, rw http.ResponseWriter) error {
	if outputErr != nil {
		rw.Write([]byte(outputErr.Error()))
		return outputErr
	}
	_, err := rw.Write(resp.([]byte))
	return err
}

func (s *server) Handler(ctx context.Context, req interface{}) (resp interface{}, err error) {
	resp = req
	return
}

func main() {
	config := &sheep_server.ServerConfig{}
	config.Addr = "127.0.0.1:80"
	config.Type = "http"
	config.LimiterType = limiter.InvokeTimeLimiterType
	config.Limit = 100
	realS := &server{}
	s, err := sheep_server.New(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	err = s.Register("GET:/test", realS)
	if err != nil {
		panic(err)
	}
	if err := s.Serve(); err != nil {
		panic(err)
	}
}
