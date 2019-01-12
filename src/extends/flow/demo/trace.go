package main

import (
	"golang.org/x/net/context"
	"fmt"
)

type Trace struct {}
func (*Trace) PloyWillRun(ploy interface{}, ctx context.Context, req *Request, rsp *Response) context.Context {
	fmt.Printf("PloyWillRun %s\n",req.Name)
	return ctx
}
func (*Trace) PloyDidRun(ploy interface{}, ctx context.Context, req *Request, rsp *Response) context.Context {
	fmt.Printf("PloyDidRun %s\n",rsp.Message)
	return ctx
}
