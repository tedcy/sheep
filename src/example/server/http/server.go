package main

import (
	"golang.org/x/net/context"
	sheep_server "coding.net/tedcy/sheep/src/server"
	"coding.net/tedcy/sheep/src/limiter"
	"io"
	"io/ioutil"
)

type server struct {
}

func (s *server) Decode(in io.Reader) (req interface{}, err error) {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return
	}
	req = b
	return
}

//func (s *server) Encode(resp interface{}, out io.Writer) error {
func (s *server) Encode(resp interface{}, out io.Writer) error {
	_, err := out.Write(resp.([]byte))
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
