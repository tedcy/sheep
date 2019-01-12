package real_server

import (
	"github.com/tedcy/sheep/src/server/real_server/grpc"
	"github.com/tedcy/sheep/src/server/real_server/http"
	"github.com/tedcy/sheep/src/server/real_server/common"
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

func New(stype string, ctx context.Context, interceptor common.ServerInterceptor, opt interface{}) (RealServerI, error) {
	switch stype {
	case "http":
		return http.New(ctx, interceptor, opt)
	case "grpc":
		return grpc.New(ctx, interceptor, opt)
	default:
		return nil, fmt.Errorf("invalid opt type %s", stype)
	}
}
