package sqldata

import (
	"strconv"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	selectTemplatesSQL = "SELECT id, template_name, parent_id, action_id, create_user FROM template ORDER BY id ASC"
)

// ReadTemplates query templates as id-template map
func ReadTemplates() (map[int]*models.Template, error) {
	rows, err := portalDB.Queryx(selectTemplatesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		templates = make(map[int]*models.Template)
		count     int
	)
	for rows.Next() {
		var s models.Template
		if err = rows.StructScan(&s); err != nil {
			continue
		}
		if s.ID > 0 {
			templates[s.ID] = &s
		}
		count++
	}
	return templates, nil
}

var (
	selectStrategiesSQL = "SELECT id, metric, tags, func, field_transform, op, right_value, max_step,step, priority, note, template_id, run_begin, run_end, recover_notify, no_data, silences_time, mark_tags, status FROM strategy ORDER BY id ASC"
)

// ReadStrategies reads all strategies as map
func ReadStrategies() (map[int]*models.Strategy, error) {
	rows, err := portalDB.Queryx(selectStrategiesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		strategies = make(map[int]*models.Strategy)
		count      int
	)
	for rows.Next() {
		var s models.Strategy
		if err = rows.StructScan(&s); err != nil {
			log.Warn("strategy-scan-error", "error", err)
			continue
		}
		if s.RightValueStr == "" {
			continue
		}
		rightValue, err := strconv.ParseFloat(s.RightValueStr, 64)
		if err != nil {
			continue
		}
		s.RightValue = rightValue
		if s.ID > 0 {
			s2 := &s
			s2.Marks()
			strategies[s2.ID] = s2
		}
		count++
	}
	return strategies, nil
}

var (
	selectExpressionSQL = "SELECT id,expression,op,right_value,max_step,priority,note FROM expression ORDER BY id ASC"
)

// ReadExpressions reads all expressions as map
func ReadExpressions() (map[int]*models.Expression, error) {
	rows, err := portalDB.Queryx(selectExpressionSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		exps  = make(map[int]*models.Expression)
		count int
	)
	for rows.Next() {
		var s models.Expression
		if err = rows.StructScan(&s); err != nil {
			log.Warn("expression-scan-error", "error", err)
			continue
		}
		if s.RightValueStr == "" {
			continue
		}
		rightValue, err := strconv.ParseFloat(s.RightValueStr, 64)
		if err != nil {
			continue
		}
		s.RightValue = rightValue
		if s.ID > 0 {
			s2 := &s
			exps[s2.ID] = s2
		}
		count++
	}
	return exps, nil
}
