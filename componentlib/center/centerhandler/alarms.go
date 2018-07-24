package centerhandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/julienschmidt/httprouter"
)

var (
	reqAlarmCount = expvar.NewDiff("http.req_alarm")
)

func init() {
	expvar.Register(reqAlarmCount)
}

func alarmsAllData(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqAlarmCount.Incr(1)
	alarms := sqldata.AlarmsAll()
	if alarms == nil {
		httputil.Response404(rw, r)
		return
	}
	hash := r.FormValue("crc")
	if hash == fmt.Sprint(alarms.CRC) {
		httputil.Response304(rw, r)
		return
	}
	rw.Header().Set("Content-Crc", fmt.Sprint(alarms.CRC))
	isGzip := r.FormValue("gzip") != ""
	dataLen, err := httputil.ResponseJSON(rw, alarms, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-alarm-all", "r", r.RemoteAddr, "hash", alarms.CRC, "bytes", dataLen, "is_gzip", isGzip)
}

func alarmsAllRequests(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqAlarmCount.Incr(1)
	// try check crc
	hash := r.FormValue("crc")
	if hash != "" {
		alarms := sqldata.AlarmsAll()
		if alarms != nil {
			if hash == fmt.Sprint(alarms.CRC) {
				httputil.Response304(rw, r)
				return
			}
		}
	}
	// response data
	alarms := sqldata.AlarmsRequests()
	if alarms == nil {
		httputil.Response404(rw, r)
		return
	}
	isGzip := r.FormValue("gzip") != ""
	dataLen, err := httputil.ResponseJSON(rw, alarms, isGzip, false)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-alarm-requests-all", "r", r.RemoteAddr, "bytes", dataLen, "is_gzip", isGzip)
}

func alarmsOneRequest(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqAlarmCount.Incr(1)
	isGzip := (r.FormValue("gzip") != "")
	st := r.FormValue("strategy_id")
	if st == "" {
		httputil.Response404(rw, r)
		return
	}
	sid, err := strconv.Atoi(st)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	data := sqldata.AlarmsForOneStrategy(sid)
	if data == nil {
		httputil.Response404(rw, r)
		return
	}
	var tempData interface{} = data
	if r.FormValue("line") != "" {
		tempM := make(map[string]string, len(data))
		for k, req := range data {
			tempM[k] = req.Line()
		}
		tempData = tempM
	}
	dataLen, err := httputil.ResponseJSON(rw, tempData, isGzip, true)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-alarm-request", "r", r.RemoteAddr, "bytes", dataLen, "is_gzip", isGzip, "st", st)
}

func alarmsOneForStrategy(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqAlarmCount.Incr(1)
	isGzip := (r.FormValue("gzip") != "")
	st := r.FormValue("strategy_id")
	if st == "" {
		httputil.Response404(rw, r)
		return
	}
	sid, err := strconv.Atoi(st)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	data := sqldata.AlarmsAll()
	if data == nil {
		httputil.Response404(rw, r)
		return
	}
	forSt := data.ForStrategies[sid]
	if forSt == nil {
		httputil.Response404(rw, r)
		return
	}
	dataLen, err := httputil.ResponseJSON(rw, forSt, isGzip, true)
	if err != nil {
		httputil.ResponseFail(rw, r, err)
		return
	}
	log.Debug("req-alarm-for-strategy", "r", r.RemoteAddr, "bytes", dataLen, "is_gzip", isGzip, "st", st)
}

func alarmsTeam(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqAlarmCount.Incr(1)
	teamName := r.FormValue("name")
	if teamName == "" {
		httputil.Response404(rw, r)
		return
	}
	team, status := sqldata.AlarmTeamBy("name", teamName)
	if team == nil {
		httputil.Response404(rw, r)
		return
	}
	httputil.ResponseJSON(rw, map[string]interface{}{
		"team":  team,
		"users": status,
	}, false, false)
	log.Debug("req-alarm-team", "r", r.RemoteAddr, "team", team)
}
