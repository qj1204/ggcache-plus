package main

import (
	"context"
	"fmt"
	pb "ggcache-plus/distributekv/grpc/ggmemcachedpb"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const ErrRPCCallNotFound = "rpc error: code = Unknown desc = record not found"

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := cli.Get(ctx, "clusters", clientv3.WithPrefix())
	if err != nil {
		fmt.Println("从 etcd 获取 grpc 通道失败")
		return
	}
	fmt.Println("从 etcd 获取 grpc 通道成功")

	addr := string(resp.Kvs[0].Value)
	fmt.Println("从 etcd 获取的地址：", addr) // localhost:10001
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Println("获取 grpc 通道失败")
		return
	}

	client_stub := pb.NewGroupCacheClient(conn)
	names := []string{"张三"}
	for _, name := range names {
		response, err := client_stub.Get(ctx, &pb.Request{Group: "scores", Key: name})
		if err != nil {
			if err.Error() == ErrRPCCallNotFound {
				fmt.Printf("没有查询学生 %s 的分数，err: %s\n", name, err.Error())
				continue
			} else {
				panic(err)
			}
		}
		fmt.Printf("成功从 RPC 返回学生 %s 分数：%s\n", name, string(response.GetValue()))
	}
}
