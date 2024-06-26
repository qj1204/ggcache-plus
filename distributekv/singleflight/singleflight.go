package singleflight

import (
	"ggcache-plus/global"
	"os"
	"sync"
)

// call 代表正在进行中，或者已经结束的请求
type call struct {
	wg  sync.WaitGroup // 避免锁重入
	val any
	err error
}

// SingleFlight 主数据结构，管理不同 key 的请求(call)
type SingleFlight struct {
	mu sync.Mutex // 保护 Group 的成员变量 m 不被并发读写而加上的锁
	m  map[string]*call
}

// Do 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次
func (sf *SingleFlight) Do(key string, fn func() (any, error)) (any, error) {
	sf.mu.Lock()
	if sf.m == nil {
		sf.m = make(map[string]*call)
	}
	if c, ok := sf.m[key]; ok {
		// 直接可以释放锁了，让其他并发请求进来
		sf.mu.Unlock()
		// 等待查询 key 值的 goroutine 阻塞返回
		Geteuid := os.Geteuid()
		global.Log.Warnf("已经在查询了，阻塞等待 goroutine 返回, 进程号: %d\n", Geteuid)
		c.wg.Wait()
		// 用于查询的 goroutine 已经返回，结果值已经存入 Call 结构体中
		return c.val, c.err
	}
	c := new(call)
	sf.m[key] = c // 添加到 g.m，表明 key 已经有对应的请求在处理
	c.wg.Add(1)   // 发起请求前加锁
	sf.mu.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	sf.mu.Lock()
	delete(sf.m, key) // 更新 g.m
	sf.mu.Unlock()

	return c.val, c.err
}
