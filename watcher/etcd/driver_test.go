package etcd

import (
	"testing"
	"fmt"
	"time"
)

func Test_List(t *testing.T) {
	c, err := New([]string{"http://172.16.176.38:2379"}, 3 * time.Second)
	if err != nil {
		panic(err)
	}
	nodes, _, err := c.List("/test1")
	if err != nil {
		panic(err)
	}
	fmt.Println(nodes)
}
