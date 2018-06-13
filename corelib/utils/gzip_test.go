package utils

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var gzipTestData = map[string]interface{}{
	"a": "aaaaa",
	"b": 123567890,
	"c": []byte("hahahahahhahahha"),
}

func TestGzip(t *testing.T) {
	Convey("gzip.json", t, func() {
		data := gzipTestData
		reader, err := GzipJSON(data, 1024)
		So(err, ShouldBeNil)

		resData := make(map[string]interface{})
		err = UngzipJSON(reader, &resData)
		So(err, ShouldBeNil)

		So(resData["a"], ShouldEqual, data["a"])
		So(resData["b"], ShouldEqual, data["b"])
		So(resData["c"], ShouldEqual, base64.StdEncoding.EncodeToString(data["c"].([]byte)))
	})
}

func BenchmarkGzipJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rd, _ := GzipJSON(gzipTestData, 1024)
		ioutil.ReadAll(rd)
	}
}
