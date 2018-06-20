package multijudge

import "github.com/baishancloud/mallard/corelib/models"

var (
	testStrategies = []*models.Strategy{
		&models.Strategy{
			Metric:         "pingAggr",
			FieldTransform: "select(lost)",
			Func:           "all(#1)",
			Operator:       ">",
			RightValue:     4,
			GroupByStr:     "node",
			TagString:      "type=nodeLost",
			Score:          0.8,
		},
		&models.Strategy{
			Metric:         "node_bandwidth",
			FieldTransform: "select(value)",
			Func:           "all(#1)",
			Operator:       "==",
			RightValue:     1,
			GroupByStr:     "node",
			Score:          0.2,
		},
	}
)

func init() {
	SetStrategies(map[int]*MultiStrategy{
		1: &MultiStrategy{
			strategies: testStrategies,
		},
	})
}
