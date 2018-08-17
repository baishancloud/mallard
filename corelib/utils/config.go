package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// ReadConfigFile reads json config from file
func ReadConfigFile(file string, v interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	return decoder.Decode(v)
}

// WriteConfigFile writes value to json config file pretty
func WriteConfigFile(file string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, b, os.ModePerm)
}
