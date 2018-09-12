package queues

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os/exec"
	"testing"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMetrics(t *testing.T) {
	Convey("queue", t, func() {
		queue := NewQueue(10, "dump")

		for i := 0; i < 3; i++ {
			dumpCount, ok := queue.Push(&Packet{
				Data: bytes.Repeat([]byte{
					byte(i + 1),
					byte(i + 2),
					byte(i + 3)}, 100),
				Type: i + 10,
				Len:  9,
			})
			So(ok, ShouldBeTrue)
			So(dumpCount, ShouldBeZeroValue)
		}
		So(queue.Len(), ShouldEqual, 3)

		packs, err := queue.Pop(5)
		So(err, ShouldBeNil)
		So(packs, ShouldHaveLength, 3)
		So(packs.DataLen(), ShouldEqual, 27)

		queue.Push(&Packet{
			Data: bytes.Repeat([]byte{
				byte(1),
				byte(2),
				byte(3)}, 100),
			Type: 10,
			Len:  9,
		})
		_, n, err := queue.Dump(100)
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 1)

		_, n, err = queue.ReadLatestDump()
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 1)
		So(queue.Len(), ShouldEqual, 1)

		exec.Command("rm", "-rf", "dump").Run()
	})

	Convey("metrics", t, func() {
		metrics, err := testMetricsPacks.ToMetrics()
		So(err, ShouldBeNil)
		So(metrics, ShouldHaveLength, 25)
	})

	Convey("metrics", t, func() {
		events, err := testEventsPacks.ToEvents()
		So(err, ShouldBeNil)
		So(events, ShouldHaveLength, 15)
	})
}

func genMetric() *models.Metric {
	return &models.Metric{
		Name:  "cpu",
		Value: 1.2,
		Fields: map[string]interface{}{
			"usr":    50.0 * rand.Float64(),
			"sys":    10.0 * rand.Float64(),
			"iowait": 40.0 * rand.Float64(),
		},
		Endpoint: "endpoint-1",
		Time:     time.Now().Unix() + rand.Int63n(100),
		Step:     60,
		Tags: map[string]string{
			"core": "1",
		},
	}
}

func genEvent() *models.Event {
	return &models.Event{
		ID:       "s_11_abcdeft",
		Status:   1,
		Time:     time.Now().Unix() + rand.Int63n(100),
		Strategy: 11,
		Endpoint: "endpoint-1",
		Step:     60,
		Tags: map[string]string{
			"core": "1",
		},
	}
}

func init() {
	testMetricsBytes, _ = json.Marshal(testMetrics)
	testMetricsGzipBytes, _ = utils.GzipBytes(testMetricsBytes)
	testMetricsPacks = []*Packet{
		&Packet{Data: testMetricsBytes},
		&Packet{Data: testMetricsGzipBytes, Type: PacketTypeGzip},
		&Packet{Data: testMetricsBytes},
		&Packet{Data: testMetricsGzipBytes, Type: PacketTypeGzip},
		&Packet{Data: testMetricsBytes},
	}

	testEventsBytes, _ = json.Marshal(testEvents)
	testEventsGzipBytes, _ = utils.GzipBytes(testEventsBytes)
	testEventsPacks = []*Packet{
		&Packet{Data: testEventsBytes},
		&Packet{Data: testEventsGzipBytes, Type: PacketTypeGzip},
		&Packet{Data: testEventsBytes},
		&Packet{Data: testEventsGzipBytes, Type: PacketTypeGzip},
		&Packet{Data: testEventsBytes},
	}
}

var (
	testMetricsBytes     []byte
	testMetricsGzipBytes []byte
	testMetrics                  = []*models.Metric{genMetric(), genMetric(), genMetric(), genMetric(), genMetric()}
	testPack                     = &Packet{Data: bytes.Repeat([]byte{1, 2, 3}, 100)}
	testPacks            Packets = []*Packet{testPack, testPack, testPack, testPack, testPack}
	testMetricsPacks     Packets

	testEvents          = []*models.Event{genEvent(), genEvent(), genEvent()}
	testEventsBytes     []byte
	testEventsGzipBytes []byte
	testEventsPacks     Packets
)

func BenchmarkPacks2Metrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testMetricsPacks.ToMetrics()
	}
}
