package redisdata

var (
	alertSubscribeQueue string
)

// SetAlertSubscribe sets redis subscribe queue name
func SetAlertSubscribe(name string) {
	alertSubscribeQueue = name
}

// ToSubscrible adds eid to subscribe
func ToSubscrible(eid string) {
	if alertSubscribeQueue == "" {
		return
	}
	queueCli.LPush(alertSubscribeQueue, eid)
	llen := queueCli.LLen(alertSubscribeQueue).Val()
	log.Debug("lpush", "queue", alertSubscribeQueue, "eid", eid, "llen", llen)
	if llen > 1e4 {
		queueCli.LTrim(alertSubscribeQueue, 50, -1)
		log.Debug("ltrim", "queue", alertSubscribeQueue)
	}
}
