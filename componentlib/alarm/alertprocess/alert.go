package alertprocess

import (
	"sync"

	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

// Processor is handler to process event data
type Processor func(redisdata.EventRecord)

var (
	processorList []Processor
	processorLock sync.RWMutex

	log = zaplog.Zap("alert")

	recvDiff         = expvar.NewDiff("alert.recv")
	recvClosedDiff   = expvar.NewDiff("alert.recv_closed")
	recvOKDiff       = expvar.NewDiff("alert.recv_ok")
	recvOutdatedDiff = expvar.NewDiff("alert.recv_outdated")
	recvProblemDiff  = expvar.NewDiff("alert.recv_problem")
)

func init() {
	expvar.Register(recvDiff, recvClosedDiff, recvOKDiff, recvOutdatedDiff, recvProblemDiff)
}

// Register registers processer to handle metrics
func Register(pc ...Processor) {
	processorLock.Lock()
	processorList = append(processorList, pc...)
	processorLock.Unlock()
}

// Process starts running all metrics and errors
func Process(evtCh <-chan redisdata.EventRecord) {
	log.Info("init", "processor", len(processorList))
	go func() {
		for {
			evt := <-evtCh
			if evt.Event != nil {
				recvDiff.Incr(1)
				switch evt.Event.Status {
				case models.EventClosed.String():
					recvClosedDiff.Incr(1)
				case models.EventOk.String():
					recvOKDiff.Incr(1)
				case models.EventOutdated.String():
					recvOutdatedDiff.Incr(1)
				case models.EventProblem.String():
					recvProblemDiff.Incr(1)
				}
				go handleEvent(evt)
			}
		}
	}()
}

func handleEvent(event redisdata.EventRecord) {
	processorLock.RLock()
	for _, pc := range processorList {
		if pc == nil {
			continue
		}
		pc(event)
	}
	processorLock.RUnlock()
}
