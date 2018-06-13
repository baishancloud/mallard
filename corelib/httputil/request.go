package httputil

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
)

var (
	transport = &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	specialRole string
)

// SetSpecialRole sets special role name
func SetSpecialRole(name string) {
	specialRole = name
}

// GetJSON gets json result from http api
func GetJSON(url string, timeout time.Duration, v interface{}) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	if specialRole != "" {
		req.Header.Set("Mallard-Role", specialRole)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return resp.StatusCode, nil
	}
	if resp.Header.Get("Content-Type") == ContentTypeGzipJSON {
		return resp.StatusCode, utils.UngzipJSON(resp.Body, v)
	}
	decoder := json.NewDecoder(resp.Body)
	return resp.StatusCode, decoder.Decode(v)
}

// GetJSONWithHash gets json result from http api and hash
func GetJSONWithHash(url string, timeout time.Duration, v interface{}) (int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", err
	}
	if specialRole != "" {
		req.Header.Set("Mallard-Role", specialRole)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	hash := resp.Header.Get("Content-Hash")
	if resp.StatusCode == 304 {
		return resp.StatusCode, hash, nil
	}
	if resp.Header.Get("Content-Type") == ContentTypeGzipJSON {
		return resp.StatusCode, hash, utils.UngzipJSON(resp.Body, v)
	}
	decoder := json.NewDecoder(resp.Body)
	return resp.StatusCode, hash, decoder.Decode(v)
}

// PostJSON posts json to url
func PostJSON(url string, timeout time.Duration, value interface{}) (*http.Response, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if specialRole != "" {
		req.Header.Set("Mallard-Role", specialRole)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client.Do(req)
}
