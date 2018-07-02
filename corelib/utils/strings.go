package utils

import "strings"

// AddSlash adds slash to note for quotes
func AddSlash(val string) string {
	val = strings.Replace(val, "\"", "\\\"", -1)
	val = strings.Replace(val, "'", "\\'", -1)
	return val
}
