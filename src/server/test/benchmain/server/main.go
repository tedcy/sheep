package main

import (
	"golang.org/x/net/context"
	sheep_server "coding.net/tedcy/sheep/src/server"
	"coding.net/tedcy/sheep/src/server/limiter_wrapper"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/reflection"
	"time"
)

var gResp *pb.HelloReply = &pb.HelloReply{Message: "Hello"}
func DefaultCb(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	for i := 0;i< 1000;i++ {
		c := make(chan struct{})	
		wait(c)
		<-c
	}
	return gResp, nil
}

func wait(c chan<- struct{}) {
	go func() {
		for i := 0;i < 10000;i++ {
			_ = i
		}
		close(c)
	}()
}

type server struct {
	cb func(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error)
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return s.cb(ctx, in)
}

func NewSheepServer(port string, cb func(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error)) (serverDone chan struct{}) {
	config := &sheep_server.ServerConfig{}
	config.LimiterWrapperType = limiter_wrapper.InvokeTimeLimiterWrapperType
	var i time.Duration
	_ = i
	config.Limit = int64(time.Millisecond * 100)
	config.WatcherAddrs = "etcd://172.16.176.38:2379"
	config.WatcherPath = "/test1"
	config.WatcherTimeout = time.Second * 3
	config.GrpcOpts = append(config.GrpcOpts, grpc.MaxConcurrentStreams(10000))
	//config.Limit = 10000000
	config.Addr = port
	realS := &server{}
	realS.cb = cb
	s, err := sheep_server.New(context.Background(), config)
	if err != nil {
		panic(err)
	}
	pb.RegisterGreeterServer(s.Server, realS)
	serverDone = make(chan struct{})
	go func() {
		<-serverDone
		s.Close()
	}()
	go func() {
		if err := s.Serve(); err != nil {
			panic(err)
		}
	}()
	return
}

func main() {
	c := make(chan struct{})
	NewSheepServer(":50051", DefaultCb)
	<-c
}
