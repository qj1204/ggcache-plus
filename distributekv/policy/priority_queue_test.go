package policy

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := priorityQueue([]*Element{})
	heap.Push(&pq, &Element{0, 2, &entry{}})
	heap.Push(&pq, &Element{0, 4, &entry{}})
	heap.Push(&pq, &Element{0, 1, &entry{}})
	heap.Push(&pq, &Element{0, 5, &entry{}})
	heap.Push(&pq, &Element{0, 3, &entry{}})
	for pq.Len() != 0 {
		e := heap.Pop(&pq).(*Element)
		fmt.Println(e.count)
	}
}
