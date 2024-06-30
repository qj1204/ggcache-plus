package distributekv

import (
	"ggcache-plus/distributekv/policy"
	"ggcache-plus/global"
	"sync"
)

type cache struct {
	mu         sync.RWMutex
	strategy   policy.CacheInterface
	cacheBytes int64
}

func newCache(strategy string, cacheBytes int64) *cache {
	onEvicted := func(key string, val policy.Value) {
		global.Log.Infof("缓存条目 [%s:%s] 被淘汰", key, val)
	}
	return &cache{
		cacheBytes: cacheBytes,
		strategy:   policy.New(strategy, cacheBytes, onEvicted),
	}
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 判断了 c.strategy 是否为 nil，如果等于 nil 再创建实例。
	// 这种方法称之为延迟初始化(Lazy Initialization)，一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。
	// 主要用于提高性能，并减少程序内存要求
	if c.strategy == nil {
		c.strategy = policy.New("lru", 2<<10, func(key string, val policy.Value) {
			global.Log.Infof("缓存 [%s:%s] 被淘汰", key, val)
		})
	}
	global.Log.Infof("添加缓存(%s, %s)", key, value)
	c.strategy.Add(key, value)
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, _, ok := c.strategy.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}
