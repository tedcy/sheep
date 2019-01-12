package main

import (
	"golang.org/x/net/context"
	sheep_server "github.com/tedcy/sheep/server"
	"github.com/tedcy/sheep/limiter"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/tedcy/sheep/server/real_server/grpc"
	real_grpc "google.golang.org/grpc"
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
	config.Addr = port
	config.WatcherAddrs = "etcd://172.16.176.38:2379"
	config.WatcherPath = "/test1"
	config.WatcherTimeout = time.Second * 3
	config.Type = "grpc"
	config.LimiterType = limiter.InvokeTimeLimiterType
	config.Limit = 100
	config.Opt = &grpc.GrpcServerOpt{
		GrpcOpts: append([]real_grpc.ServerOption{}, real_grpc.MaxConcurrentStreams(10000)),
	}
	realS := &server{}
	realS.cb = cb
	s, err := sheep_server.New(context.Background(), config)
	if err != nil {
		panic(err)
	}
	h, ok := s.GetRegisterHandler().(*real_grpc.Server)
	if !ok {
		panic("invalid grpc handler")
	}
	pb.RegisterGreeterServer(h, realS)
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
