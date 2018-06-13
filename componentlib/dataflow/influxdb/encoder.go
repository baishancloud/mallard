package influxdb

import (
	"bytes"
	"errors"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	client "github.com/influxdata/influxdb/client/v2"
)

// Encoder is encoder for influxdb point
type Encoder struct {
	Db string
}

const (
	precision = "s"
)

// Encode inplement encode method
func (ic Encoder) Encode(points []*client.Point) ([]byte, error) {
	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  ic.Db,
		Precision: precision,
	})
	if err != nil {
		return nil, err
	}
	batchPoints.AddPoints(points)
	b := bytes.NewBuffer(make([]byte, 0, 102400))
	for _, p := range batchPoints.Points() {
		if _, err := b.WriteString(p.PrecisionString(precision)); err != nil {
			return nil, err
		}
		if err := b.WriteByte('\n'); err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}

// ErrorEmptyValue means metric value is empty
var ErrorEmptyValue = errors.New("empty-name-endpoint")

func metric2Point(m *models.Metric) (*client.Point, error) {
	if m.Endpoint == "" || m.Name == "" {
		return nil, ErrorEmptyValue
	}
	mf := map[string]interface{}{
		"value": m.Value,
	}
	for k, v := range m.Fields {
		mf[k] = v
	}
	mt := map[string]string{
		"endpoint": m.Endpoint,
	}
	for k, v := range m.Tags {
		mt[k] = v
	}
	return client.NewPoint(
		m.Name,
		mt,
		mf,
		time.Unix(m.Time, 0),
	)
}
