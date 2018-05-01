package main

import (
	"testing"
	"coding.net/tedcy/sheep/src/client"
	"golang.org/x/net/context"
	"time"
	"google.golang.org/grpc"
)

func BenchmarkCall(b *testing.B) {
	b.StopTimer()
	reinit()
	addList([]string{"127.0.0.1:50051"})
	serverdone := newserver(":50051", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	//c.Target = "test://"
	c = cpConfig(c)
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond * 100)
	b.SetParallelism(10000)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
	//for i := 0; i < b.N; i++ {
		for pb.Next() {
			benchmark_callOnce(conn)
		}
	//}
	})
}

func BenchmarkGrpc(b *testing.B) {
	b.StopTimer()
	reinit()
	serverdone := newserver(":50051", defaultCb)
	defer close(serverdone)
	conn, err := grpc.DialContext(context.Background(), "127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond * 100)
	b.SetParallelism(10000)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		//for i := 0; i < b.N; i++ {
		for pb.Next() {
			benchmark_callOnce(conn)
		}
		//}
	})
}


