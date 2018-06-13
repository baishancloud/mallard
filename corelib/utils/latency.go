package utils

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	// InitValue is init value of latency
	InitValue int64 = -100
	// FailValue is fail value of latency
	FailValue int64 = -1
)

var (
	// ErrorOversize mean read from latency but index is out of range
	ErrorOversize = errors.New("over size")
)

// Latency is latency history recorder, use to get min latency index
// It do not use lock to ensure latency data are thread-safe. (it seems no need to make them safe)
type Latency struct {
	history   []int64
	size      int
	resetSize int
	setCount  int
}

// NewLatency creates new lantecy object with size and resetcount size
func NewLatency(size int, resetSize int) *Latency {
	h := &Latency{
		history:   make([]int64, size),
		size:      size,
		resetSize: resetSize,
	}
	for i := range h.history {
		h.history[i] = InitValue
	}
	return h
}

// Reset reset latency history to no values
func (l *Latency) Reset() {
	for i := range l.history {
		l.history[i] = InitValue
	}
	l.setCount = 0
}

// Set sets latency value to index one
func (l *Latency) Set(idx int, latency int64) {
	if idx+1 > len(l.history) {
		return
	}
	l.history[idx] = latency
	l.setCount++
	l.shouldReset()
}

func (l *Latency) shouldReset() {
	if l.resetSize > 0 && l.setCount > l.resetSize {
		l.Reset()
	}
}

// SetFail sets latency fail value to index one
func (l *Latency) SetFail(idx int) {
	l.Set(idx, FailValue)
}

// Get gets min latency value index
func (l *Latency) Get() int {
	if l.size == 1 {
		return 0
	}
	var (
		minIndex   = 0
		minLatency = l.history[0]
		isFail     = 0
	)
	for i, lat := range l.history {
		if lat == InitValue {
			return l.Rand()
		}
		if lat == FailValue {
			isFail++
			continue
		}
		if i == 0 {
			continue
		}
		if minLatency < 0 {
			minIndex = i
			minLatency = lat
			continue
		}
		if lat > 0 && lat < minLatency {
			minIndex = i
			minLatency = lat
		}
	}
	if isFail == l.size {
		l.Reset()
		return l.Rand()
	}
	return minIndex
}

// Rand gets random index
func (l *Latency) Rand() int {
	if l.size == 1 {
		return 0
	}
	var count int
	for {
		if count > l.size {
			return l.Get()
		}
		idx := rand.Intn(l.size*10) % l.size
		if l.history[idx] == FailValue {
			count++
			continue
		}
		return idx
	}
}

// History returns all latency values
func (l *Latency) History() []int64 {
	cp := make([]int64, len(l.history))
	copy(cp, l.history)
	return cp
}

// GetValue get value by index,
// if index is over size, return error
func (l *Latency) GetValue(idx int) (int64, error) {
	if l.size <= idx {
		return 0, ErrorOversize
	}
	return l.history[idx], nil
}
