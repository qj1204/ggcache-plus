package distributekv

import (
	"context"
	"fmt"
	"ggcache-plus/distributekv/consistenthash"
	pb "ggcache-plus/distributekv/ggmemcachedpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"sync"
)

// grpcGetter grpc客户端结构体，实现了PeerGetter接口
type grpcGetter struct {
	addr string
}

// Get 通过grpc客户端获取远程节点数据
func (g *grpcGetter) Get(in *pb.Request, out *pb.Response) error {
	conn, err := grpc.Dial(g.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	// 创建一个grpc客户端
	client := pb.NewGroupCacheClient(conn)
	// 调用Get方法，获取远程节点数据
	response, err := client.Get(context.Background(), in)
	out.Value = response.Value
	return err
}

type GrpcPool struct {
	selfAddr    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	grpcGetters map[string]*grpcGetter
}

func NewGrpcPool(self string) *GrpcPool {
	return &GrpcPool{
		selfAddr: self,
	}
}

func (p *GrpcPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.grpcGetters = make(map[string]*grpcGetter, len(peers))
	for _, peer := range peers {
		p.grpcGetters[peer] = &grpcGetter{
			addr: peer,
		}
	}
}

// PickPeer 根据传入的 key 选择相应节点
func (p *GrpcPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.selfAddr {
		p.Log("Pick peer %s", peer)
		return p.grpcGetters[peer], true
	}
	return nil, false
}

func (p *GrpcPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.selfAddr, fmt.Sprintf(format, v...))
}

// Get 实现了GroupCacheServer接口
func (p *GrpcPool) Get(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	p.Log("%s %s", in.Group, in.Key)
	response := &pb.Response{}

	group := GetGroup(in.Group)
	if group == nil {
		p.Log("no such group %v", in.Group)
		return response, fmt.Errorf("no such group %v", in.Group)
	}
	value, err := group.Get(in.Key)
	if err != nil {
		p.Log("get key %v error %v", in.Key, err)
		return response, err
	}

	response.Value = value.ByteSlice()
	return response, nil
}

func (p *GrpcPool) Run() {
	lis, err := net.Listen("tcp", p.selfAddr)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	// 将p实例注册为一个grpc服务
	pb.RegisterGroupCacheServer(server, p)

	reflection.Register(server)
	err = server.Serve(lis)
	if err != nil {
		panic(err)
	}
}
