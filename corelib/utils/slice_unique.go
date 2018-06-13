package utils

// IntSliceUnique returns unique int slice
func IntSliceUnique(ss []int) []int {
	m := make(map[int]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	r := make([]int, 0, len(m))
	for s := range m {
		r = append(r, s)
	}
	return r
}

// Int64SliceUnique returns unique int64 slice
func Int64SliceUnique(ss []int64) []int64 {
	m := make(map[int64]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	r := make([]int64, 0, len(m))
	for s := range m {
		r = append(r, s)
	}
	return r
}

// StringSliceUnique returns unique string slice
func StringSliceUnique(ss []string) []string {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	r := make([]string, 0, len(m))
	for s := range m {
		r = append(r, s)
	}
	return r
}

// KeysOfStructMap returns keys of struct map
func KeysOfStructMap(m map[string]struct{}) []string {
	if len(m) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
