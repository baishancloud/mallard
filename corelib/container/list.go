package container

import (
	"container/list"
	"sync"
)

// List is a list maked with container/list
type List struct {
	sync.RWMutex
	list *list.List
}

// NewList creates list
func NewList() *List {
	return &List{list: list.New()}
}

// Push pushes elememt at front
func (l *List) Push(v interface{}) {
	l.Lock()
	l.list.PushFront(v)
	l.Unlock()
}

// PushBatch pushes some elememts at front
func (l *List) PushBatch(vs []interface{}) {
	l.Lock()
	for _, item := range vs {
		l.list.PushFront(item)
	}
	l.Unlock()
}

// Pop pops element from back
func (l *List) Pop() interface{} {
	l.Lock()
	if elem := l.list.Back(); elem != nil {
		item := l.list.Remove(elem)
		l.Unlock()
		return item
	}
	l.Unlock()
	return nil
}

// PopBatch pops some elements from back
func (l *List) PopBatch(max int) []interface{} {
	l.Lock()
	count := l.list.Len()
	if count == 0 {
		l.Unlock()
		return []interface{}{}
	}
	if count > max {
		count = max
	}
	items := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		item := l.list.Remove(l.list.Back())
		items = append(items, item)
	}
	l.Unlock()
	return items
}

// Len return length of container/list
func (l *List) Len() int {
	l.RLock()
	defer l.RUnlock()
	return l.list.Len()
}

// LimitedList is list with size limited
type LimitedList struct {
	maxSize int
	list    *List
}

// NewLimitedList creates limited list with size
func NewLimitedList(maxSize int) *LimitedList {
	return &LimitedList{list: NewList(), maxSize: maxSize}
}

// Pop pops elememt from back
func (ll *LimitedList) Pop() interface{} {
	return ll.list.Pop()
}

// PopBatch pops some elements from back
func (ll *LimitedList) PopBatch(max int) []interface{} {
	return ll.list.PopBatch(max)
}

// Push pushes element at front
func (ll *LimitedList) Push(v interface{}) bool {
	if ll.list.Len() >= ll.maxSize {
		return false
	}
	ll.list.Push(v)
	return true
}

// PushBatch pusheds some elements at front
func (ll *LimitedList) PushBatch(vs []interface{}) bool {
	if ll.list.Len() >= ll.maxSize {
		return false
	}
	ll.list.PushBatch(vs)
	return true
}

// Len returns length of list
func (ll *LimitedList) Len() int {
	return ll.list.Len()
}
