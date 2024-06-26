package main

import (
	"fmt"
	"ggcache-plus/core"
	"ggcache-plus/distributekv"
	"ggcache-plus/distributekv/etcd"
	"ggcache-plus/flags"
	"ggcache-plus/global"
)

func main() {
	core.InitConf()
	global.Log = core.InitLogger()
	global.DB = core.InitGorm()

	option := flags.Parse()
	if option.Run() {
		return
	}

	// 新建 cache 实例
	group := distributekv.NewGroupInstance("scores")

	// New 一个自己实现的服务实例
	addr := fmt.Sprintf("localhost:%d", option.Port)
	svr, err := distributekv.NewServer(addr)
	if err != nil {
		global.Log.Fatal(err)
	}

	// 设置同伴节点包括自己（同伴的地址从 etcd 中获取）
	addrs, err := etcd.GetPeers("clusters")
	if err != nil { // 如果查询失败使用默认的地址
		addrs = []string{"localhost:10001"}
	}

	global.Log.Info("从 etcd 处获取的 server 地址: ", addrs)
	// 将节点打到哈希环上
	svr.SetPeers(addrs)
	// 为 Group 注册服务 Picker
	group.RegisterPeers(svr)
	global.Log.Info("ggmemcached 运行在: ", addr)

	// 启动服务（注册服务至 etcd、计算一致性 hash）
	err = svr.Start()
	if err != nil {
		global.Log.Fatal(err)
	}
}
