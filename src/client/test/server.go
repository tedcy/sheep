package main

import (
	"log"
	"net"

	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"
	"time"
)

type server struct {
	cb func(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error)
}

func defaultCb(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func slowCb(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	time.Sleep(time.Millisecond * 50)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func errCb(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return nil, fmt.Errorf("test err")
}

func afterTimeErr2Success() func(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	t := time.NewTimer(time.Second * 50)
	var b bool
	go func() {
		<-t.C
		b = true
	}()
	return func(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
		if b {
			return slowCb(ctx, in)
		}
		return errCb(ctx, in)
	}
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return s.cb(ctx, in)
}

func newserver(port string, cb func(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error)) (serverDone chan struct{}) {
	realS := &server{}
	realS.cb = cb
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, realS)
	reflection.Register(s)
	serverDone = make(chan struct{})
	go func() {
		<-serverDone
		s.Stop()
	}()
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return
}
