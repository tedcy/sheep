package main

import (
	"golang.org/x/net/context"
	sheep_server "github.com/tedcy/sheep/server"
	"github.com/tedcy/sheep/limiter"
	"github.com/tedcy/sheep/server/real_server/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	real_grpc "google.golang.org/grpc"
	"time"
)

type server struct {
}

var gResp *pb.HelloReply = &pb.HelloReply{Message: "Hello"}
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return gResp, nil
}

func main() {
	config := &sheep_server.ServerConfig{}
	config.Addr = ":50051"
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
	s, err := sheep_server.New(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer s.Close()
	h, ok := s.GetRegisterHandler().(*real_grpc.Server)
	if !ok {
		panic("invalid grpc handler")
	}
	pb.RegisterGreeterServer(h, realS)
	if err := s.Serve(); err != nil {
		panic(err)
	}
}
