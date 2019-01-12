package main

import (
	sheep_server "github.com/tedcy/sheep/server"
	"github.com/tedcy/sheep/extends/flow"
	"github.com/tedcy/sheep/common"
)

func main() {
	var config sheep_server.ServerConfig
	//err := config.Read(&config, "server.toml")
	//common.Assert(err)
	config.Type = "grpc"
	config.Addr = ":8000"

	server, err := flow.NewServer(&config)
	common.Assert(err)

	testApi := new(TestApi)
	_, err = server.NewFlow(testApi)
	common.Assert(err)
	initBidding(testApi.FlowI)

	err = server.Serve()
	common.Assert(err)

	common.Hung()
}


func initBidding(flow flow.FlowI) {
	flow.AddTrace(new(Trace))
	flow.AddPloy(new(RegionFilling))
	flow.AddPloy(new(RegionFilling1))
}

