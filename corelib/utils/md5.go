package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5HashString returns string's hash string
func MD5HashString(text string) string {
	return MD5HashBytes([]byte(text))
}

// MD5HashBytes returns bytes's hash string
func MD5HashBytes(value []byte) string {
	hasher := md5.New()
	hasher.Write(value)
	return hex.EncodeToString(hasher.Sum(nil))
}
