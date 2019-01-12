package main

import (
	"golang.org/x/net/context"
)

type RegionFilling1 struct {}

func (*RegionFilling1) Run(ctx context.Context, req *Request, rsp *Response) context.Context {
	return ctx
}
