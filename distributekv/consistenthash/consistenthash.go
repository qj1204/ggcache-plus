package consistenthash

import (
	"ggcache-plus/global"
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	hash     Hash           // 哈希函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环虚拟节点
	hashMap  map[int]string // 虚拟节点与真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称
}

// New creates a Map instance
func New(replicas int, hash Hash) *Map {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &Map{
		replicas: replicas,
		hash:     hash,
		hashMap:  map[int]string{},
	}
}

// Add 往哈希环中添加节点（实际上添加的是虚拟节点）
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(key + strconv.Itoa(i))))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// GetTruthNode 获取哈希环中最接近提供的键的节点
func (m *Map) GetTruthNode(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 查找第一个匹配的虚拟节点的下标，如果没有找到就返回 len(m.keys)
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	global.Log.Infof("计算出 key:%s 的 hash: %d, 顺时针选择的虚拟节点下标 idx: %d", key, hash, idx)
	global.Log.Infof("选择的真实节点：%s", m.hashMap[m.keys[idx%len(m.keys)]])

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
