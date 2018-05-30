package real_server

import (
	"coding.net/tedcy/sheep/src/server/real_server/grpc"
	"coding.net/tedcy/sheep/src/server/real_server/http"
	"golang.org/x/net/context"
	"net"
	"fmt"
)

type RealServerI interface {
	Register(protoDesc interface{},imp interface{}) error
	GetRegisterHandler() interface{}
	Serve(lis net.Listener) error
	Stop() error
}

func New(stype string, ctx context.Context, opt interface{}) (RealServerI, error) {
	switch stype {
	case "http":
		return http.New(ctx, opt)
	case "grpc":
		return grpc.New(ctx, opt)
	default:
		return nil, fmt.Errorf("invalid opt type %s", stype)
	}
}
