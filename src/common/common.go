package common

import (
	"google.golang.org/grpc/peer"
	//"golang.org/x/net/context"
	"fmt"
	"net"
	"errors"
	"strings"
)

var ErrNoAvailableClients = errors.New("no available clients")

type KV struct {
	Key		string
	Weight	int
}

func GetClietIP(pr *peer.Peer) (string, error) {
	//return "127.0.0.1:50051", nil
    if pr.Addr == net.Addr(nil) {
        return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
    }
    addSlice := strings.Split(pr.Addr.String(), ":")
    return addSlice[0], nil
}

/*func GetClietIP(ctx context.Context) (string, error) {
	//return "127.0.0.1:50051", nil
    pr, ok := peer.FromContext(ctx)
    if !ok {
        return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
    }
    if pr.Addr == net.Addr(nil) {
        return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
    }
    addSlice := strings.Split(pr.Addr.String(), ":")
    return addSlice[0], nil
}*/
