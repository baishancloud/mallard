package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
)

// GzipBytes compress bytes to bytes
func GzipBytes(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)/3))
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GzipJSON encodes value as compressed json bytes
func GzipJSON(value interface{}, cap int) (io.Reader, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, cap))
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

// GzipJSONBytes encodes values as compressed json and returns bytes
func GzipJSONBytes(value interface{}, cap int) ([]byte, error) {
	reader, err := GzipJSON(value, cap)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

// UngzipJSON decodes compressed bytes to json object
func UngzipJSON(reader io.Reader, value interface{}) error {
	rd, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer rd.Close()

	decoder := json.NewDecoder(rd)
	return decoder.Decode(value)
}

// UngzipBytes uncompresses bytes
func UngzipBytes(data []byte) ([]byte, error) {
	rd, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	return ioutil.ReadAll(rd)
}
