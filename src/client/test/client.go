package test

import (
	"coding.net/tedcy/sheep/src/client"
	"coding.net/tedcy/sheep/src/watcher/test"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
	"sync"
	"sync/atomic"
	"time"
)

var listNotify = make(chan []string)
var watchNotify = make(chan struct{})

func Reinit() {
	time.Sleep(time.Second)
	reinitAddrMap()
	close(listNotify)
	close(watchNotify)
	listNotify = make(chan []string)
	watchNotify = make(chan struct{})
	test.DefaultList(listNotify)
	test.DefaultWatch(watchNotify)
}

func AddList(list []string) {
	go func() { listNotify <- list }()
	go func() { watchNotify <- struct{}{} }()
}

var addrMap *sync.Map = &sync.Map{}

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
	c.TargetPath = "test://0.0.0.0/test1"
	if config.TargetPath != "" {
		c.TargetPath = config.TargetPath
	}
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

func NewClient(callCounts int, config *client.DialConfig) error {
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
func CreateBenchCall() func(interface{}) error{
	var conn *grpc.ClientConn
	return func (c interface{}) error{
		if conn == nil {
			conn = c.(*grpc.ClientConn)
		}
		_, err := pb.NewGreeterClient(conn).SayHello(context.Background(), gReq)
		return err	
	}
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
