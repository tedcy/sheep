package main

import (
	"fmt"
	"coding.net/tedcy/sheep/src/client"
	"golang.org/x/net/context"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	c, err := client.New()
	if err != nil {
		panic(err)
	}
	c.EnableBreak()
	c.WithBalanceType(client.RespTimeBalancer)

	conn, err := c.DialContext(context.Background(), "")
	if err != nil {
		panic(err)
	}
	realConn := pb.NewGreeterClient(conn)
	resp, err := realConn.SayHello(context.Background(), &pb.HelloRequest{Name : "name"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.Message)
}

