package msggcall

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
	"github.com/baishancloud/mallard/componentlib/compute/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/configapi"
)

var (
	commandFile string
	actionFile  string
	msggFile    string
)

// SetFiles sets files
func SetFiles(cmd, action, msgg string) {
	commandFile = cmd
	if _, err := os.Stat(commandFile); err != nil {
		log.Warn("stat-cmd-file-error", "error", err)
		commandFile = ""
	}
	actionFile = action
	if _, err := os.Stat(actionFile); err != nil {
		log.Warn("stat-action-file-error", "error", err)
		actionFile = ""
	}
	msggFile = msgg
	if _, err := os.Stat(msggFile); err != nil {
		log.Warn("stat-msgg-file-error", "error", err)
		msggFile = ""
	}
}

var (
	log = zaplog.Zap("msgg")

	requests     = make(map[string]map[int64]*msggRequest)
	requestsLock sync.RWMutex
)

// Call adds event msgg requests to call
func Call(record redisdata.EventRecord) {
	if !record.IsHigh {
		return
	}
	if record.Event.Status == models.EventOutdated.String() || record.Event.Status == models.EventClosed.String() {
		return
	}
	if err := models.FillEventStrategy(record.Event); err != nil {
		log.Warn("fill-strategy-error", "event", record.Event, "error", err)
		return
	}
	st, ok := record.Event.Strategy.(*models.Strategy)
	if !ok {
		log.Warn("nil-strategy-error", "event", record.Event)
		return
	}

	// fill uic
	uic := "000"
	action := configapi.AlarmActionForStrategy(st.ID)
	if action != nil {
		uic = action.Uic
	}

	if commandFile != "" {
		go CallCommand(record.Event, st.Note, uic)
	}
	if actionFile != "" {
		go CallAction(record.Event, st.Note, uic)
	}
	if msggFile != "" {
		go CallMsgg(record.Event, st)
	}
}

// CallCommand call command script
func CallCommand(event *models.EventFull, note string, uic string) {
	args := []string{event.Endpoint, note, strconv.Itoa(event.Priority()), event.PushedTags["sertypes"], uic, strconv.FormatFloat(event.LeftValue, 'f', 3, 64)}
	output, err := runWithTimeout(commandFile, args, time.Second*30)
	if err != nil {
		log.Warn("call-cmd-error", "error", err, "eid", event.ID)
		return
	}
	log.Info("call-cmd", "eid", event.ID, "status", event.Status, "output", string(output))
}

// CallAction call action script
func CallAction(event *models.EventFull, note string, uic string) {
	pushTagsJSON, _ := json.Marshal(event.PushedTags)
	pushFieldsJSON, _ := json.Marshal(event.Fields)
	args := []string{event.Endpoint, note, strconv.Itoa(event.Priority()), string(pushTagsJSON), uic, strconv.FormatFloat(event.LeftValue, 'f', 3, 64), string(pushFieldsJSON)}
	output, err := runWithTimeout(actionFile, args, time.Second*30)
	if err != nil {
		log.Warn("call-action-error", "error", err, "event", event)
		return
	}
	log.Info("call-action", "eid", event.ID, "status", event.Status, "output", string(output))
}

// CallMsgg call msgg script
func CallMsgg(event *models.EventFull, st *models.Strategy) {
	if event.Status == models.EventOk.String() {
		CleanRequests(event.ID)
		// add recover callback
		if st.RecoverNotify > 0 {
			AddRequests(event, st, true)
		}
		return
	}
	AddRequests(event, st, false)
}

var (
	msggAddCount   = expvar.NewDiff("alert.msgg_add")
	msggCleanCount = expvar.NewDiff("alert.msgg_clean")
)

func init() {
	expvar.Register(msggAddCount, msggCleanCount)
}

// AddRequests add event to msgg request
func AddRequests(event *models.EventFull, st *models.Strategy, isRecover bool) {
	level := 5 - st.Priority
	if level < 0 {
		log.Warn("priority-negative", "eid", event.ID, "status", event.Status, "st", st.ID, "priority", st.Priority)
		return
	}
	stReqs := configapi.AlarmRequestByStrategy(st.ID)
	if len(stReqs) == 0 {
		stReqs = make(map[string]*sqldata.AlarmSendRequest, 1)
		log.Warn("no-reqs", "eid", event.ID)
		// fill empty req
		defaultReq := &sqldata.AlarmSendRequest{
			Step:     0,
			Uic:      "000",
			Template: "000",
		}
		action := configapi.AlarmActionForStrategy(st.ID)
		if action != nil {
			defaultReq.Uic = action.Uic
			tpl := configapi.AlarmTemplateForAction(action.ID)
			if tpl != nil {
				defaultReq.Template = tpl.Name
			}
		}
		stReqs["fill-default"] = defaultReq
	}
	reqs := make(map[int64]*msggRequest, len(stReqs))
	now := time.Now().Unix()
	for _, req := range stReqs {
		t := now + int64(req.Step)
		for {
			if _, ok := reqs[t]; ok {
				t++
				continue
			}
			break
		}
		reqs[t] = &msggRequest{
			SendRequest: req,
			Recover:     isRecover,
			Note:        st.Note,
			LeftValue:   event.LeftValue,
			Endpoint:    event.Endpoint,
			Level:       level,
			Time:        event.EventTime,
			Sertypes:    event.PushedTags["sertypes"],
			Status:      event.Status,
		}
	}
	requestsLock.Lock()
	eid := event.ID
	if len(requests[eid]) == 0 {
		requests[eid] = reqs
		log.Debug("add-reqs", "eid", eid, "reqs", requests[eid])
	} else {
		for t, req := range reqs {
			requests[eid][t] = req
			log.Debug("add-reqs", "eid", eid, "t", t, "req", req)
		}
	}
	msggAddCount.Incr(int64(len(reqs)))
	requestsLock.Unlock()
}

// CleanRequests cleans requests
func CleanRequests(eid string) {
	requestsLock.Lock()

	msggCleanCount.Incr(int64(len(requests[eid])))
	log.Debug("clean-OK", "eid", eid, "reqs", len(requests[eid]))
	delete(requests, eid)

	requestsLock.Unlock()
}

func runWithTimeout(cmd string, args []string, timeout time.Duration) ([]byte, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	command := exec.CommandContext(ctx, cmd, args...)
	output, err := command.Output()
	if err != nil && command.Process != nil {
		command.Process.Kill()
	}
	return output, err
}
