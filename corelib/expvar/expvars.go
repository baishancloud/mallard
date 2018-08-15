package expvar

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/sysprocfs"
)

var (
	log = zaplog.Zap("stat")

	counterFactory []interface{}
	counterLock    sync.RWMutex

	currentLock sync.RWMutex
	currentVars map[string]interface{}
)

// Register registers all counters
func Register(counters ...interface{}) {
	counterLock.Lock()
	counterFactory = append(counterFactory, counters...)
	counterLock.Unlock()
}

// ExposeFactory exposes values to map
func ExposeFactory(values []interface{}, withSelf bool) map[string]interface{} {
	result := make(map[string]interface{}, len(values))
	for _, ct := range values {
		if avg, ok := ct.(Averager); ok {
			result[avg.Name()+".avg"] = utils.FixFloat(avg.Avg())
			continue
		}
		if qps, ok := ct.(QPSer); ok {
			result[qps.Name()+".qps"] = utils.FixFloat(qps.QPS())
			continue
		}
		if diff, ok := ct.(Differ); ok {
			result[diff.Name()+".diff"] = diff.Diff()
			continue
		}
		if cnt, ok := ct.(Counter); ok {
			result[cnt.Name()+".cnt"] = cnt.Count()
			continue
		}
	}
	if withSelf {
		for k, v := range getSelf() {
			result[k] = v
		}
	}
	return result
}

// Expose returns all current vars
func Expose(withSelf bool) map[string]interface{} {
	counterLock.RLock()
	result := ExposeFactory(counterFactory, withSelf)
	counterLock.RUnlock()
	currentLock.Lock()
	currentVars = result
	currentLock.Unlock()
	return result
}

func getSelf() map[string]interface{} {
	p, _ := sysprocfs.SelfProcessStat()
	if p == nil {
		return nil
	}
	result := map[string]interface{}{
		"procs.mem":         utils.FixFloat(p.Memory),
		"procs.mem_percent": utils.FixFloat(float64(p.MemoryPercent)),
		"procs.goroutine":   float64(p.Goroutines),
		"procs.cpu":         utils.FixFloat(p.CPUPercent),
		"procs.fds":         float64(p.Fds),
		"procs.read":        float64(p.Read),
		"procs.read_bytes":  float64(p.ReadBytes),
		"procs.write":       float64(p.Write),
		"procs.write_bytes": float64(p.WriteBytes),
	}
	return result
}

// HTTPHandler is http handler of expvars
func HTTPHandler(rw http.ResponseWriter, r *http.Request) {
	currentLock.RLock()
	defer currentLock.RUnlock()
	encoder := json.NewEncoder(rw)
	if err := encoder.Encode(currentVars); err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
}

// PrintAlways prints expvars to file in time loop
func PrintAlways(metricName string, file string, interval time.Duration) {
	time.Sleep(time.Second * 10)
	step := int(interval.Seconds())
	utils.Ticker(interval, printFunc(file, metricName, step))
}

func printFunc(file string, metricName string, step int) func() {
	return func() {
		values := Expose(true)
		log.Debug("stats", "perf", values)
		if file != "" {
			metric := &models.Metric{
				Name:   metricName,
				Fields: values,
				Time:   time.Now().Unix(),
				Step:   step,
			}
			b, _ := json.Marshal([]*models.Metric{metric})
			ioutil.WriteFile(file, b, os.ModePerm)
		}
	}
}
