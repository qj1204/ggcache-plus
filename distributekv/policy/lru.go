package policy

import (
	"container/list"
	"time"
)

// LRUCache is an LRU cache. It is not safe for concurrent access.
type LRUCache struct {
	maxBytes  int64                         // 最大内存
	nBytes    int64                         // 当前已使用内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // key 是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

func newLRUCache(maxBytes int64, onEvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
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
		c.ll.MoveToBack(element) // 双向链表作为队列，队首队尾是相对的，在这里约定 back 为队尾
		kv := element.Value.(*entry)
		kv.touch()
		return kv.value, kv.updateAt, true
	}
	return
}

func (c *LRUCache) CleanUp(ttl time.Duration) {
	for element := c.ll.Front(); element != nil; element = element.Next() {
		if element.Value.(*entry).expired(ttl) {
			kv := c.ll.Remove(element).(*entry)
			delete(c.cache, kv.key)
			c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		} else {
			break
		}
	}
}

// Len the number of cache entries
func (c *LRUCache) Len() int {
	return c.ll.Len()
}
