package lru

import "container/list"

type Value interface {
	Len() int // 值所占用的内存大小
}

// entry 键值对，双向链表节点的数据类型（这里多保存一个key的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射）
type entry struct {
	key   string
	value Value
}

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes  int64                         // 最大内存
	nBytes    int64                         // 当前已使用内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // key 是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Add(key string, value Value) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element) // 将该节点移动到队尾
		kv := element.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		element = c.ll.PushFront(&entry{key, value}) // 在队尾添加新节点
		c.cache[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	element := c.ll.Back() // 队首节点
	if element != nil {
		c.ll.Remove(element)
		kv := element.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element) // 双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
