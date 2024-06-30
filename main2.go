package main

import (
	"ggcache-plus/core"
	"ggcache-plus/distributekv"
	"ggcache-plus/distributekv/http"
	"ggcache-plus/flags"
	"ggcache-plus/global"
)

func main() {
	core.InitConf()
	global.Log = core.InitLogger()
	global.DB = core.InitGorm()
	core.InitEtcd()

	option := flags.Parse()
	if option.Run() {
		return
	}

	/* if you have a configuration center, both api client and http server configurations can be pulled from the configuration center */
	serverAddrMap := map[int]string{
		10001: "http://127.0.0.1:10001",
		10002: "http://127.0.0.1:10002",
		10003: "http://127.0.0.1:10003",
	}
	var serverAddrs []string
	for _, v := range serverAddrMap {
		serverAddrs = append(serverAddrs, v)
	}

	g := distributekv.NewGroupInstance("scores")
	//  start http api server for client load balancing
	if option.Api {
		go http.StartHTTPAPIServer("http://127.0.0.1:8000", g)
	}
	// start http server to provide caching service
	http.StartHTTPCacheServer(serverAddrMap[option.Port], serverAddrs, g)
}
