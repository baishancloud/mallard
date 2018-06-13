package container

// Queue is queue interface
type Queue interface {
	Push(interface{})
	PushBatch([]interface{})
	Pop() interface{}
	PopBatch(max int) []interface{}
	Len() int
}

var (
	_ Queue = new(List)
)

// LimitedQueue is interface for limited size queue
type LimitedQueue interface {
	Push(v interface{}) bool
	PushBatch([]interface{}) bool
	Pop() interface{}
	PopBatch(max int) []interface{}
	Len() int
}

var (
	_ LimitedQueue = new(LimitedList)
)
