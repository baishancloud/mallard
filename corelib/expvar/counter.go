package expvar

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Counter is an interface for increment value counter
	Counter interface {
		Name() string
		Set(int64)
		Incr(int64)
		Count() int64
	}
	// Differ is an interface for difference value counter
	Differ interface {
		Counter
		Diff() int64
	}
	// QPSer is an interface for QPS value counter
	QPSer interface {
		Counter
		QPS() float64
	}
	// Averager is an interface for numbers mean counter
	Averager interface {
		Set(int64)
		Name() string
		Avg() float64
	}
)

var (
	_ Counter = (*BaseMeter)(nil)
	_ Counter = (*DiffMeter)(nil)
	_ Counter = (*QPSMeter)(nil)

	_ Differ = (*DiffMeter)(nil)
	_ Differ = (*QPSMeter)(nil)
	_ QPSer  = (*QPSMeter)(nil)

	_ Averager = (*AvgMeter)(nil)
)

// BaseMeter is main increment counter
// diff and qps counters are based on BaseMeter
type BaseMeter struct {
	value *int64
	name  string
	group string
}

// NewBase creates base counter with name
func NewBase(name string) *BaseMeter {
	var i int64
	return &BaseMeter{
		value: &i,
		name:  name,
	}
}

// Set sets a specific value
func (bc *BaseMeter) Set(v int64) {
	atomic.StoreInt64(bc.value, v)
}

// Incr increases counter by a given value
func (bc *BaseMeter) Incr(v int64) {
	atomic.AddInt64(bc.value, v)
}

// Count returns current value
func (bc *BaseMeter) Count() int64 {
	return atomic.LoadInt64(bc.value)
}

// Name returns counter's name
func (bc *BaseMeter) Name() string {
	return bc.name
}

// DiffMeter is values difference counter
type DiffMeter struct {
	BaseMeter
	lastValue    int64
	isFirstValue int
}

// NewDiff creates difference counter with name
func NewDiff(name string) *DiffMeter {
	var i int64
	return &DiffMeter{
		isFirstValue: 0,
		BaseMeter: BaseMeter{
			value: &i, // not be nil
			name:  name,
		},
	}
}

// Set sets diff value
func (dc *DiffMeter) Set(v int64) {
	if dc.isFirstValue <= 2 {
		dc.isFirstValue++
	}
	dc.BaseMeter.Set(v)
}

// Incr increases diff value
func (dc *DiffMeter) Incr(v int64) {
	if dc.isFirstValue <= 2 {
		dc.isFirstValue++
	}
	dc.BaseMeter.Incr(v)
}

// Diff return difference value from previous diff calling
func (dc *DiffMeter) Diff() int64 {
	value := dc.Count()
	diff := value - dc.lastValue
	dc.lastValue = value
	if dc.isFirstValue <= 1 {
		return 0
	}
	return diff
}

// QPSMeter is value time qps counter
type QPSMeter struct {
	DiffMeter
	lastDiff int64
	lastTime time.Time
	lastQPS  float64
}

// NewQPS creates qps counter with name
func NewQPS(name string) *QPSMeter {
	dc := NewDiff(name)
	return &QPSMeter{
		DiffMeter: *dc,
		lastTime:  time.Now(),
	}
}

// QPS return qps value
// if time delta < 10 seconds, return last qps value
func (qc *QPSMeter) QPS() float64 {
	delta := time.Since(qc.lastTime).Seconds()
	if delta < 10 {
		return qc.lastQPS
	}
	diff := qc.DiffMeter.Diff()
	qps := float64(diff) / float64(delta)
	qc.lastTime = time.Now()
	qc.lastQPS = qps
	qc.lastDiff = diff
	return qps
}

// Diff return difference value
// if time delta < 10 seconds, return last diff value
func (qc *QPSMeter) Diff() int64 {
	qc.QPS()
	return qc.lastDiff
}

// AvgMeter is average counter
type AvgMeter struct {
	name    string
	lock    sync.Mutex
	size    float64
	avg     float64
	isFirst bool
}

// NewAverage creates new average counter with calculation size and name
// size sets weighted refactor to generate mean value
func NewAverage(name string, size int) *AvgMeter {
	return &AvgMeter{
		size:    float64(size),
		isFirst: true,
		name:    name,
	}
}

// Name return counter's name
func (avg *AvgMeter) Name() string {
	return avg.name
}

// Set sets a specific value
// it generates average immediately
func (avg *AvgMeter) Set(v int64) {
	avg.lock.Lock()
	if avg.isFirst {
		avg.avg = float64(v)
		avg.isFirst = false
	} else {
		sum := avg.avg*(avg.size-1) + float64(v)
		avg.avg = sum / avg.size
	}
	avg.lock.Unlock()
}

// Avg return current average
func (avg *AvgMeter) Avg() float64 {
	avg.lock.Lock()
	defer avg.lock.Unlock()
	return avg.avg
}
