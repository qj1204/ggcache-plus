package etcd

import (
	"ggcache-plus/global"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
)

// EtcdDial 向 grpc 请求一个服务
// 通过提供一个 etcd client 和 service name 即可获取 grpc 通道
func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	// NewBuilder 创建一个解析器生成器。用于解析客户端发来的请求路径，从而确认要连接的对象
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		global.Log.Error("etcd 解析器错误：", err)
		return nil, err
	}

	// Dial 创建到给定目标的客户端连接（有了通道就可以建立与服务端的连接了）
	// WithResolvers 允许在 ClientConn 本地注册一系列解析器实现，而无需通过 resolver.Register 进行全局注册。
	// 它们将仅与当前 Dial 使用的方案进行匹配，并优先于全局注册。
	global.Log.Info("grpc 连接的服务端地址为：", service) // ggmemcached/localhost:10003
	return grpc.Dial(
		strings.Split(service, "/")[1], // localhost:10003
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}
