package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
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

// MD5File returns file's hash string
func MD5File(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
