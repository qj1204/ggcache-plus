package main

import (
	"ggcache-plus/core"
	"ggcache-plus/distributekv"
	"ggcache-plus/distributekv/etcd"
	"ggcache-plus/distributekv/grpc"
	"ggcache-plus/flags"
	"ggcache-plus/global"
	"strconv"
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

	// 新建 cache 实例
	group := distributekv.NewGroupInstance("scores")

	// New 一个自己实现的服务实例
	addr := "localhost:" + strconv.Itoa(option.Port)
	updateChan := make(chan bool)
	svr, err := grpc.NewServer(updateChan, addr)
	if err != nil {
		global.Log.Fatal(err)
		return
	}

	go etcd.DynamicServices(updateChan, global.Config.GGroupCache.Name)

	// 设置同伴节点包括自己（同伴的地址从 etcd 中获取）
	addrs, err := etcd.ListServicePeers(global.Config.GGroupCache.Name)
	if err != nil { // 如果查询失败使用默认的地址
		addrs = global.Config.GGroupCache.Addr
	}
	global.Log.Info("从 etcd 处获取的 server 地址: ", addrs)

	// 将节点打到哈希环上
	svr.SetPeers(addrs)

	// 为 Group 注册服务 Picker
	group.RegisterPeers(svr)
	global.Log.Info(global.Config.GGroupCache.Name + " 运行在: " + addr)

	// 启动服务（注册服务至 etcd、计算一致性 hash）
	err = svr.Start()
	if err != nil {
		global.Log.Fatal(err)
	}
}
