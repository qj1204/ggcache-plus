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

// DefaultEtcdConfig register 模块提供服务注册至 etcd 的能力
var (
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// etcdAdd 以租约模式添加一对 kv 至 etcd
func etcdAdd(client *clientv3.Client, lid clientv3.LeaseID, service string, addr string) error {
	em, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	//return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoints.Endpoint{Addr: addr})
	return em.AddEndpoint(client.Ctx(), fmt.Sprintf("%s/%s", service, addr), endpoints.Endpoint{Addr: addr, Metadata: "ggmemcached services"}, clientv3.WithLease(lid))
}

// Register 注册一个服务至 etcd
// 注意 Register 将不会 return（如果没有 error 的话）
func Register(service string, addr string, stop chan error) error {
	// 使用默认配置创建一个 etcd client
	cli, err := clientv3.New(DefaultEtcdConfig)
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
			if err != nil {
				global.Log.Error(err.Error())
			}
			return err
		case <-cli.Ctx().Done(): // 监听服务被取消的新号
			global.Log.Info("服务关闭")
			return nil
		case _, ok := <-ch: // 监听租约撤销新号
			// 监听租约
			if !ok {
				global.Log.Info("心跳通道关闭")
				// 撤销撤销给定的租约
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
			global.Log.Infof("Recv reply from service: %s/%s, ttl:%d", service, addr, resp.TTL)
		}
	}
}

// const etcdUrl = "http://localhost:2379"
// const serviceName = "groupcache"
// const ttl = 10

// var etcdClient *clientv3.Client

// func etcdRegister(addr string) error {
// 	log.Printf("etcdRegister %s\b", addr)
// 	etcdClient, err := clientv3.NewFromURL(etcdUrl)

// 	if err != nil {
// 		return err
// 	}

// 	em, err := endpoints.NewManager(etcdClient, serviceName)
// 	if err != nil {
// 		return err
// 	}

// 	lease, _ := etcdClient.Grant(context.TODO(), ttl)

// 	err = em.AddEndpoint(context.TODO(), fmt.Sprintf("%s/%s", serviceName, addr), endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lease.ID))
// 	if err != nil {
// 		return err
// 	}
// 	//etcdClient.KeepAlive(context.TODO(), lease.ID)
// 	alive, err := etcdClient.KeepAlive(context.TODO(), lease.ID)
// 	if err != nil {
// 		return err
// 	}

// 	go func() {
// 		for {
// 			<-alive
// 			fmt.Println("etcd server keep alive")
// 		}
// 	}()

// 	return nil
// }

// func etcdUnRegister(addr string) error {
// 	log.Printf("etcdUnRegister %s\b", addr)
// 	if etcdClient != nil {
// 		em, err := endpoints.NewManager(etcdClient, serviceName)
// 		if err != nil {
// 			return err
// 		}
// 		err = em.DeleteEndpoint(context.TODO(), fmt.Sprintf("%s/%s", serviceName, addr))
// 		if err != nil {
// 			return err
// 		}
// 		return err
// 	}

// 	return nil
// }
