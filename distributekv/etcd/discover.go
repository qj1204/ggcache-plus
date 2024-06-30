package etcd

import (
	"context"
	"fmt"
	"ggcache-plus/global"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
)

// Discovery 创建到给定服务的客户端连接
func Discovery(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	// NewBuilder 创建一个解析器生成器。用于解析客户端发来的请求路径，从而确认要连接的对象
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		global.Log.Error("etcd解析器错误：", err)
		return nil, err
	}
	global.Log.Info("etcd连接的服务为：", service) // GGroupCache
	return grpc.Dial(
		"etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)), // 轮询
	)
}

// EtcdDial 节点间的grpc通信
func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	// NewBuilder 创建一个解析器生成器。用于解析客户端发来的请求路径，从而确认要连接的对象
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		global.Log.Error("etcd解析器错误：", err)
		return nil, err
	}
	return grpc.Dial(
		strings.Split(service, "/")[1],
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// ListServicePeers 根据名称查找可用服务节点列表
func ListServicePeers(prefix string) ([]string, error) {
	cli, err := clientv3.New(global.DefaultEtcdConfig)
	if err != nil {
		global.Log.Errorf("未能连接到etcd，%v", err)
		return []string{}, err
	}

	endPointsManager, err := endpoints.NewManager(cli, prefix)
	if err != nil {
		global.Log.Errorf("创建服务节点管理器失败，%v", err)
		return []string{}, err
	}

	Key2EndpointMap, err := endPointsManager.List(context.Background())
	if err != nil {
		global.Log.Errorf("服务节点管理器加载列表失败，%v", err)
		return []string{}, err
	}

	var peers []string
	for key, endpoint := range Key2EndpointMap {
		peers = append(peers, endpoint.Addr)
		global.Log.Infof("找到服务节点 [%s] (%s):(%s)", key, endpoint.Addr, endpoint.Metadata)
	}

	return peers, nil
}

func DynamicServices(update chan bool, service string) {
	cli, err := clientv3.New(global.DefaultEtcdConfig)
	if err != nil {
		global.Log.Errorf("未能连接到etcd，%v", err)
		return
	}
	defer cli.Close()

	// Subscription and publishing mechanism
	watchChan := cli.Watch(context.Background(), service, clientv3.WithPrefix())

	// 每次用户往指定的服务中添加或者删除新的实例地址时，watchChan 后台都能通过 WithPrefix() 扫描到实例数量的变化并以  watchResp.Events 事件的方式返回
	// 当发生变更时，往 update channel 发送一个信号，告知 endpoint manager 重新构建哈希映射
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- true // 通知 endpoint manager 重新构建节点视图
				global.Log.Warnf("服务节点更新: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- true // 通知 endpoint manager 重新构建节点视图
				global.Log.Warnf("服务节点移除: %s", string(ev.Kv.Key))
			}
		}
	}
}
