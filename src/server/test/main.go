package main

import (
	"golang.org/x/net/context"
	sheep_server "github.com/tedcy/sheep/src/server"
	sheep_server_grpc "github.com/tedcy/sheep/src/server/real_server/grpc"
	sheep_client "github.com/tedcy/sheep/src/client"
	"github.com/tedcy/sheep/src/client/test"
	"github.com/tedcy/sheep/src/limiter"
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
	config.LimiterType = limiter.InvokeTimeLimiterType
	var i time.Duration
	_ = i
	config.Limit = int64(time.Millisecond * 100)
	//config.Limit = 10000000
	config.Type = "grpc"
	config.Addr = port
	config.WatcherAddrs = "etcd://172.16.176.38:2379"
	config.WatcherPath = "/test1"
	config.WatcherTimeout = time.Second * 3
	config.Opt = &sheep_server_grpc.GrpcServerOpt{
		GrpcOpts : append([]grpc.ServerOption{}, grpc.MaxConcurrentStreams(10000)),
	}
	realS := &server{}
	realS.cb = cb
	s, err := sheep_server.New(context.Background(), config)
	if err != nil {
		panic(err)
	}
	pb.RegisterGreeterServer(s.GetRegisterHandler().(*grpc.Server), realS)
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
	NewSheepServer("127.0.0.1:50051", DefaultCb)
	time.Sleep(time.Millisecond * 500)
	config := &sheep_client.DialConfig{}
	config.TargetPath = "etcd://172.16.176.38:2379/test1"
	err := test.NewClient(10000, config)
	if err != nil {
		panic(err)
	}
	<-c
}
