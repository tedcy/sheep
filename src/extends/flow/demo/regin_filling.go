package main

import (
	"golang.org/x/net/context"
	"github.com/tedcy/sheep/src/common"
)

type RegionFilling struct {}

func (*RegionFilling) Run(ctx context.Context, req *Request, rsp *Response) context.Context {
	rsp.Message = req.Name
	var err error
	// todo something
	if err != nil {
		return common.WithError(ctx, err)
    }

	return ctx
}
