package policy

import (
	"testing"
)

func TestLFUCache_Get(t *testing.T) {
	lfu := newLFUCache(10, nil)
	lfu.Add("key1", String("1234"))
	t.Log(lfu.Get("key1"))
	if v, _, ok := lfu.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestLFUCache_Remove(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	curCap := len(k1 + k2 + v1 + v2)
	lfu := newLFUCache(int64(curCap), nil)
	lfu.Add(k1, String(v1))
	lfu.Add(k1, String(v1))
	lfu.Add(k2, String(v2))
	lfu.Add(k3, String(v3))

	//for k, v := range lfu.cache {
	//	fmt.Printf("%s%v\n", k, v)
	//}
	if _, _, ok := lfu.Get("key2"); ok || lfu.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
