package common

import (
	"google.golang.org/grpc/peer"
	"golang.org/x/net/context"
	"fmt"
	"net"
	"strings"
)

type KV struct {
	Key		string
	Weight	int
}

func GetClietIP(ctx context.Context) (string, error) {
    pr, ok := peer.FromContext(ctx)
    if !ok {
        return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
    }
    if pr.Addr == net.Addr(nil) {
        return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
    }
    addSlice := strings.Split(pr.Addr.String(), ":")
    return addSlice[0], nil
}
