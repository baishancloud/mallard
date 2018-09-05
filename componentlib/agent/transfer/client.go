package transfer

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/utils"
)

var (
	tfrClient = NewClient(time.Second*10, 5, "mallard2-agent")
)

// Client is simple client to send data to url
type Client struct {
	timeout   time.Duration
	transport *http.Transport
	token     string
}

// ClientError is error fo client response error
type ClientError struct {
	Status int
	Body   []byte
}

// Error implements error
func (ce ClientError) Error() string {
	return fmt.Sprintf("bad-status-%d,%s", ce.Status, ce.Body)
}

// NewClient create client object with retry time, timeout and maxConn number
func NewClient(timeout time.Duration, maxConn int, token string) *Client {
	tr := &http.Transport{
		MaxIdleConns:          maxConn,
		MaxIdleConnsPerHost:   3,
		ResponseHeaderTimeout: timeout,
		IdleConnTimeout:       time.Minute,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &Client{
		timeout:   timeout,
		transport: tr,
		token:     token,
	}
}

// GET request data to url with headers
func (c *Client) GET(url string, headers map[string]string) (*http.Response, time.Duration, error) {
	st := time.Now()
	resp, err := c.requestOnce("GET", url, headers, nil)
	return resp, time.Since(st), err
}

// POST post bytes data to url
func (c *Client) POST(url string, data interface{}, dataLen int) (*http.Response, time.Duration, error) {
	buf, err := utils.GzipJSON(data, 10240)
	if err != nil {
		return nil, 0, err
	}
	var (
		resp *http.Response
		st   = time.Now()
	)
	headers := map[string]string{
		"Data-Length":  strconv.Itoa(dataLen),
		"Content-Type": "application/gzip+json",
	}
	resp, err = c.requestOnce("POST", url, headers, buf)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, 0, ClientError{
			Status: resp.StatusCode,
			Body:   body,
		}
	}
	return resp, time.Since(st), err
}

func (c *Client) requestOnce(method string, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	req.Header.Set("Agent-Endpoint", serverinfo.Hostname())
	req.Header.Set("User-Agent", "mallard2-agent")
	if c.token != "" {
		for k, v := range httptoken.BuildHeader(c.token) {
			req.Header.Add(k, v)
		}
	}
	client := &http.Client{
		Timeout:   c.timeout,
		Transport: c.transport,
	}
	return client.Do(req)
}
