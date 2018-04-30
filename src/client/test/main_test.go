package main

import (
	"coding.net/tedcy/sheep/src/client"
	"coding.net/tedcy/sheep/src/watcher/test"
	"testing"
)

var listNotify = make(chan []string)
var watchNotify = make(chan struct{})

func init() {
	test.DefaultList(listNotify)
	test.DefaultWatch(watchNotify)
}

func Test_WatcherInit(t *testing.T) {
	serverdone := newserver(":50051", defaultCb)
	defer close(serverdone)
	go func() {
		go func() { listNotify <- []string{"127.0.0.1:50051"} }()
		go func() { watchNotify <- struct{}{} }()
	}()
	c := &client.DialConfig{}
	err := newClient(c)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
