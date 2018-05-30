package main

import (
	"golang.org/x/net/context"
	sheep_server "coding.net/tedcy/sheep/src/server"
	"coding.net/tedcy/sheep/src/limiter"
	"coding.net/tedcy/sheep/src/server/real_server/grpc"
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
	config.Addr = "127.0.0.1:50051"
	config.WatcherAddrs = "etcd://172.16.176.38:2379"
	config.WatcherPath = "/test1"
	config.WatcherTimeout = time.Second * 3
	config.Type = "grpc"
	config.Opt = &grpc.GrpcServerOpt{
		LimiterType: limiter.InvokeTimeLimiterType,
		Limit: int64(time.Millisecond * 100),
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
