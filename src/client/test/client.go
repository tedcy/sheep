package main

import (
	"fmt"
	"coding.net/tedcy/sheep/src/client"
	"golang.org/x/net/context"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc"
    "time"
)

func cpConfig(config *client.DialConfig) (c *client.DialConfig) {
	c = &client.DialConfig{}
	c.EnableBreak = true
	c.BalancerType = client.RespTimeBalancer
	c.Target = "test://"
	if config.BalancerType != 0 {
		c.BalancerType = config.BalancerType
	}
	if config.EnableBreak != false {
		c.EnableBreak = config.EnableBreak
	}
	if config.Timeout != 0 {
		c.Timeout = config.Timeout
	}
	return
}

func newClient(config *client.DialConfig) error{
	c := cpConfig(config)
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond * 100)
	return callOnce(conn)
}

func callOnce(conn *grpc.ClientConn) error{
	realConn := pb.NewGreeterClient(conn)
	resp, err := realConn.SayHello(context.Background(), &pb.HelloRequest{Name : "name"})
	if err != nil {
		fmt.Println(err)
		return err
	}
	_ = resp
	//fmt.Println("resp: " + resp.Message)
	return nil
}

func callUntilOk(conn *grpc.ClientConn) {
	for {
		err := callOnce(conn)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
}
