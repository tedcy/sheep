package main

import (
	"coding.net/tedcy/sheep/src/client"
	"coding.net/tedcy/sheep/src/client/test"
	//"coding.net/tedcy/sheep/src/client/test/benchmain/kettytest"
	"coding.net/tedcy/sheep/src/common/bench"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

func TestBenchCall() {
	bConfig := &bench.BenchConfig{}
	bConfig.InitFunc = func() (data interface{}, cs []chan<-struct{}){
		test.Reinit()
		test.AddList([]string{"127.0.0.1:50051"})
		//test.AddList([]string{"127.0.0.1:50051","127.0.0.1:50052"})
		serverdone := test.Newserver(":50051", test.DefaultCb)
		cs = append(cs, serverdone)
		//serverdone = test.Newserver(":50052", test.DefaultCb)
		//cs = append(cs, serverdone)
		c := &client.DialConfig{}
		c = test.CpConfig(c)
		conn, err := client.DialContext(context.Background(), c)
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
	bConfig.Name = "sheep"
	bConfig.Time = time.Second * 60
	bench.New(bConfig).Run()
}

func TestBenchGrpc() {
	bConfig := &bench.BenchConfig{}
	bConfig.InitFunc = func() (data interface{}, cs []chan<-struct{}){
		test.Reinit()
		serverdone := test.Newserver(":50051", test.DefaultCb)
		cs = append(cs, serverdone)
		conn, err := grpc.DialContext(context.Background(), "127.0.0.1:50051", grpc.WithInsecure())
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
	bench.New(bConfig).Run()
}

func main() {
	//kettytest.TestBenchKetty()
	TestBenchCall()
	TestBenchGrpc()
}
