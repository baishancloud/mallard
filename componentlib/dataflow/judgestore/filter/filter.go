package filter

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/zaplog"
)

type (
	// Item is filter rule to cleanup metric raw fields and tags
	Item struct {
		Tags   []string `json:"tags"`
		Fields []string `json:"fields"`
		Expire int      `json:"expire"`
	}
	// Filters is map of filter rules
	Filters map[string]Item
)

// ForMetric is filter settings for one named metric
type ForMetric struct {
	Name   string
	Tags   map[string]bool
	Fields map[string]bool
	Expire int
}

// ForMetrics is group of filter settings to some metrics
type ForMetrics map[string]*ForMetric

// ForMetricsHandler is handler to handle filters of metrics
type ForMetricsHandler func(filters ForMetrics)

// Query is query conditions for store query metrics
type Query struct {
	Metric    string            `json:"metric"`
	TimeRange [2]int64          `json:"time"`
	Tag       map[string]string `json:"tag,omitempty"`
	Endpoint  string            `json:"ep,omitempty"`
}

var (
	// ErrQueryMetricNil means metric name in query is empty
	ErrQueryMetricNil = errors.New("query-metric-nil")
	// ErrQueryTimeRangeZero means time range is zero in query
	ErrQueryTimeRangeZero = errors.New("query-time-zero")
	// ErrQueryTimeBeginOver means begin is over end in query
	ErrQueryTimeBeginOver = errors.New("query-time-begin-over-end")
	// ErrQueryTagsDismatched means tags params is not matched as pairs
	ErrQueryTagsDismatched = errors.New("query-tags-values-dismatched")
)

// IsValid returns valid status of query params
func (q Query) IsValid() error {
	if q.Metric == "" {
		return ErrQueryMetricNil
	}
	if q.TimeRange[0] == 0 || q.TimeRange[1] == 0 {
		return ErrQueryTimeRangeZero
	}
	if q.TimeRange[0] > q.TimeRange[1] {
		return ErrQueryTimeBeginOver
	}
	return nil
}

// BuildTags parses params slices to tags map
func (q Query) BuildTags(tags, values []string) (map[string]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	if len(tags) != len(values) {
		return nil, ErrQueryTagsDismatched
	}
	tv := make(map[string]string, len(tags))
	for i, t := range tags {
		tv[t] = values[i]
	}
	return tv, nil
}

// QueriedMetric is metric value
type QueriedMetric struct {
	Name     string            `json:"name"`
	Time     int64             `json:"time,omitempty"`
	Value    float64           `json:"value,omitempty"`
	Fields   interface{}       `json:"fields,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Endpoint string            `json:"endpoint,omitempty"`
}

// ParseFilter parses all filters to each metric
func ParseFilter(filters Filters) map[string]*ForMetric {
	if len(filters) == 0 {
		return nil
	}
	list := make(map[string]*ForMetric)
	for name, item := range filters {
		fl := &ForMetric{
			Name:   name,
			Tags:   make(map[string]bool),
			Fields: make(map[string]bool),
			Expire: item.Expire,
		}
		for _, tag := range item.Tags {
			fl.Tags[tag] = true
		}
		for _, fd := range item.Fields {
			fl.Fields[fd] = true
		}
		list[name] = fl
	}
	return list
}

// MetricFilter is metrics writing filter
type MetricFilter struct {
	cachedLock sync.RWMutex
	cached     ForMetrics
	updateFn   ForMetricsHandler
}

var (
	cachedLock sync.RWMutex
	cached     ForMetrics
)

// ReadFile reads filter file
func ReadFile(file string, updateFn ForMetricsHandler) error {
	filters, err := ReadFromFile(file)
	if err != nil {
		return err
	}
	if updateFn != nil {
		updateFn(filters)
	}
	cachedLock.Lock()
	cached = filters
	cachedLock.Unlock()
	return nil
}

var (
	log = zaplog.Zap("filter")
)

// SyncFile reads filter file in time loop
func SyncFile(file string, updateFn ForMetricsHandler, interval time.Duration) {
	log.Info("set-sync", "file", file, "interval", int(interval.Seconds()))

	var lastModTime int64
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		info, err := os.Stat(file)
		if err != nil {
			<-ticker.C
			continue
		}
		if info.ModTime().Unix() == lastModTime {
			<-ticker.C
			continue
		}
		if err := ReadFile(file, updateFn); err != nil {
			log.Warn("read-file-error", "file", file, "error", err)
		}
		lastModTime = info.ModTime().Unix()
	}
}

// Get gets current filters config
func Get() ForMetrics {
	cachedLock.Lock()
	defer cachedLock.Unlock()
	return cached
}

// ReadFromFile creates metric filters from file
func ReadFromFile(file string) (ForMetrics, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	filters := make(Filters)
	if err = json.Unmarshal(data, &filters); err != nil {
		return nil, err
	}
	return ParseFilter(filters), nil
}
