package utils

import "strings"

// AddSlash adds slash to note for quotes
func AddSlash(val string) string {
	val = strings.Replace(val, "\"", "\\\"", -1)
	val = strings.Replace(val, "'", "\\'", -1)
	return val
}

// Substr concats string with proper size
func Substr(val string, size int) string {
	data := []rune(val)
	if len(data) < size {
		return val
	}
	return string(data[:size])
}
