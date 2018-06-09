package main

import (
	"coding.net/tedcy/sheep/src/client/test"
	"coding.net/tedcy/sheep/src/common/bench"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

func TestBenchCallSheepServer() {
	bConfig := &bench.BenchConfig{}
	bConfig.InitFunc = func() (data interface{}, cs []chan<-struct{}){
		test.Reinit()
		conn, err := grpc.DialContext(context.Background(), "172.16.186.217:50051", grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		clientdone := make(chan struct{})
		go func() {
			<-clientdone
			conn.Close()
		}()
		cs = append(cs, clientdone)
		data = conn
		time.Sleep(time.Millisecond * 100)
		return
	}
	bConfig.BenchFunc = test.CreateBenchCall()
	bConfig.Name = "grpc"
	bConfig.Time = time.Second * 60
	bConfig.Goroutines = []int{10,10,10,10,10,10,10}
	bench.New(bConfig).Run()
}

func main() {
	TestBenchCallSheepServer()
}