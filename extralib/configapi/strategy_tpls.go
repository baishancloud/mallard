package configapi

import (
	"fmt"
	"html/template"
	"math"
	"reflect"
	"strings"

	"github.com/baishancloud/mallard/corelib/utils"
)

func silentRecovery() {
	recover()
}

var strategyTplFuncs = template.FuncMap{
	"bytesformat": func(value interface{}) string {
		defer silentRecovery()

		var size float64

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			size = float64(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			size = float64(v.Uint())
		case reflect.Float32, reflect.Float64:
			size = v.Float()
		default:
			return ""
		}

		var KB float64 = 1 << 10
		var MB float64 = 1 << 20
		var GB float64 = 1 << 30
		var TB float64 = 1 << 40
		var PB float64 = 1 << 50

		filesizeFormat := func(filesize float64, suffix string) string {
			return strings.Replace(fmt.Sprintf("%.1f %s", filesize, suffix), ".0", "", -1)
		}

		var result string
		if size < KB {
			result = filesizeFormat(size, "bytes")
		} else if size < MB {
			result = filesizeFormat(size/KB, "KB")
		} else if size < GB {
			result = filesizeFormat(size/MB, "MB")
		} else if size < TB {
			result = filesizeFormat(size/GB, "GB")
		} else if size < PB {
			result = filesizeFormat(size/TB, "TB")
		} else {
			result = filesizeFormat(size/PB, "PB")
		}

		return result
	},
	"mul": func(a interface{}, v ...interface{}) float64 {
		val, _ := utils.ToFloat64(a)
		for _, b := range v {
			b2, _ := utils.ToFloat64(b)
			val = val * b2
		}
		return val
	},
	"round": round,
	"round2": func(a interface{}) float64 {
		return round(a, 2)
	},
}

var round = func(a interface{}, p interface{}) float64 {
	roundOn := .5
	val, _ := utils.ToFloat64(a)
	places, _ := utils.ToFloat64(p)

	var round float64
	pow := math.Pow(10, places)
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}
