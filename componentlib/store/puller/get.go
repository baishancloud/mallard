package puller

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/baishancloud/mallard/corelib/httptoken"
)

var (
	transport = &http.Transport{
		MaxIdleConns:        300,
		MaxIdleConnsPerHost: 50,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
)

func getURL(url string, timeout time.Duration) (*http.Response, time.Duration, error) {
	t := time.Now()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range httptoken.BuildHeader("store-puller") {
		req.Header.Set(k, v)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	resp, err := client.Do(req)
	return resp, time.Since(t), err
}
