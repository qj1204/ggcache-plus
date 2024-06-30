package http

import (
	"ggcache-plus/distributekv"
	"ggcache-plus/global"
	"net/http"
)

func StartHTTPCacheServer(addr string, addrs []string, group *distributekv.Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs)
	group.RegisterPeers(peers)
	global.Log.Infof("service is running at %v", addr)
	global.Log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// StartHTTPAPIServer todo: gin 路由拆分请求负载
func StartHTTPAPIServer(apiAddr string, group *distributekv.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			val, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(val.ByteSlice())
		}))
	global.Log.Infof("fontend server is running at %v", apiAddr)
	global.Log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
