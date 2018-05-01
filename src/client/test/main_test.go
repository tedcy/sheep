package main

import (
	"coding.net/tedcy/sheep/src/client"
	"coding.net/tedcy/sheep/src/watcher/test"
	"testing"
	"time"
)

var listNotify = make(chan []string)
var watchNotify = make(chan struct{})

func reinit() {
	reinitAddrMap()
	close(listNotify)
	close(watchNotify)
	listNotify = make(chan []string)
	watchNotify = make(chan struct{})
	test.DefaultList(listNotify)
	test.DefaultWatch(watchNotify)
}

func addList(list []string) {
	go func() { listNotify <- list }()
	go func() { watchNotify <- struct{}{} }()
}

func Test_WatcherInit(t *testing.T) {
	reinit()
	addList([]string{"127.0.0.1:50051"})
	serverdone := newserver(":50051", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	err := newClient(1, c)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_WatcherChange(t *testing.T) {
	reinit()
	addList([]string{"127.0.0.1:50051"})
	addList([]string{"127.0.0.1:50052"})
	serverdone := newserver(":50052", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	err := newClient(1, c)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

//服务器故障导致开路
func Test_BreakerOpen(t *testing.T) {
	count := 1000
	reinit()
	addList([]string{"127.0.0.1:50051", "127.0.0.1:50052"})
	serverdone := newserver(":50051", defaultCb)
	defer close(serverdone)
	serverdone = newserver(":50052", errCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	newClient(count, c)
	printResult()
	count1 := getAddr("127.0.0.1:50051")
	//count2 := getAddr("127.0.0.1:50052")
	if count1 < int64(float64(count)*0.8) {
		t.Fatalf("defaultCb's calls %d can't smaller than count*0.8", count1)
	}
}

//服务器故障又恢复进入半开路状态
//bug:底层库必须调用的时候才判断超过多少时间变成开路状态
//测试正常的话是server2被开路，半开路，变成闭路
func Test_BreakerHalfOpen(t *testing.T) {
	count := 2000
	reinit()
	addList([]string{"127.0.0.1:50051", "127.0.0.1:50052"})
	serverdone := newserver(":50051", slowCb)
	defer close(serverdone)
	serverdone = newserver(":50052", afterTimeErr2Success())
	defer close(serverdone)
	c := &client.DialConfig{}
	newClient(count, c)
	printResult()
	//count1 := getAddr("127.0.0.1:50051")
	count2 := getAddr("127.0.0.1:50052")
	if count2 < int64(float64(count)*0.05) {
		t.Fatalf("defaultCb's calls %d can't smaller than count*0.05", count2)
	}
}

//服务器时延变化
func Test_WeightChange(t *testing.T) {
	count := 1000
	reinit()
	addList([]string{"127.0.0.1:50051", "127.0.0.1:50052"})
	serverdone := newserver(":50051", createSlowCb(time.Millisecond*50))
	defer close(serverdone)
	serverdone = newserver(":50052", createSlowCb(time.Millisecond*100))
	defer close(serverdone)
	c := &client.DialConfig{}
	newClient(count, c)
	printResult()
	//count1 := getAddr("127.0.0.1:50051")
	count2 := getAddr("127.0.0.1:50052")
	if count2 > int64(float64(count)*0.5) {
		t.Fatalf("slowCb's calls %d can't bigger than sum*0.5", count2)
	}
}
