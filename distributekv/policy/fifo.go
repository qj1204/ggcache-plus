package policy

import (
	"container/list"
	"time"
)

type FIFOCache struct {
	maxBytes  int64                         // 最大内存
	nBytes    int64                         // 当前已使用内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // key 是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

func newFIFOCache(maxBytes int64, onEvicted func(string, Value)) *FIFOCache {
	return &FIFOCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (f *FIFOCache) Add(key string, value Value) {
	if element, ok := f.cache[key]; ok {
		kv := element.Value.(*entry) // 获取节点的值
		f.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		kv := &entry{key, value, nil}
		kv.touch()
		element = f.ll.PushBack(kv) // 在队尾添加新节点
		f.cache[key] = element
		f.nBytes += int64(len(kv.key)) + int64(kv.value.Len())
	}

	for f.maxBytes != 0 && f.maxBytes < f.nBytes {
		f.RemoveFront()
	}
}

func (f *FIFOCache) RemoveFront() {
	element := f.ll.Front() // 队首节点
	if element != nil {
		kv := f.ll.Remove(element).(*entry)
		delete(f.cache, kv.key)
		f.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if f.OnEvicted != nil {
			f.OnEvicted(kv.key, kv.value)
		}
	}
}

func (f *FIFOCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if element, ok := f.cache[key]; ok {
		kv := element.Value.(*entry)
		return kv.value, kv.updateAt, true
	}
	return
}

func (f *FIFOCache) CleanUp(ttl time.Duration) {
	for element := f.ll.Front(); element != nil; element = element.Next() {
		if element.Value.(*entry).expired(ttl) {
			kv := f.ll.Remove(element).(*entry)
			delete(f.cache, kv.key)
			f.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if f.OnEvicted != nil {
				f.OnEvicted(kv.key, kv.value)
			}
		} else {
			break
		}
	}
}

func (f *FIFOCache) Len() int {
	return f.ll.Len()
}
