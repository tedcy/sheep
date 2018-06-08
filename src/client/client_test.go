package client

import (
	"testing"
	"fmt"
)

func Test_splitTargetPath(t *testing.T) {
	fmt.Println(splitTargetPath("etcd://172.16.176.38:2379,ip:port/path"))
}
