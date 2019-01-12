package main

import (
	"golang.org/x/net/context"
	"github.com/tedcy/sheep/src/extends/flow"
	"github.com/tedcy/sheep/src/extends/flow/demo/test"
)

// implement grpc service
type TestApi struct {
	flow.FlowI
}

type Request struct {
	Name		string
}

type Response struct {
	Message		string
}

func (this *TestApi) Handler(ctx context.Context, req *test.TestRequest) (rsp *test.TestResponse, err error) {
	iReq, err := this.reqTrans(req)
	if err != nil {
		return
	}

	iRsp := &Response{}
	ctx = this.Executor(ctx, iReq, iRsp)
	if ctx != nil {
		err = ctx.Err()
		if err != nil {
			return
		}
	}

	rsp, err = this.rspTrans(iRsp)
	return
}

func (this *TestApi) reqTrans(testReq *test.TestRequest) (req *Request, err error) {
	req = &Request{}
	req.Name = testReq.Name
	// 协议转换
	return
}
func (this *TestApi) rspTrans(rsp *Response) (testRsp *test.TestResponse, err error) {
	testRsp = &test.TestResponse{}
	testRsp.Message = rsp.Message
	// 协议转换
	return
}
