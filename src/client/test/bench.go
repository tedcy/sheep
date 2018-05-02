package test

import (
	"testing"
	"coding.net/tedcy/sheep/src/client"
	"golang.org/x/net/context"
	"time"
	"google.golang.org/grpc"
	"sync/atomic"
	"sync"
)

func bench(conn *grpc.ClientConn, do func(*grpc.ClientConn) error) (uint32, time.Duration){
	var count uint32
	var sumT int64
	var tcount time.Duration = 60
	c1 := make(chan struct{})
	wg := &sync.WaitGroup{}
	for i := 0;i < 10000;i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			now := time.Now()
			for ;; {
				select {
				case <-c1:
					atomic.AddInt64(&sumT, int64(time.Now().Sub(now)))
					return
				default:
				}
				do(conn)
				atomic.AddUint32(&count, 1)
			}
		}()
	}
	t1 := time.After(time.Second * tcount)
	go func() {
		<-t1
		close(c1)
	}()
	wg.Wait()
	qps := count / uint32(time.Duration(sumT).Seconds())
	delay := time.Duration(sumT) / time.Duration(count)
	print(qps)
	print(" - ")
	println(delay.String())
	return qps, delay
}

func testBench(conn *grpc.ClientConn) {
	time.Sleep(time.Millisecond * 100)
	
	qps1, delay1 := bench(conn, func(*grpc.ClientConn)error{return nil})
	qps, delay := bench(conn, B_callOnce)
	print("qps: ")
	println(int(1/(1/float64(qps) - 1/float64(qps1))))
	print((delay - delay1).String())
}

func TestBenchCall(t *testing.T) {
	Reinit()
	AddList([]string{"127.0.0.1:50051","127.0.0.1:50052"})
	serverdone := Newserver(":50051", defaultCb)
	defer close(serverdone)
	serverdone = Newserver(":50052", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	c = CpConfig(c)
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	testBench(conn)	
}

func TestBenchGrpc(t *testing.T) {
	Reinit()
	AddList([]string{"127.0.0.1:50051"})
	serverdone := Newserver(":50051", defaultCb)
	defer close(serverdone)
	conn, err := grpc.DialContext(context.Background(), "127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond * 100)
	testBench(conn)	
}

func BenchmarkCall(b *testing.B) {
	b.StopTimer()
	Reinit()
	AddList([]string{"127.0.0.1:50051"})
	serverdone := Newserver(":50051", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	//c.Target = "test://"
	c = CpConfig(c)
	conn, err := client.DialContext(context.Background(), c)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond * 100)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
	//for i := 0; i < b.N; i++ {
		for pb.Next() {
			B_callOnce(conn)
		}
	//}
	})
}

func BenchmarkGrpc(b *testing.B) {
	b.StopTimer()
	Reinit()
	serverdone := Newserver(":50051", defaultCb)
	defer close(serverdone)
	conn, err := grpc.DialContext(context.Background(), "127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond * 100)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		//for i := 0; i < b.N; i++ {
		for pb.Next() {
			B_callOnce(conn)
		}
		//}
	})
}


