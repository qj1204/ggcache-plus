package distributekv

// PeerPicker 根据传入的 key 选择相应节点
type PeerPicker interface {
	Pick(key string) (peer Fetcher, ok bool)
}

// Fetcher 节点（grpc客户端）获取远程节点数据
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error) //
}

// Retriever 从数据源获取数据，并且将获取的数据添加到缓存中
type Retriever interface {
	Retrieve(key string) ([]byte, error)
}
