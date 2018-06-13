package models

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/baishancloud/mallard/corelib/utils"
)

// Metric is one value of monitoring metric
type Metric struct {
	Name     string                 `json:"name,omitempty"`
	Time     int64                  `json:"time,omitempty"`
	Value    float64                `json:"value,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	Tags     map[string]string      `json:"tags,omitempty"`
	Endpoint string                 `json:"endpoint,omitempty"`
	Step     int                    `json:"step,omitempty"`
}

// Hash generates unique hash of the metric
func (m *Metric) Hash() string {
	hash0 := utils.MD5HashString(m.Name)
	hash1 := utils.MD5HashString(m.TagString(true))
	return hash0[:4] + hash1[:28] + utils.MD5HashString(m.Endpoint)[:8]
}

var metricHashSkipTag = map[string]bool{
	"sertypes":       true,
	"cachegroup":     true,
	"storagegroup":   true,
	"hang_status":    true,
	"use_status":     true,
	"fault_status":   true,
	"agent_endpoint": true,
}

// TagString returns tags and endpoint string as keyword for the metric
func (m *Metric) TagString(withEndpoint bool) string {
	str := make([]string, 0, len(m.Tags)+2)
	str = append(str, "name="+m.Name)
	if withEndpoint {
		str = append(str, "endpoint="+m.Endpoint)
	}
	for tag, value := range m.Tags {
		if metricHashSkipTag[tag] {
			continue
		}
		str = append(str, tag+"="+value)
	}
	sort.Sort(sort.StringSlice(str))
	return strings.Join(str, ",")
}

// String prints memory friendly
func (m Metric) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// FillTags fill addon tag values
func (m *Metric) FillTags(sertypes, cachegroup, storagegroup, gendpoint string) {
	if len(m.Tags) == 0 {
		m.Tags = make(map[string]string)
	}
	if m.Tags["sertypes"] == "" && sertypes != "" {
		m.Tags["sertypes"] = sertypes
	}
	if m.Tags["cachegroup"] == "" && cachegroup != "" {
		m.Tags["cachegroup"] = cachegroup
	}
	if m.Tags["storagegroup"] == "" && storagegroup != "" {
		m.Tags["storagegroup"] = storagegroup
	}
	if m.Endpoint != gendpoint {
		m.Tags["agent_endpoint"] = gendpoint
	}
}

// FullTags return all tags with serv and endpoint data
func (m *Metric) FullTags() map[string]string {
	fullTags := make(map[string]string, len(m.Tags)+1)
	for k, v := range m.Tags {
		fullTags[k] = v
	}
	fullTags["endpoint"] = m.Endpoint
	return fullTags
}

// MetricRaw is old metric data struct
type MetricRaw struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Fields    string      `json:"fields"`
	Step      int         `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
	Raw       interface{} `json:"-"`
}

// ToNew converts old Metric to new Metric object
func (m *MetricRaw) ToNew() (*Metric, error) {
	floatValue, err := utils.ToFloat64(m.Value)
	if err != nil {
		return nil, err
	}
	mc := &Metric{
		Name:     m.Metric,
		Time:     m.Timestamp,
		Step:     m.Step,
		Value:    floatValue,
		Endpoint: m.Endpoint,
	}
	if mc.Fields, err = ExtractFields(m.Fields); err != nil {
		return nil, err
	}
	if mc.Tags, err = ExtractTags(m.Tags); err != nil {
		return nil, err
	}
	return mc, nil
}
