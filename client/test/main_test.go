package test

import (
	"github.com/tedcy/sheep/client"
	"github.com/tedcy/sheep/watcher/test"
	"testing"
	"time"
)

func Test_WatcherInit(t *testing.T) {
	Reinit()
	AddList([]string{"127.0.0.1:50051"})
	serverdone := Newserver(":50051", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	err := NewClient(1, c)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_WatcherChange(t *testing.T) {
	Reinit()
	AddList([]string{"127.0.0.1:50051"})
	AddList([]string{"127.0.0.1:50052"})
	serverdone := Newserver(":50052", defaultCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	err := NewClient(1, c)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

//服务器故障导致开路
func Test_BreakerOpen(t *testing.T) {
	count := 1000
	Reinit()
	AddList([]string{"127.0.0.1:50051", "127.0.0.1:50052"})
	serverdone := Newserver(":50051", defaultCb)
	defer close(serverdone)
	serverdone = Newserver(":50052", errCb)
	defer close(serverdone)
	c := &client.DialConfig{}
	NewClient(count, c)
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
	Reinit()
	AddList([]string{"127.0.0.1:50051", "127.0.0.1:50052"})
	serverdone := Newserver(":50051", slowCb)
	defer close(serverdone)
	serverdone = Newserver(":50052", afterTimeErr2Success())
	defer close(serverdone)
	c := &client.DialConfig{}
	NewClient(count, c)
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
	Reinit()
	AddList([]string{"127.0.0.1:50051", "127.0.0.1:50052"})
	serverdone := Newserver(":50051", createSlowCb(time.Millisecond*50))
	defer close(serverdone)
	serverdone = Newserver(":50052", createSlowCb(time.Millisecond*100))
	defer close(serverdone)
	c := &client.DialConfig{}
	NewClient(count, c)
	printResult()
	//count1 := getAddr("127.0.0.1:50051")
	count2 := getAddr("127.0.0.1:50052")
	if count2 > int64(float64(count)*0.5) {
		t.Fatalf("slowCb's calls %d can't bigger than sum*0.5", count2)
	}
}
