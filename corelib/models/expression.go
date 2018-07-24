package models

// Expression is expression for metrics judger
type Expression struct {
	ID            int     `json:"id" db:"id"`
	Expression    string  `json:"expression" db:"expression"`
	Operator      string  `json:"operator" db:"op"`
	RightValueStr string  `json:"-" db:"right_value"` // critical value
	RightValue    float64 `json:"right_value" db:"-"` // critical value
	MaxStep       int     `json:"max_step" db:"max_step"`
	Priority      int     `json:"priority" db:"priority"`
	Note          string  `json:"note" db:"note"`
}
