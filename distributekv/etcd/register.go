package etcd

import (
	"context"
	"errors"
	"fmt"
	"ggcache-plus/global"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"time"
)

// Register 注册一个服务至 etcd
// 注意 Register 将不会 return（如果没有 error 的话）
func Register(service string, addr string, stop chan error) error {
	// 使用默认配置创建一个 etcd client
	cli, err := clientv3.New(global.DefaultEtcdConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("创建etcd客户端失败：%s", err.Error()))
	}
	defer cli.Close()

	// 调用客户端的 Grant 方法创建一个租约，配置 5s 过期
	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return errors.New(fmt.Sprintf("创建租约失败：%s", err.Error()))
	}

	leaseId := resp.ID
	// 注册服务 将服务地址与租约关联
	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		global.Log.Fatalf("服务地址与租约关联失败：%s", err.Error())
		return err
	}

	// 设置服务心跳检测
	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return errors.New(fmt.Sprintf("设置心跳检测失败：%s", err.Error()))
	}
	global.Log.Infof("[%s] 注册服务完成 ok", addr)
	for {
		select {
		case err := <-stop: // 监听服务取消注册的信号
			etcdDel(cli, service, addr)
			if err != nil {
				global.Log.Error(err.Error())
			}
			return err
		case <-cli.Ctx().Done(): // 监听服务被取消的信号
			return errors.New("服务关闭")
		case _, ok := <-ch: // 监听租约撤销信号
			// 监听租约
			if !ok {
				global.Log.Info("保持通道关闭，撤销给定的租约")
				etcdDel(cli, service, addr)
				return errors.New("保持通道关闭，撤销给定的租约")
			}
		default:
			time.Sleep(3 * time.Second)
			//global.Log.Infof("Recv reply from service: %s/%s, ttl:%d", service, addr, resp.TTL)
		}
	}
}

// etcdAdd 以key - value的形式存储在etcd中，key的形式为service/addr, value的形式为endpoint{addr, metadata}
func etcdAdd(client *clientv3.Client, lid clientv3.LeaseID, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	return endPointsManager.AddEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr),
		endpoints.Endpoint{Addr: addr, Metadata: global.Config.GGroupCache.Name + " services"},
		clientv3.WithLease(lid))
}

func etcdDel(client *clientv3.Client, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	return endPointsManager.DeleteEndpoint(client.Ctx(), fmt.Sprintf("%s/%s", service, addr), nil)
}
