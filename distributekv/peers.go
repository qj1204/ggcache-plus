package distributekv

import pb "ggcache-plus/distributekv/ggmemcachedpb"

// PeerPicker 根据传入的 key 选择相应节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 通过grpc客户端获取远程节点数据
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error //
}
