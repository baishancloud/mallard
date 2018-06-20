package multijudge

import (
	"encoding/json"
	"io/ioutil"

	"github.com/baishancloud/mallard/corelib/models"
)

func init() {
	t := map[int]*MultiStrategy{
		1: &MultiStrategy{
			Strategies: []*models.Strategy{
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
			},
		},
		2: &MultiStrategy{
			Strategies: []*models.Strategy{
				&models.Strategy{
					Metric:         "sys.load",
					FieldTransform: "select(1min)",
					Func:           "all(#1)",
					Operator:       "<",
					RightValue:     4,
					GroupByStr:     "endpoint",
					Score:          0.8,
				},
				&models.Strategy{
					Metric:         "sys.cpu_core",
					FieldTransform: "select(use)",
					Func:           "all(#1)",
					Operator:       "<",
					RightValue:     4,
					GroupByStr:     "endpoint",
					Score:          0.2,
				},
			},
		},
	}
	b, _ := json.MarshalIndent(t, "", "\t")
	ioutil.WriteFile("testdata.json", b, 0644)
	SetStrategies(t)
}
