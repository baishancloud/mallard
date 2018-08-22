package configapi

import (
	"bytes"
	"html/template"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStrategyTpls(t *testing.T) {
	Convey("tpl-funcs", t, func() {
		tplText := "round {{mul .Value 100 | round2}}"
		noteTmpl, err := template.New("note").Funcs(strategyTplFuncs).Parse(tplText)
		So(err, ShouldBeNil)
		buffer := bytes.NewBuffer(nil)
		err = noteTmpl.Execute(buffer, map[string]interface{}{
			"Value": 0.78289437837,
		})
		So(err, ShouldBeNil)
		// return buffer.String(), nil
		So(buffer.String(), ShouldEqual, "round 78.29")
	})
}
