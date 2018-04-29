package main

import (
	"fmt"
	"coding.net/tedcy/sheep/src/client"
	"golang.org/x/net/context"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
    "time"
)

func main() {
	c := &client.DialConfig{}
	c.EnableBreak = true
	c.BalancerType = client.RespTimeBalancer
	c.Target = "etcd://172.16.176.38:2379"
	c.Path = "/test1"
	c.Timeout = time.Second * 3
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 2)
	realConn := pb.NewGreeterClient(conn)
	resp, err := realConn.SayHello(context.Background(), &pb.HelloRequest{Name : "name"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("resp: " + resp.Message)
}

