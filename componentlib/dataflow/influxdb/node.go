package influxdb

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/baishancloud/mallard/corelib/expvar"
)

var (
	transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     time.Minute * 3,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
)
var (
	conflictKeyword = []byte("conflict")
	timeoutKeyword  = []byte("timeout")
	refusedKeyword  = "connection refused"
)

// Node is one influxdb node
type Node struct {
	URL       string
	User      string
	Password  string
	Name      string
	GroupName string
	client    *http.Client

	failCounter     *expvar.DiffMeter
	sendCounter     *expvar.DiffMeter
	reqCounter      *expvar.DiffMeter
	retryCount      *expvar.DiffMeter
	conflictCounter *expvar.DiffMeter
	latencyCounter  *expvar.AvgMeter
	sizeCounter     *expvar.AvgMeter
}

// NodeOption is option of a node
type NodeOption struct {
	URL       string
	User      string
	Password  string
	Name      string
	GroupName string
	Timeout   time.Duration
}

// NewNode creates one node with option
func NewNode(opt NodeOption) *Node {
	n := &Node{
		URL:             opt.URL,
		User:            opt.User,
		Password:        opt.Password,
		Name:            opt.Name,
		GroupName:       opt.GroupName,
		failCounter:     expvar.NewDiff("fail"),
		sendCounter:     expvar.NewDiff("metric"),
		reqCounter:      expvar.NewDiff("req"),
		latencyCounter:  expvar.NewAverage("duration", 50),
		retryCount:      expvar.NewDiff("retry"),
		conflictCounter: expvar.NewDiff("conflict"),
		sizeCounter:     expvar.NewAverage("size", 50),
	}
	n.client = &http.Client{
		Timeout:   opt.Timeout,
		Transport: transport,
	}
	return n
}

// Send send bytes to node
func (n *Node) Send(data []byte, pLen int64, retry bool) {
	st := time.Now()
	n.reqCounter.Incr(1)

	request, err := http.NewRequest("POST", n.URL, bytes.NewReader(data))
	if err != nil {
		log.Warn("req-error", "g", n.Name, "len", pLen, "error", err)
		n.failCounter.Incr(pLen)
		return
	}
	request.Header.Add("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	if n.User != "" {
		request.SetBasicAuth(n.User, n.Password)
	}
	resp, err := n.client.Do(request)
	if err != nil {
		log.Warn("send-error", "g", n.Name, "len", pLen, "error", err)
		n.failCounter.Incr(pLen)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		n.failCounter.Incr(1)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Warn("send-fail", "g", n.Name, "len", pLen, "status", resp.StatusCode, "resp", string(body))
		if bytes.Contains(body, conflictKeyword) {
			log.Warn("conflict-fail", "g", n.Name, "len", pLen)
			n.conflictCounter.Incr(1)
			return
		}
		if bytes.Contains(body, timeoutKeyword) {
			log.Warn("timeout-fail", "g", n.Name, "len", pLen)
			return
		}
		if retry {
			log.Debug("retry", "g", n.Name, "len", pLen)
			n.Send(data, pLen, false)
		}
		return
	}
	duration := time.Since(st).Nanoseconds() / 1e6
	// log.Debug("send-ok", "g", n.Name, "len", pLen, "status", resp.StatusCode, "du", duration)
	n.latencyCounter.Set(duration)
	n.sendCounter.Incr(pLen)
	n.sizeCounter.Set(int64(len(data)))
}

func (n *Node) counters() []interface{} {
	return []interface{}{
		n.reqCounter, n.failCounter, n.sendCounter,
		n.latencyCounter, n.conflictCounter, n.retryCount,
		n.sizeCounter,
	}
}
