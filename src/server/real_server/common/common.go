package common

import (
	"golang.org/x/net/context"
)

type ServerHandler func(ctx context.Context, req interface{}) (resp interface{}, err error)
type ServerInterceptor func(ctx context.Context, req interface{}, handler ServerHandler) (resp interface{}, err error)

func MergeInterceptor(is []ServerInterceptor) (result ServerInterceptor) {
	return func(ctx context.Context, req interface{}, handler ServerHandler) (interface{}, error){
		var index int
		index = -1
		var h ServerHandler
		h = func(ctx context.Context, req interface{}) (interface{}, error) {
			index++
			if index >= len(is) {
				return handler(ctx, req)
			}else {
				return is[index](ctx, req, h)
			}
		}
		return h(ctx, req)
	}
}
