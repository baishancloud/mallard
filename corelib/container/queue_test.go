package container

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var testValue = []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func TestList(t *testing.T) {
	Convey("List", t, func() {
		l := NewList()
		l.PushBatch(testValue)
		l.Push(testValue[0])
		So(l.Pop().(int), ShouldEqual, 1)
		values := l.PopBatch(3)
		So(values[0].(int), ShouldEqual, 2)
		So(values[1].(int), ShouldEqual, 3)
		So(values[2].(int), ShouldEqual, 4)
		So(l.Len(), ShouldEqual, 7)
	})

	Convey("LimitList", t, func() {
		l := NewLimitedList(10)
		l.PushBatch(testValue)

		miss := l.Push(testValue[0])
		So(miss, ShouldBeFalse)

		miss = l.PushBatch(testValue)
		So(miss, ShouldBeFalse)

		So(l.Pop().(int), ShouldEqual, 1)
		values := l.PopBatch(3)
		So(values[0].(int), ShouldEqual, 2)
		So(values[1].(int), ShouldEqual, 3)
		So(values[2].(int), ShouldEqual, 4)
		So(l.Len(), ShouldEqual, 6)

		miss = l.Push(testValue[0])
		So(miss, ShouldBeTrue)
	})
}

func BenchmarkList(b *testing.B) {
	l := NewList()
	for i := 0; i < b.N; i++ {
		l.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if l.Pop() == nil {
			break
		}
	}
}

func BenchmarkListBatch(b *testing.B) {
	l := NewList()
	for i := 0; i < b.N; i++ {
		l.PushBatch(testValue)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(l.PopBatch(5)) == 0 {
			break
		}
	}
}
