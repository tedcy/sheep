package main

import (
	"coding.net/tedcy/sheep/src/client"
	"coding.net/tedcy/sheep/src/client/test"
	"golang.org/x/net/context"
	"time"
	"google.golang.org/grpc"
	"sync/atomic"
	"sync"
	"fmt"
)

func bench(gocount int, conn *grpc.ClientConn, do func(*grpc.ClientConn) error) (uint32, time.Duration){
	var count uint32
	var sumT int64
	var delta time.Duration = time.Second * 100
	wg := &sync.WaitGroup{}
	after := time.Now().Add(delta)
	for i := 0;i < gocount;i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var now time.Time
			for ;; {
				now = time.Now()
				do(conn)
				atomic.AddInt64(&sumT, int64(time.Now().Sub(now)))
				atomic.AddUint32(&count, 1)
				if after.Before(now) {
					break
				}
			}
		}()
	}
	wg.Wait()
	qps := count / uint32(delta.Seconds())
	delay := time.Duration(sumT) / time.Duration(count)
	//fmt.Printf("qps: %d delay: %s\n", qps, delay)
	return qps, delay
}

func testBench(name string, conn *grpc.ClientConn) {
	time.Sleep(time.Millisecond * 100)
	
	ss := []int{1,10,100,1000,10000,100000}
	for _, s := range ss {
		qps1, delay1 := bench(s,conn, func(*grpc.ClientConn)error{return nil})
		qps, delay := bench(s,conn, test.B_callOnce)
		fmt.Printf("name:%s c:%d qps:%d delay:%s\n", 
			name,
			s,
			int(1/(1/float64(qps) - 1/float64(qps1))),
			delay - delay1)
	}
}

func TestBenchCall() {
	test.Reinit()
	test.AddList([]string{"127.0.0.1:50051"})
	//test.AddList([]string{"127.0.0.1:50051","127.0.0.1:50052"})
	serverdone := test.Newserver(":50051", test.DefaultCb)
	defer close(serverdone)
	//serverdone = test.Newserver(":50052", test.DefaultCb)
	//defer close(serverdone)
	c := &client.DialConfig{}
	c = test.CpConfig(c)
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	testBench("sheep",conn)	
}

func TestBenchGrpc() {
	test.Reinit()
	serverdone := test.Newserver(":50051", test.DefaultCb)
	defer close(serverdone)
	conn, err := grpc.DialContext(context.Background(), "127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	testBench("grpc",conn)	
}

func main() {
	TestBenchCall()
	TestBenchGrpc()
}
