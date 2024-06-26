package distributekv

import (
	"context"
	"errors"
	"fmt"
	"ggcache-plus/distributekv/consistenthash"
	"ggcache-plus/distributekv/etcd"
	pb "ggcache-plus/distributekv/grpc/ggmemcachedpb"
	"ggcache-plus/global"
	"ggcache-plus/utils"
	"google.golang.org/grpc"
	"net"
	"strings"
	"sync"
)

// server 模块为 ggmemcached 之间提供了通信能力
// 这样部署在其他机器上的 ggmemcached 可以通过 server 获取缓存
// 至于找哪一个主机，由一致性 hash 负责
const (
	defaultAddr     = "localhost:10001"
	defaultReplicas = 50
)

// Server 服务端，实现了 GroupCacheServer 接口
type Server struct {
	addr        string     // format: ip:port
	status      bool       // true: running false: stop
	stopsSignal chan error // 通知 etcd revoke 服务
	mu          sync.Mutex
	consHash    *consistenthash.Map
	grpcClients map[string]*grpcClient
}

func NewServer(addr string) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}
	if !utils.IsValidPerrAddr(addr) {
		return nil, errors.New(fmt.Sprintf("无效地址 %s，格式应为：ip:port", addr))
	}
	return &Server{addr: addr}, nil
}

// Get 实现了GroupCacheServer接口的Get方法
func (s *Server) Get(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	group, key := req.GetGroup(), req.GetKey()
	global.Log.Infof("[%s] Recv RPC Request - (%s)/(%s)", s.addr, group, key)

	resp := &pb.Response{}
	if key == "" || group == "" {
		return resp, errors.New("group和key不能为空！")
	}

	g := GetGroup(group)
	global.Log.Info("group的名字为：", group)
	if g == nil {
		global.Log.Errorf("没有名为 %s 的group", group)
		return resp, errors.New(fmt.Sprintf("group %s not found", group))
	}
	value, err := g.Get(key)
	if err != nil {
		global.Log.Errorf("未能获取到key %s 的值，err：%s", key, err.Error())
		return resp, err
	}
	resp.Value = value.ByteSlice()
	return resp, nil
}

// SetPeers 将各个远端主机 IP 配置到 Server 里
func (s *Server) SetPeers(peerAddrs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consHash = consistenthash.New(defaultReplicas, nil)
	s.consHash.Add(peerAddrs...)
	s.grpcClients = make(map[string]*grpcClient)

	for _, peerAddr := range peerAddrs {
		if !utils.IsValidPerrAddr(peerAddr) {
			panic(fmt.Sprintf("peer %s 的地址格式无效，应为 ip:port", peerAddr))
		}
		s.grpcClients[peerAddr] = NewClient(fmt.Sprintf("ggmemcached/%s", peerAddr))
	}
}

// Pick 根据一致性哈希选出 key 应该存放在的 cache
func (s *Server) Pick(key string) (Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	peerAddr := s.consHash.GetTruthNode(key)
	if peerAddr == s.addr {
		global.Log.Infof("选到我自己，我的地址是：[%s]", s.addr)
		return nil, false
	}
	global.Log.Infof("节点 [%s] 选择的远程节点地址为：[%s]", s.addr, peerAddr)
	return s.grpcClients[peerAddr], true
}

// Start 启动 Cache 服务
func (s *Server) Start() error {
	s.mu.Lock()

	if s.status {
		s.mu.Unlock()
		return errors.New(fmt.Sprintf("服务 %s 已经启动", s.addr))
	}

	// ------------启动服务----------------
	// 1. 设置 status = true 表示服务器已经在运行
	// 2. 初始化 stop channel，用于通知 registry stop keepalive
	// 3. 初始化 tcp socket 并开始监听
	// 4. 注册 rpc 服务至 grpc，这样 grpc 收到 request 可以分发给 server 处理
	// 5. 将自己的服务名/Host地址注册至 etcd，这样 client 就可以通过 etcd 获取服务 Host 地址进行通信；这样做的好处是：client 只需要知道服务名称以及 etcd 的 Host 就可以获取
	// 指定服务的 IP，无需将它们写死在 client 代码中
	s.status = true
	s.stopsSignal = make(chan error)
	port := strings.Split(s.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return errors.New(fmt.Sprintf("未能监听 %s，错误：%v", s.addr, err))
	}
	grpcServer := grpc.NewServer()
	pb.RegisterGroupCacheServer(grpcServer, s)

	// 注册服务至 etcd
	go func() {
		// Register never return unless stop signal received (blocked)
		err := etcd.Register("ggmemcached", s.addr, s.stopsSignal)
		if err != nil {
			global.Log.Error("注册错误：", err.Error())
		}
		// close channel
		close(s.stopsSignal)
		// close tcp listen
		err = lis.Close()
		if err != nil {
			global.Log.Error(err.Error())
		}
		global.Log.Infof("[%s] Revoke service and close tcp socket ok.", s.addr)
	}()

	s.mu.Unlock()
	// serve接受侦听器列表上的传入连接，为每个连接创建一个新的ServerTransport和服务goroutine。
	// 服务goroutines读取grpc请求，然后调用注册的处理程序来回复它们。当lis.Accept失败并出现致命错误时，Serve返回。当此方法返回时，LIS将关闭。
	// 除非调用Stop或GracefulStop，否则SERVE将返回非零错误。
	if err := grpcServer.Serve(lis); s.status && err != nil {
		return errors.New(fmt.Sprintf("failed to serve %s, error: %v", s.addr, err))
	}
	return nil
}
