package processor

import (
	"sync"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

// Processor is metric values processer
type Processor func([]*models.Metric)

var (
	processorList []Processor
	processorLock sync.RWMutex

	log = zaplog.Zap("processor")
)

// Register registers processer to handle metrics
func Register(pc ...Processor) {
	processorLock.Lock()
	processorList = append(processorList, pc...)
	processorLock.Unlock()
}

// Process starts running all metrics and errors
func Process(mCh <-chan []*models.Metric, evtCh <-chan []*models.Event, eCh <-chan error) {
	go processMetrics(mCh)
	go processEvents(evtCh)
	go processError(eCh)
}

func processMetrics(mCh <-chan []*models.Metric) {
	for {
		metrics, ok := <-mCh
		if len(metrics) == 0 {
			continue
		}
		metrics = FillMetrics(metrics)
		go handleMetrics(metrics)
		if !ok {
			log.Info("metrics-break")
			break
		}
	}
}

func handleMetrics(metrics []*models.Metric) {
	processorLock.RLock()
	for _, pc := range processorList {
		if pc == nil {
			continue
		}
		pc(metrics)
	}
	processorLock.RUnlock()
}

func processError(eCh <-chan error) {
	for {
		err, ok := <-eCh
		if err == nil {
			continue
		}
		log.Warn("fail", "error", err)
		if !ok {
			log.Info("error-break")
			break
		}
	}
}

// EventProcessor is event values processer
type EventProcessor func([]*models.Event)

var (
	eventProcessorList []EventProcessor
	eventProcessorLock sync.RWMutex
)

// RegisterEvent registers processer to handle events
func RegisterEvent(pc ...EventProcessor) {
	eventProcessorLock.Lock()
	eventProcessorList = append(eventProcessorList, pc...)
	eventProcessorLock.Unlock()
}

func processEvents(eCh <-chan []*models.Event) {
	for {
		events, ok := <-eCh
		if len(events) == 0 {
			continue
		}
		go handleEvents(events)
		if !ok {
			log.Info("events-break")
			break
		}
	}
}

func handleEvents(events []*models.Event) {
	eventProcessorLock.RLock()
	for _, pc := range eventProcessorList {
		if pc == nil {
			continue
		}
		pc(events)
	}
	eventProcessorLock.RUnlock()
}
