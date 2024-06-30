package policy

import (
	"container/list"
	"ggcache-plus/global"
	"sync"
	"time"
)

// LRUCache is an LRU cache. It is not safe for concurrent access.
type LRUCache struct {
	maxBytes  int64      // 最大内存
	nBytes    int64      // 当前已使用内存
	ll        *list.List // 双向链表
	mu        sync.RWMutex
	cache     map[string]*list.Element      // key 是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

func newLRUCache(maxBytes int64, onEvicted func(string, Value)) *LRUCache {
	l := &LRUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
	ttl := time.Duration(global.Config.GGroupCache.TTL) * time.Second

	go func() {
		ticker := time.NewTicker(time.Duration(global.Config.GGroupCache.CleanUpInterval) * time.Minute) // 每一分钟清理一次过期缓存
		defer ticker.Stop()
		for {
			<-ticker.C
			l.CleanUp(ttl)
			global.Log.Warn("触发过期缓存清理后台任务...")
		}
	}()
	return l
}

func (c *LRUCache) Add(key string, value Value) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToBack(element)     // 将该节点移动到队尾
		kv := element.Value.(*entry) // 获取节点的值
		kv.touch()                   // 更新时间
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value // 更新值
	} else {
		kv := &entry{key, value, nil}
		kv.touch()                  // 更新时间
		element = c.ll.PushBack(kv) // 在队尾添加新节点
		c.cache[key] = element      // 更新字典
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *LRUCache) RemoveOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	element := c.ll.Front() // 队首节点
	if element != nil {
		kv := c.ll.Remove(element).(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if element, ok := c.cache[key]; ok {
		if ent := element.Value.(*entry); ent != nil {
			if !ent.expired(time.Duration(global.Config.GGroupCache.TTL) * time.Second) {
				c.ll.MoveToBack(element) // 双向链表作为队列，队首队尾是相对的，在这里约定 back 为队尾
				kv := element.Value.(*entry)
				kv.touch()
				return kv.value, kv.updateAt, true
			}
		}
	}
	return
}

func (c *LRUCache) CleanUp(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for e := c.ll.Front(); e != nil; {
		next := e.Next()
		if ent, ok := e.Value.(*entry); ok && ent != nil {
			if ent.expired(ttl) {
				c.ll.Remove(e)
				delete(c.cache, ent.key)
				c.nBytes -= int64(len(ent.key)) + int64(ent.value.Len())
				if c.OnEvicted != nil {
					c.OnEvicted(ent.key, ent.value)
				}
			}
		}
		e = next
	}
}

// Len the number of cache entries
func (c *LRUCache) Len() int {
	return c.ll.Len()
}
