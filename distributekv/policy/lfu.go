package policy

import (
	"container/heap"
	"time"
)

type LFUCache struct {
	maxBytes  int64                         // 最大内存
	nBytes    int64                         // 当前已使用内存
	pq        *priorityQueue                // 优先队列
	cache     map[string]*Element           // key 是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

func newLFUCache(maxBytes int64, onEvicted func(string, Value)) *LFUCache {
	queue := priorityQueue(make([]*Element, 0))
	return &LFUCache{
		maxBytes:  maxBytes,
		pq:        &queue,
		cache:     make(map[string]*Element),
		OnEvicted: onEvicted,
	}
}

func (p *LFUCache) Add(key string, value Value) {
	if element, ok := p.cache[key]; ok {
		heap.Fix(p.pq, element.index) // 更新优先队列
		kv := element.Value.(*entry)  // 获取节点的值
		element.referenced()          // 更新引用次数和时间
		p.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value // 更新值
	} else {
		element = &Element{0, 0, &entry{key, value, nil}}
		element.referenced()
		heap.Push(p.pq, element)
		p.cache[key] = element
		p.nBytes += int64(len(key)) + int64(value.Len())
	}
	for p.maxBytes != 0 && p.maxBytes < p.nBytes {
		p.Remove()
	}
}

func (p *LFUCache) Remove() {
	element := heap.Pop(p.pq).(*Element)
	if element != nil {
		kv := element.Value.(*entry)
		delete(p.cache, kv.key)
		p.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if p.OnEvicted != nil {
			p.OnEvicted(kv.key, kv.value)
		}
	}
}

func (p *LFUCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if element, ok := p.cache[key]; ok {
		heap.Fix(p.pq, element.index)
		kv := element.Value.(*entry)
		element.referenced()
		return kv.value, kv.updateAt, ok
	}
	return
}

func (p *LFUCache) CleanUp(ttl time.Duration) {
	for _, element := range *p.pq {
		if element.Value.(*entry).expired(ttl) {
			kv := heap.Remove(p.pq, element.index).(*Element).Value.(*entry)
			delete(p.cache, kv.key)
			p.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if p.OnEvicted != nil {
				p.OnEvicted(kv.key, kv.value)
			}
		}
	}
}

func (p *LFUCache) Len() int {
	return p.pq.Len()
}
