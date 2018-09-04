package sysprocfs

import (
	"io/ioutil"
	"strconv"
)

const (
	procsFilesDir = "/proc"
)

// ProcsCount returns number procs count
func ProcsCount() (int64, error) {
	var (
		count int64
	)
	files, err := ioutil.ReadDir(procsFilesDir)
	if len(files) > 0 {
		for _, file := range files {
			_, err := strconv.Atoi(file.Name())
			if err == nil {
				count++
			}
		}
	}
	return count, err
}
