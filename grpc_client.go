package main

import (
	"context"
	"ggcache-plus/core"
	"ggcache-plus/distributekv/etcd"
	pb "ggcache-plus/distributekv/grpc/ggroupcachepb"
	"ggcache-plus/global"
	"math/rand"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	MaxRetries          = 3
	InitialRetryWaitSec = 1
	ErrRPCCallNotFound  = "rpc error: code = Unknown desc = record not found"
)

func main() {
	core.InitConf()
	global.Log = core.InitLogger()
	global.DB = core.InitGorm()
	core.InitEtcd()

	cli, err := clientv3.New(global.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	// 服务发现（直接根据服务名字获取与服务的虚拟端连接）
	conn, err := etcd.Discovery(cli, global.Config.GGroupCache.Name)
	if err != nil {
		panic(err)
	}
	client_stub := pb.NewGGroupCacheClient(conn)

	names := []string{"王五", "张三", "李四", "老二", "赵六"}
	for i := 0; i < 5; i++ {
		names = append(names, names...)
	}
	// 打散
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	for {
		for _, name := range names {
			searchFunc := func() (*pb.Response, error) {
				return client_stub.Get(context.Background(), &pb.Request{
					Group: "scores",
					Key:   name,
				})
			}

			for i := 0; i < MaxRetries; i++ { // 重试机制
				resp, err := searchFunc()
				if err != nil {
					if err.Error() == ErrRPCCallNotFound {
						global.Log.Warnf("查询不到学生 %s 的成绩", name)
						break
					}
					global.Log.Errorf("本次查询学生 %s 分数的 rpc 调用出现故障，重试次数 %d", name, i+1)
					waitTime := time.Duration(InitialRetryWaitSec*(1<<uint(i))) * time.Second // 退避算法
					time.Sleep(waitTime)
				} else {
					global.Log.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
					break
				}
			}
		}
	}
}
