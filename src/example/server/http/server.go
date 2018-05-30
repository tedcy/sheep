package main

import (
	"golang.org/x/net/context"
	sheep_server "coding.net/tedcy/sheep/src/server"
	"coding.net/tedcy/sheep/src/limiter"
	"coding.net/tedcy/sheep/src/server/real_server/http"
	"time"
	"io"
	"io/ioutil"
)

type server struct {
}

func (s *server) Handler(ctx context.Context, in io.Reader, out io.Writer) (err error) {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return
	}
	out.Write(b)
	return
}

func main() {
	config := &sheep_server.ServerConfig{}
	config.Addr = "127.0.0.1:80"
	config.Type = "http"
	config.Opt = &http.HttpServerOpt{
		LimiterType: limiter.InvokeTimeLimiterType,
		Limit: int64(time.Millisecond * 100),
	}
	realS := &server{}
	s, err := sheep_server.New(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	s.Register("POST:/test", realS)
	if err := s.Serve(); err != nil {
		panic(err)
	}
}
