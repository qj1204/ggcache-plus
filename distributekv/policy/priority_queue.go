package policy

// priorityQueue 使用数组实现的优先队列
type priorityQueue []*Element

// Element 优先队列的元素类型
type Element struct {
	index int // 元素在数组中的索引
	count int // 访问次数
	Value any
}

// Referenced 更新访问次数
func (l *Element) referenced() {
	l.count++
	l.Value.(*entry).touch()
}

// Less 比较两个元素的大小(默认i < j)
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].count == pq[j].count {
		// 访问次数相同时，比较访问时间
		return pq[i].Value.(*entry).updateAt.Before(*pq[j].Value.(*entry).updateAt)
	}
	// 访问次数不同时，比较访问次数
	return pq[i].count < pq[j].count
}

// Len 返回队列长度
func (pq priorityQueue) Len() int {
	return len(pq)
}

// Swap 交换两个元素
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Pop 弹出队尾元素
func (pq *priorityQueue) Pop() any {
	oldPQ := *pq
	n := len(oldPQ)
	e := oldPQ[n-1]
	oldPQ[n-1] = nil //避免内存泄露
	newPQ := oldPQ[0 : n-1]
	*pq = newPQ
	return e
}

// Push 插入新元素
func (pq *priorityQueue) Push(x any) {
	e := x.(*Element)
	e.index = len(*pq)
	*pq = append(*pq, e)
}
