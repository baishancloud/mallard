package eventdata

import (
	"bytes"
	"errors"
	"html/template"
	"strconv"
	"strings"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/configapi"
)

func getStrategyID(eid string) int {
	if strings.HasPrefix(eid, "s_") {
		sl := strings.Split(eid, "_")
		if len(sl) < 3 {
			return 0
		}
		v, _ := strconv.Atoi(sl[1])
		return v
	}
	return 0
}

var (
	// ErrStrategyIDZero means startegy id is 0
	ErrStrategyIDZero = errors.New("strategy-id-zero")
	// ErrStrategyNil means strategy value is nil
	ErrStrategyNil = errors.New("strategy-nil")
)

// Convert converts simple event to full event with strategy info
func Convert(event *models.Event, fillStrategy bool) (*models.EventFull, string, error) {
	evt := &models.EventFull{
		ID:          event.ID,
		Status:      event.Status.String(),
		EventTime:   event.Time,
		Endpoint:    event.Endpoint,
		LeftValue:   event.LeftValue,
		CurrentStep: event.Step,
		PushedTags:  event.FullTags(),
		Fields:      event.Fields,
	}
	var err error
	if fillStrategy {
		strategyID := event.Strategy
		if strategyID < 1 {
			strategyID = getStrategyID(event.ID)
		}
		if strategyID == 0 {
			return nil, "", ErrStrategyIDZero
		}
		if strings.HasPrefix(event.ID, "s_") || strings.HasPrefix(event.ID, "nodata_s_") {
			st := configapi.GetStrategyByID(strategyID)
			if st == nil {
				return nil, "", ErrStrategyNil
			}
			st2 := new(models.Strategy)
			*st2 = *st
			st2.Note, err = renderEventNote(st2.Note, evt)
			if err != nil {
				return nil, "", err
			}
			evt.Strategy = st2
			if len(st2.MarkTags) > 0 {
				evt.PushedTags["strategy_tags"] = strings.Join(st2.MarkTags, ",")
			}
			return evt, st2.Note, err
		}
	}
	return evt, "", nil
}

func renderEventNote(note string, event *models.EventFull) (string, error) {
	var result string
	if !strings.Contains(note, "{{") {
		result = note
	} else {
		buffer := bytes.NewBuffer(nil)
		noteTmpl, err := template.New("note").Parse(note)
		if err != nil {
			result = note
			return result, err
		}
		if err = noteTmpl.Execute(buffer, event); err != nil {
			result = note
			return result, err
		}
		if buffer.Len() > 0 {
			result = buffer.String()
		}
	}
	if strings.HasPrefix(event.ID, "nodata_") {
		result = "[NODATA]" + result
	}
	return result, nil
}
