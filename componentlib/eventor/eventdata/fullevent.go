package eventdata

import (
	"errors"
	"strings"

	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/extralib/configapi"
)

var (
	// ErrStrategyIDZero means startegy id is 0
	ErrStrategyIDZero = errors.New("strategy-id-zero")
	// ErrStrategyNil means strategy value is nil
	ErrStrategyNil = errors.New("strategy-nil")
	// ErrExpressionNil means expression value is nil
	ErrExpressionNil = errors.New("expression-nil")
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
		judgeUnitID := event.JudgeUnitID()
		if judgeUnitID == 0 {
			return nil, "", ErrStrategyIDZero
		}
		if strings.HasPrefix(event.ID, "s_") || strings.HasPrefix(event.ID, "nodata_s_") {
			st := configapi.GetStrategyByID(judgeUnitID)
			if st == nil {
				return nil, "", ErrStrategyNil
			}
			st2 := new(models.Strategy)
			*st2 = *st
			st2.Note, err = renderEventNote(st2.ID, st2.Note, evt)
			if err != nil {
				st2.Note = st.Note
			}
			evt.Strategy = st2
			if len(st2.MarkTags) > 0 {
				evt.PushedTags["strategy_tags"] = strings.Join(st2.MarkTags, ",")
			}
			return evt, st2.Note, err
		}
		if strings.HasPrefix(event.ID, "e_") {
			exp := configapi.GetExpressionByID(judgeUnitID)
			if exp == nil {
				return nil, "", ErrExpressionNil
			}
			exp2 := new(models.Expression)
			*exp2 = *exp
			exp2.Note, err = renderEventNote(exp2.ID, exp2.Note, evt)
			if err != nil {
				exp2.Note = exp.Note
			}
			evt.Strategy = exp2
			return evt, exp2.Note, err
		}
	}
	return evt, "", nil
}

func renderEventNote(id int, note string, event *models.EventFull) (string, error) {
	result, err := configapi.RenderStrategyTpl(id, event)
	if err != nil {
		if result == "" {
			result = note
		}
		return result, err
	}
	if strings.HasPrefix(event.ID, "nodata_") {
		result = "[NODATA]" + result
	}
	return result, nil
}
