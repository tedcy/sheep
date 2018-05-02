package test

import (
	"coding.net/tedcy/sheep/src/client"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
	"sync"
	"sync/atomic"
	"time"
)

var addrMap *sync.Map

func reinitAddrMap() {
	addrMap = &sync.Map{}
}

func addAddr(addr string) {
	var temp int64
	i, ok := addrMap.LoadOrStore(addr, &temp)
	if ok {
		realI, _ := i.(*int64)
		atomic.AddInt64(realI, 1)
	}
}

func getAddr(addr string) int64 {
	i, ok := addrMap.Load(addr)
	if ok {
		realI, _ := i.(*int64)
		return *realI
	}
	return 0
}

func printResult() {
	addrMap.Range(func(key interface{}, value interface{}) bool {
		realKey, _ := key.(string)
		realValue, _ := value.(*int64)
		fmt.Printf("%s count: %d\n", realKey, *realValue)
		return true
	})
	return
}

func CpConfig(config *client.DialConfig) (c *client.DialConfig) {
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

func newClient(callCounts int, config *client.DialConfig) error {
	c := CpConfig(config)
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond * 100)
	if callCounts == 1 {
		return callOnce(conn)
	}
	for i := 0; i < callCounts; i++ {
		callOnce(conn)
	}
	return nil
}

func callOnce(conn *grpc.ClientConn) error {
	var p peer.Peer
	realConn := pb.NewGreeterClient(conn)
	resp, err := realConn.SayHello(context.Background(), &pb.HelloRequest{Name: "name"}, grpc.Peer(&p))
	if err != nil {
		//fmt.Println(err)
		return err
	}
	if p.Addr != nil {
		addAddr(p.Addr.String())
	}
	_ = resp
	//fmt.Println("resp: " + resp.Message)
	return nil
}

var gReq *pb.HelloRequest = &pb.HelloRequest{Name: "name"}
func B_callOnce(conn *grpc.ClientConn) error {
	pb.NewGreeterClient(conn).SayHello(context.Background(), gReq)
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
