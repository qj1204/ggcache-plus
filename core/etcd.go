package core

import (
	"ggcache-plus/global"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func InitEtcd() {
	global.DefaultEtcdConfig = clientv3.Config{
		Endpoints:   global.Config.Etcd.Address,
		DialTimeout: time.Duration(global.Config.Etcd.TTL) * time.Second,
	}
}
