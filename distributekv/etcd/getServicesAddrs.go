package etcd

import (
	"context"
	"ggcache-plus/global"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// GetPeers 从etcd中获取配置项（服务注册发现）
func GetPeers(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		global.Log.Error("创建etcd客户端失败：", err)
		return []string{}, err
	}

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		global.Log.Error("从etcd获取节点地址列表失败：", err)
		return []string{}, err
	}

	var peers []string
	for _, kv := range resp.Kvs {
		peers = append(peers, string(kv.Value))
	}

	global.Log.Info("从etcd获取节点地址列表成功，列表为：", peers)
	return peers, nil
}
