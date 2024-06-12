package hnsw

import (
	"container/heap"
	"fmt"
)

type Comparator interface {
	Less(i, j *Item) bool
}

// MaxComparator implements the Comparator interface for a max-heap.
type MaxComparator struct{}

func (c MaxComparator) Less(i, j *Item) bool {
	return i.dist > j.dist
}

// MinComparator implements the Comparator interface for a min-heap.
type MinComparator struct{}

func (c MinComparator) Less(i, j *Item) bool {
	return i.dist < j.dist
}

type Item struct {
	id    Id
	dist  float32
	index int
}

type Heapy interface {
	heap.Interface
	Insert(id Id, dist float32)
	IsEmpty() bool
	Len() int
	PopItem() *Item
	Top() *Item
	Take(count int) (*BaseQueue, error)
	update(item *Item, id Id, dist float32)
}

// Nothing from BaseQueue should be used. Only use the Max and Min queue.
// BaseQueue isn't even a heap! It misses the Less() method which the Min/Max queue implement.
type BaseQueue struct {
	visitedIds map[Id]*Item
	items      []*Item
	comparator Comparator
}

func (bq *BaseQueue) Take(count int, comparator Comparator) (*BaseQueue, error) {
	if len(bq.items) < count {
		return nil, fmt.Errorf("queue only has %v items, but want to take %v", len(bq.items), count)
	}

	pq := NewBaseQueue(comparator)

	ct := 0
	for {
		if ct == count {
			break
		}

		peeled, err := bq.PopItem()
		if err != nil {
			return nil, err
		}

		pq.Insert(peeled.id, peeled.dist)

		ct++
	}

	return pq, nil
}

func (bq BaseQueue) Len() int { return len(bq.items) }
func (bq BaseQueue) Swap(i, j int) {
	pq := bq.items
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (bq *BaseQueue) Push(x any) {
	n := len(bq.items)
	item := x.(*Item)
	item.index = n
	bq.items = append(bq.items, item)
}

func (bq *BaseQueue) Top() *Item {
	if len(bq.items) == 0 {
		return nil
	}
	return bq.items[0]
}

func (bq *BaseQueue) IsEmpty() bool {
	return len(bq.items) == 0
}

func (bq *BaseQueue) Pop() any {
	old := bq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	bq.items = old[0 : n-1]
	return item
}

func (bq *BaseQueue) Less(i, j int) bool {
	return bq.comparator.Less(bq.items[i], bq.items[j])
}

func (bq *BaseQueue) Insert(id Id, dist float32) {
	if item, ok := bq.visitedIds[id]; ok {
		bq.update(item, id, dist)
		return
	}

	newItem := Item{id: id, dist: dist}
	heap.Push(bq, &newItem)
	bq.visitedIds[id] = &newItem

}

func NewBaseQueue(comparator Comparator) *BaseQueue {
	bq := &BaseQueue{
		visitedIds: map[Id]*Item{},
		comparator: comparator,
	}
	heap.Init(bq)
	return bq
}

func (bq *BaseQueue) PopItem() (*Item, error) {
	if bq.Len() == 0 {
		return nil, fmt.Errorf("no items to peel")
	}
	popped := heap.Pop(bq).(*Item)
	delete(bq.visitedIds, popped.id)
	return popped, nil
}

func (bq *BaseQueue) update(item *Item, id Id, dist float32) {
	item.id = id
	item.dist = dist
	heap.Fix(bq, item.index)
}

func FromBaseQueue(bq *BaseQueue, comparator Comparator) *BaseQueue {
	newBq := NewBaseQueue(comparator)

	for _, item := range bq.items {
		newBq.Insert(item.id, item.dist)
	}

	return newBq
}

func FromItems(items []*Item, comparator Comparator) *BaseQueue {
	bq := &BaseQueue{
		visitedIds: map[Id]*Item{},
		items:      items,
		comparator: comparator,
	}

	heap.Init(bq)

	return bq
}