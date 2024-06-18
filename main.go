package main

import (
	"flag"
	"fmt"
	"ggcache-plus/distributekv"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *distributekv.Group {
	// 创建一个 Group 实例，该实例使用 GetterFunc 作为回调函数。在缓存未命中时，将会调用该函数从数据源获取数据
	return distributekv.NewGroup("scores", 2<<10, distributekv.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group *distributekv.Group) {
	httpPool := distributekv.NewHTTPPool(addr)
	httpPool.Set(addrs...)
	group.RegisterPeers(httpPool)
	log.Println("distributekv is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], httpPool))
}

func startAPIServer(apiAddr string, group *distributekv.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func startCacheServerGrpc(addr string, addrs []string, group *distributekv.Group) {
	grpcPool := distributekv.NewGrpcPool(addr)
	grpcPool.Set(addrs...)
	group.RegisterPeers(grpcPool)
	log.Println("distributekv is running at", addr)
	grpcPool.Run()
}

func startGRPCServer() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "distributekv server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: ":8001",
		8002: ":8002",
		8003: ":8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServerGrpc(addrMap[port], addrs, group)
}

func main() {
	//var port int
	//var api bool
	//flag.IntVar(&port, "port", 8001, "distributekv server port")
	//flag.BoolVar(&api, "api", false, "Start a api server?")
	//flag.Parse()
	//
	//apiAddr := "http://localhost:9999"
	//addrMap := map[int]string{
	//	8001: "http://localhost:8001",
	//	8002: "http://localhost:8002",
	//	8003: "http://localhost:8003",
	//}
	//
	//var addrs []string
	//for _, v := range addrMap {
	//	addrs = append(addrs, v)
	//}
	//
	//group := createGroup()
	//if api {
	//	go startAPIServer(apiAddr, group)
	//}
	//startCacheServer(addrMap[port], addrs, group)

	startGRPCServer()
}
