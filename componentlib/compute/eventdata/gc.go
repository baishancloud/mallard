package eventdata

import (
	"strings"
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/redisdata"
	"github.com/baishancloud/mallard/corelib/expvar"
)

var (
	// GCExpire is expire time to gc the value in cache redis db
	GCExpire int64 = 24 * 3600

	gcRactor    = 100
	gcCount     = expvar.NewBase("rd.gc")
	dbSizeCount = expvar.NewBase("rd.db_keys")
)

func init() {
	expvar.Register(gcCount, dbSizeCount)
}

// StartGC starts gc process in time loop
func StartGC(interval time.Duration) {
	time.Sleep(time.Second * 5) // wait to run
	gcRactor = 86400 / int(interval.Seconds())
	log.Debug("gc-ractor", "ractor", gcRactor)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		gcOnce()
	}
}

func gcOnce() {
	size, err := redisdata.CacheDBSize()
	if err != nil {
		log.Warn("gc-dbsize-error", "error", err)
		return
	}
	cli := redisdata.GetCacheCli()
	if cli == nil {
		log.Warn("gc-client-nil")
		return
	}
	keyLen := int(size) / gcRactor
	now := time.Now().Unix()
	var count int
	for i := 0; i < keyLen; i++ {
		key := cli.RandomKey().Val()
		if key == "" {
			continue
		}
		if strings.HasPrefix(key, "alarms") || strings.HasPrefix(key, "nodata") {
			continue
		}
		tUnix, _ := cli.HGet(key, "lastest_time").Int64()
		if tUnix > 0 && now-tUnix >= GCExpire {
			cli.Del(key)
			log.Info("gc-del", "eid", key, "time", time.Unix(tUnix, 0).Format("01-02 15:04:04"))
			count++
		}
	}
	gcCount.Set(int64(count))
	dbSizeCount.Set(size)
	log.Info("gc-ok", "count", count, "keys", keyLen)
}
