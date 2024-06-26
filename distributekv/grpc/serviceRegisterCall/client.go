package main

import (
	"context"
	"fmt"
	pb "ggcache-plus/distributekv/grpc/ggmemcachedpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc/balancer/roundrobin"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	etcdUrl            = "localhost:2379"
	serviceName        = "ggmemcached"
	ErrRPCCallNotFound = "rpc error: code = Unknown desc = record not found"
)

func main() {
	etcdClient, err := clientv3.NewFromURL(etcdUrl)
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(
		"etcd:///"+serviceName,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
	)

	if err != nil {
		fmt.Printf("conn err: %v", err)
		return
	}

	ServerClient := pb.NewGroupCacheClient(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		resp, err := ServerClient.Get(ctx, &pb.Request{
			Group: "scores",
			Key:   "李四",
		})
		if err != nil {
			fmt.Printf("err: %s", err.Error())
			return
		}
		fmt.Printf("查询到 %s 的分数为：%v\n", "李四", string(resp.Value))

		resp, err = ServerClient.Get(ctx, &pb.Request{
			Group: "scores",
			Key:   "王五",
		})
		if err != nil {
			fmt.Printf("err: %s", err.Error())
			return
		}
		fmt.Printf("查询到 %s 的分数为：%s\n", "王五", string(resp.Value))

		resp, err = ServerClient.Get(ctx, &pb.Request{
			Group: "scores",
			Key:   "xiaoxin",
		})
		if err != nil {
			if err.Error() == ErrRPCCallNotFound {
				fmt.Printf("没有查询到学生 %s 的成绩\n", "xiaoxin")
				continue
			} else {
				fmt.Printf("调用失败：%s\n", err.Error())
				return
			}
		}
		fmt.Printf("调用成功, 学生 %s 的成绩为 %s\n", "xiaoxin", string(resp.Value))
	}
}
