package hnsw

import "container/heap"

type Item struct {
	id    NodeId
	dist  float64
	index int
}

type Heapy interface {
	heap.Interface
	Insert(id NodeId, dist float64)
	IsEmpty() bool
	Len() int
	Peel() *Item
	Peek() *Item

	update(item *Item, id NodeId, dist float64)
}

// Nothing from baseQueue should be used. Only use the Max and Min queue.
// baseQueue isn't even a heap! It misses the Less() method which the Min/Max queue implement.
type baseQueue struct{ items []*Item }

func (bq baseQueue) Len() int { return len(bq.items) }
func (bq baseQueue) Swap(i, j int) {
	pq := bq.items
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (bq *baseQueue) Push(x any) {
	n := len(bq.items)
	item := x.(*Item)
	item.index = n
	bq.items = append(bq.items, item)
}

func (bq *baseQueue) Peek() *Item {
	if len(bq.items) == 0 {
		return nil
	}
	return bq.items[0]
}

func (bq *baseQueue) IsEmpty() bool {
	return len(bq.items) == 0
}

func (bq *baseQueue) Pop() any {
	old := bq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	bq.items = old[0 : n-1]
	return item
}

type MinQueue struct{ baseQueue }

type MaxQueue struct{ baseQueue }

func NewMinQueue() *MinQueue {
	mq := &MinQueue{}
	heap.Init(mq)
	return mq
}

func NewMaxQueue() *MaxQueue {
	mq := &MaxQueue{}
	heap.Init(mq)
	return mq
}

func (mq *MinQueue) Insert(id NodeId, dist float64) {
	heap.Push(mq, &Item{id: id, dist: dist})
}

func (mq *MaxQueue) Insert(id NodeId, dist float64) {
	heap.Push(mq, &Item{id: id, dist: dist})
}

func (mq *MinQueue) Less(i, j int) bool {
	return mq.items[i].dist < mq.items[j].dist
}

func (mq *MaxQueue) Less(i, j int) bool {
	return mq.items[i].dist > mq.items[j].dist
}

func (mq *MinQueue) Peel() *Item {
	if mq.Len() == 0 {
		return nil
	}
	return heap.Pop(mq).(*Item)
}

func (mq *MaxQueue) Peel() *Item {
	if mq.Len() == 0 {
		return nil
	}
	return heap.Pop(mq).(*Item)
}

func (mq *MinQueue) update(item *Item, id NodeId, dist float64) {
	item.id = id
	item.dist = dist
	heap.Fix(mq, item.index)
}
func (mq *MaxQueue) update(item *Item, id NodeId, dist float64) {
	item.id = id
	item.dist = dist
	heap.Fix(mq, item.index)
}