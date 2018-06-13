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

func getURL(url string, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range httptoken.BuildHeader("store-puller") {
		req.Header.Set(k, v)
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client.Do(req)
}
