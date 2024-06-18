package policy

import (
	"time"
)

type Interface interface {
	Get(string) (Value, *time.Time, bool) // 根据 key 获取值
	Add(string, Value)                    // 添加键值对
	CleanUp(ttl time.Duration)            // 清理过期数据
	Value
}

type Value interface {
	Len() int // 值所占用的内存大小
}

// entry 键值对，双向链表节点的数据类型（这里多保存一个key的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射）
type entry struct {
	key      string
	value    Value
	updateAt *time.Time
}

// expired 判断是否过期
func (ele *entry) expired(duration time.Duration) (ok bool) {
	if ele.updateAt == nil {
		ok = false
	} else {
		ok = ele.updateAt.Add(duration).Before(time.Now())
	}
	return
}

// touch 更新数据访问时间
func (ele *entry) touch() {
	nowTime := time.Now()
	ele.updateAt = &nowTime
}

func New(name string, maxBytes int64, onEvicted func(string, Value)) Interface {
	switch name {
	case "fifo":
		return newFIFOCache(maxBytes, onEvicted)
	case "lru":
		return newLRUCache(maxBytes, onEvicted)
	case "lfu":
		return newLFUCache(maxBytes, onEvicted)
	}
	return nil
}
