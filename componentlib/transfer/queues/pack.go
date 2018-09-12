package queues

import (
	"encoding/json"
	"io"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
)

const (
	// PacketTypeGzip means values is gzipped
	PacketTypeGzip = 1
)

type (
	// Packet is alias of bytes
	Packet struct {
		Data []byte `json:"data,omitempty"`
		Type int    `json:"type,omitempty"`
		Len  int    `json:"len,omitempty"`
		Time int64  `json:"time,omitempty"`
	}
	// Packets is list of several packet
	Packets []Packet
)

// Decode decodes pack to value
func (p Packet) Decode(v interface{}) error {
	data := p.Data
	if p.Type == PacketTypeGzip {
		var err error
		data, err = utils.UngzipBytes(data)
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(data, v)
}

// ToMetrics converts packets to metrics
func (ps Packets) ToMetrics() ([]*models.Metric, error) {
	metrics := make([]*models.Metric, 0, len(ps)*10)
	for _, p := range ps {
		ms := make([]*models.Metric, 0, 10)
		if err := p.Decode(&ms); err != nil {
			return nil, err
		}
		metrics = append(metrics, ms...)
	}
	return metrics, nil
}

// ToMetricsList converts packets to metrics list, same length slice with packets count
func (ps Packets) ToMetricsList() ([][]*models.Metric, error) {
	metrics := make([][]*models.Metric, 0, len(ps))
	for _, p := range ps {
		ms := make([]*models.Metric, 0, 10)
		if err := p.Decode(&ms); err != nil {
			return nil, err
		}
		metrics = append(metrics, ms)
	}
	return metrics, nil
}

// ToEvents converts packets to events
func (ps Packets) ToEvents() ([]*models.Event, error) {
	events := make([]*models.Event, 0, len(ps)*5)
	for _, p := range ps {
		es := make([]*models.Event, 0, 5)
		if err := p.Decode(&es); err != nil {
			return nil, err
		}
		events = append(events, es...)
	}
	return events, nil
}

// DataLen returns data length of the packets
func (ps Packets) DataLen() int {
	var count int
	for _, p := range ps {
		count += p.Len
	}
	return count
}

// PacketsFromReader reads packets from io.Reader
func PacketsFromReader(reader io.Reader, dataLen int) (Packets, error) {
	packets := make([]Packet, 0, dataLen)
	decoder := json.NewDecoder(reader)
	return packets, decoder.Decode(&packets)
}
