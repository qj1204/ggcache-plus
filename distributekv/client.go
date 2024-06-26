package distributekv

import (
	"context"
	"ggcache-plus/distributekv/etcd"
	pb "ggcache-plus/distributekv/grpc/ggmemcachedpb"
	"ggcache-plus/global"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// grpcClient grpc客户端（就是一个peer节点，实现了Fetcher接口）
type grpcClient struct {
	serviceName string // 服务名称 ggmemcached/ip:addr
}

// Fetch 通过grpc访问etcd服务注册中心，获取远程节点缓存数据
func (gc *grpcClient) Fetch(group string, key string) ([]byte, error) {
	// 创建一个 etcd client
	cli, err := clientv3.New(etcd.DefaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// 发现服务，取得与服务的链接
	global.Log.Info("grpc 客户端的服务名称为：", gc.serviceName)
	conn, err := etcd.EtcdDial(cli, gc.serviceName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 创建一个 grpc 客户端
	client := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 使用带有超时自动取消的上下文和指定请求调用客户端的 Get 方法发起 rpc 请求调用
	global.Log.Infof("grpc 查询的group: %s，key: %s", group, key)
	resp, err := client.Get(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		global.Log.Errorf("grpc 客户端获取错误：%s", err.Error())
		return nil, err
	}
	return resp.Value, nil
}

func NewClient(service string) *grpcClient {
	return &grpcClient{serviceName: service}
}
