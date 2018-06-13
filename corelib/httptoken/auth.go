package httptoken

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
)

var (
	// HashToken is hash token that adds into raw value to hash
	HashToken = "mallard2"
	// HashFormat is format of raw value, use HashToken and HashDuration to build
	HashFormat = "%s-" + HashToken + "-%d"
	// HashDuration is expiry of hash token
	HashDuration int64 = 1800

	// KeyHeader is http header key of hash raw value
	KeyHeader = "Hash-Key"
	// CodeHeader is http header key of hashed value
	CodeHeader = "Hash-Code"
)

// BuildHeader build hash header with dataType and local ip
func BuildHeader(dataType string, t ...int64) map[string]string {
	var now int64
	if len(t) > 0 {
		now = t[0]
	} else {
		now = time.Now().Unix()
	}
	hostname, _ := os.Hostname()
	hour := now / HashDuration
	key := utils.MD5HashString(dataType + HashToken + hostname)
	code := utils.MD5HashString(fmt.Sprintf(HashFormat, key, hour))
	return map[string]string{
		KeyHeader:  key,
		CodeHeader: code,
	}
}

// CheckHeader check hash with hash requirement
func CheckHeader(header http.Header) bool {
	key := header.Get(KeyHeader)
	if key == "" {
		return false
	}
	code := header.Get(CodeHeader)
	if code == "" {
		return false
	}
	hour := time.Now().Unix() / HashDuration
	myCode := utils.MD5HashString(fmt.Sprintf(HashFormat, key, hour))
	if myCode == code {
		return true
	}
	lastHour := hour - 1
	myCode = utils.MD5HashString(fmt.Sprintf(HashFormat, key, lastHour))
	if myCode == code {
		return true
	}
	nextHour := hour + 1
	myCode = utils.MD5HashString(fmt.Sprintf(HashFormat, key, nextHour))
	if myCode == code {
		return true
	}
	return false
}

// CheckHeaderResponse check http rquest and write response if fail, return false
func CheckHeaderResponse(rw http.ResponseWriter, r *http.Request) bool {
	if !CheckHeader(r.Header) {
		rw.Header().Add("Connection", "close")
		rw.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
