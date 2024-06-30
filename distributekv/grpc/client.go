package grpc

import (
	"context"
	"ggcache-plus/distributekv/etcd"
	pb "ggcache-plus/distributekv/grpc/ggroupcachepb"
	"ggcache-plus/global"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// grpcClient grpc客户端（就是一个peer节点，实现了Fetcher接口）
type grpcClient struct {
	serviceName string // 服务名称 GGroupCache/localhost:10002
}

// Fetch 通过grpc访问etcd服务注册中心，获取远程节点缓存数据
func (gc *grpcClient) Fetch(group string, key string) ([]byte, error) {
	// 创建一个 etcd client
	cli, err := clientv3.New(global.DefaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// 发现服务，取得与服务的链接
	start := time.Now()
	conn, err := etcd.EtcdDial(cli, gc.serviceName)
	global.Log.Warnf("本次 grpc dial 的耗时为: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 创建一个 grpc 客户端
	client := pb.NewGGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// 使用带有超时自动取消的上下文和指定请求调用客户端的 Get 方法发起 rpc 请求调用
	resp, err := client.Get(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	global.Log.Warnf("本次 grpc Call 的耗时为: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		global.Log.Errorf("grpc 客户端获取错误：%s", err.Error())
		return nil, err
	}
	return resp.Value, nil
}
