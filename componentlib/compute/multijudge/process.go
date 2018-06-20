package multijudge

import (
	"sync"
	"sync/atomic"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	stopFlag int64
	stopWg   sync.WaitGroup
)

var (
	judgeFn []func([]*models.Metric)
)

func RegisterFn(fns ...func([]*models.Metric)) {
	judgeFn = append(judgeFn, fns...)
}

// Process processes metrics queue
func Process(queue <-chan []*models.Metric) {
	for {
		if atomic.LoadInt64(&stopFlag) > 0 {
			return
		}
		metrics := <-queue
		if len(metrics) > 0 && len(judgeFn) > 0 {
			stopWg.Add(1)
			for _, fn := range judgeFn {
				if fn != nil {
					fn(metrics)
				}
			}
			stopWg.Done()
		}
	}
}

func Stop() {
	atomic.StoreInt64(&stopFlag, 1)
	stopWg.Wait()
}
