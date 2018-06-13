package msggcall

import (
	"fmt"
	"strings"
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
	"github.com/baishancloud/mallard/componentlib/compute/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
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
}

var (
	msggCallCount      = expvar.NewDiff("alert.msgg_call")
	msggCallErrorCount = expvar.NewDiff("alert.msgg_call_error")
	msggCallZeroCount  = expvar.NewDiff("alert.msgg_call_zero")
	msggWaitCount      = expvar.NewBase("alert.msgg_wait")
	msggMarkCount      = expvar.NewDiff("alert.msgg_mark")
)

func init() {
	expvar.Register(msggCallCount, msggCallErrorCount, msggCallZeroCount, msggWaitCount, msggMarkCount)
}

// ScanRequests starts scanning requests
func ScanRequests(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		now := <-ticker.C
		nowUnix := now.Unix()
		requestsLock.Lock()
		var wait int64
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
			}
			for t, msggReq := range reqs {
				if nowUnix >= t {
					go runMsggRequest(eid, msggReq)
					delete(requests[eid], t)
					log.Debug("del-req-timeup", "eid", eid, "t", t)
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
