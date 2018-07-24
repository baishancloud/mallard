package msggcall

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
	"github.com/baishancloud/mallard/componentlib/compute/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/utils"
)

type msggRequest struct {
	SendRequest *sqldata.AlarmSendRequest `json:"request,omitempty"`
	Recover     bool                      `json:"recover,omitempty"`
	Note        string                    `json:"note,omitempty"`
	Endpoint    string                    `json:"endpoint,omitempty"`
	LeftValue   float64                   `json:"left_value,omitempty"`
	Level       int                       `json:"level,omitempty"`
	Time        int64                     `json:"time,omitempty"`
	Sertypes    string                    `json:"sertypes,omitempty"`
	Status      string                    `json:"status,omitempty"`
	EventID     string                    `json:"event_id,omitempty"`
}

var (
	msggCallCount      = expvar.NewDiff("alert.msgg_call")
	msggCallErrorCount = expvar.NewDiff("alert.msgg_call_error")
	msggCallZeroCount  = expvar.NewDiff("alert.msgg_call_zero")
	msggWaitCount      = expvar.NewBase("alert.msgg_wait")
	msggMarkCount      = expvar.NewDiff("alert.msgg_mark")
	msggMergeCount     = expvar.NewDiff("alert.msgg_merge")
)

func init() {
	expvar.Register(msggCallCount, msggCallErrorCount, msggCallZeroCount, msggWaitCount, msggMarkCount, msggMergeCount)
}

// ScanRequests starts scanning requests
func ScanRequests(interval time.Duration, mergeLevel int, mergeSize int) {
	var count int64
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		count++
		slowFlag := count%2 == 0
		now := <-ticker.C
		nowUnix := now.Unix()
		requestsLock.Lock()
		var wait int64
		shouldAlarms := make(map[string]*msggRequest)
		for eid, reqs := range requests {
			if redisdata.HasNoAlarmFlag(eid) {
				for t, msggReq := range reqs {
					if msggReq.Recover {
						continue
					}
					msggMarkCount.Incr(1)
					delete(reqs, t)
					log.Debug("clean-NOALARM", "eid", eid, "t", t)
				}
				continue
			}
			for t, msggReq := range reqs {
				if slowFlag && msggReq.Level >= mergeLevel {
					log.Debug("wait-low-level", "eid", eid, "t", t, "diff", nowUnix-t, "step", msggReq.SendRequest.Step)
					wait++
					continue
				}
				if nowUnix >= t {
					// go runMsggRequest(eid, msggReq)
					delete(requests[eid], t)
					log.Debug("del-req-timeup", "eid", eid, "t", t)
					shouldAlarms[eid] = msggReq
				} else {
					wait++
					log.Debug("wait-req", "eid", eid, "t", t, "diff", nowUnix-t, "step", msggReq.SendRequest.Step)
				}
			}
			if len(requests[eid]) == 0 {
				delete(requests, eid)
				log.Debug("clean-EMPTY", "eid", eid)
			}
		}
		msggWaitCount.Set(wait)
		requestsLock.Unlock()
		handleRequests(shouldAlarms, mergeSize)
		if count == 100 {
			count = 0 // reset number
		}
	}
}

func handleRequests(requests map[string]*msggRequest, mergeSize int) {
	mergedRequests := make(map[string][]*msggRequest)
	for eid, req := range requests {
		if req.Recover {
			mergedRequests[eid] = append(mergedRequests[eid], req)
			continue
		}
		idx := strings.LastIndex(eid, "_")
		if idx < 0 {
			mergedRequests[eid] = append(mergedRequests[eid], req)
			continue
		}
		key := eid[:idx]
		mergedRequests[key] = append(mergedRequests[key], req)
	}
	for key, reqs := range mergedRequests {
		if len(reqs) < mergeSize {
			for _, req := range reqs {
				go runMsggRequest(req.EventID, req)
			}
			continue
		}
		log.Debug("merge-msgg", "key", key, "reqs", len(reqs))
		var eps []string
		var note []rune
		for _, req := range reqs {
			eps = append(eps, req.Endpoint)
			if len(note) < 512 && req.Note != "" {
				note = append(note, []rune(";")...)
				note = append(note, []rune(req.Note)...)
			}
		}
		eps = utils.StringSliceUnique(eps)
		sort.Sort(sort.StringSlice(eps))
		onlyReq := reqs[0]
		onlyReq.Endpoint = strings.Join(eps, ",")
		onlyReq.Note = "【共 " + strconv.Itoa(len(reqs)) + " 条】" + strings.TrimPrefix(string(note), ";") + "..."
		go runMsggRequest(onlyReq.EventID, onlyReq)
		msggMergeCount.Incr(1)
	}
}

func runMsggRequest(eid string, msggReq *msggRequest) {
	msggCallCount.Incr(1)
	args := lineMsggRequest(eid, msggReq)
	if len(args) == 0 {
		log.Info("call-zero", "eid", eid, "status", msggReq.Status, "note", msggReq.Note)
		msggCallZeroCount.Incr(1)
		return
	}
	output, err := runWithTimeout(msggFile, args, time.Second*30)
	if err != nil {
		log.Warn("call-msgg-error", "error", err, "eid", eid)
		msggCallErrorCount.Incr(1)
		return
	}
	log.Info("call-msgg", "eid", eid, "status", msggReq.Status, "args", args, "output", string(output))
	go calculateUserCount(msggReq)
}

func lineMsggRequest(eid string, req *msggRequest) []string {
	s := []string{req.Note,
		fmt.Sprint(req.Level),
		req.Endpoint,
		fmt.Sprint(req.LeftValue),
		time.Unix(req.Time, 0).Format("01/02 15:04:05"),
		req.Sertypes,
	}
	var isLined bool
	if len(req.SendRequest.Emails) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(req.SendRequest.Emails, ","))
		isLined = true
	}
	if len(req.SendRequest.Wechats) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(req.SendRequest.Wechats, ","))
		isLined = true
	}
	if len(req.SendRequest.SMSs) == 0 {
		s = append(s, ",")
	} else {
		s = append(s, strings.Join(req.SendRequest.SMSs, ","))
		isLined = true
	}
	if !req.Recover {
		if len(req.SendRequest.Phones) == 0 {
			s = append(s, ",")
		} else {
			s = append(s, strings.Join(req.SendRequest.Phones, ","))
			isLined = true
		}
	} else {
		s = append(s, ",")
		s[0] = "【报警恢复】" + req.Note
	}

	teamName := req.SendRequest.Template + ","
	if req.SendRequest.Template != "000" {
		isLined = true
	}

	if req.SendRequest.Team == "" {
		teamName += "000"
	} else {
		teamName += req.SendRequest.Team
		isLined = true
	}
	teamName += ","
	if req.SendRequest.TeamCN == "" {
		teamName += "000"
	} else {
		teamName += req.SendRequest.TeamCN
		isLined = true
	}
	if !isLined {
		return nil
	}
	s = append(s, teamName, eid)
	return s
}
