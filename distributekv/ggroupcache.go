package distributekv

import (
	"errors"
	"ggcache-plus/distributekv/singleflight"
	"ggcache-plus/global"
	"gorm.io/gorm"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// RetrieverFunc 实现 Retriever 接口
type RetrieverFunc func(key string) ([]byte, error)

// Retrieve 函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。
func (f RetrieverFunc) Retrieve(key string) ([]byte, error) {
	return f(key)
}

// Group 控制中心，负责与用户交互，控制缓存存储和获取的流程
type Group struct {
	name      string                     // 缓存的命名空间
	retriever Retriever                  // 缓存未命中时获取源数据的回调
	mainCache *cache                     // 并发缓存
	peers     PeerPicker                 // 选择远程节点
	loader    *singleflight.SingleFlight // use singleflight.SingleFlight to make sure that each key is only fetched once
}

// NewGroup 创建一个缓存命名空间
func NewGroup(name string, strategy string, cacheBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Retriever 不能为空！")
	}
	g := &Group{
		name:      name,
		retriever: retriever,
		mainCache: newCache(strategy, cacheBytes),
		loader:    &singleflight.SingleFlight{},
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g
	return g
}

// RegisterPeers 注册节点选择器到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("节点选择器已经注册！")
	}
	g.peers = peers
}

// GetGroup 根据命名空间获取 Group 对象（对实际缓存进行管理）
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

// Get 从缓存中查找缓存数据，如果不存在则调用 load 方法获取
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key 必须存在！")
	}
	global.Log.Infof("尝试从缓存中获取数据，key: %s...", key)
	if v, ok := g.mainCache.get(key); ok {
		global.Log.Infof("key：%s，缓存命中...", key)
		return v, nil
	}
	return g.load(key)
}

// load 先从远端获取数据，没有的话再调用 getLocally 从本地获取数据
func (g *Group) load(key string) (ByteView, error) {
	// 无论并发调用者的数量如何，每个键只被获取一次(本地或远程)
	val, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.Pick(key); ok {
				bytes, err := peer.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				global.Log.Error(err.Error())
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return val.(ByteView), nil
	}
	return ByteView{}, err
}

// getLocally 调用用户回调函数 g.retriever.Retrieve() 获取源数据，并且将源数据添加到缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.Retrieve(key) // 实际上就是执行RetrieverFunc函数
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Log.Warnf("对于不存在的 key，为了防止缓存穿透，先存入缓存中并设置合理过期时间")
			g.mainCache.add(key, ByteView{})
		}
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 将从底层数据库中查询到的数据填充到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
