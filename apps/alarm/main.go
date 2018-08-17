package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/alarm/alertdata"
	"github.com/baishancloud/mallard/componentlib/alarm/alertprocess"
	"github.com/baishancloud/mallard/componentlib/alarm/msggcall"
	"github.com/baishancloud/mallard/componentlib/eventor/redisdata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/configapi"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"net/http"
	_ "net/http/pprof"
)

var (
	version    = "2.5.3"
	configFile = "config.json"
	cfg        = defaultConfig()
	log        = zaplog.Zap("alarm")
)

func prepare() {
	osutil.Flags(version, BuildTime, cfg)
	runtime.GOMAXPROCS(runtime.NumCPU() / 4)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	if err := utils.ReadConfigFile(configFile, &cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
	log.SetDebug(cfg.Debug)
}

func main() {
	prepare()

	redisCli, err := initRedis(cfg.Redis.Addr, "", 0, time.Second*10)
	if err != nil {
		log.Fatal("init-redis-fail", "error", err)
	}

	configapi.SetForInterval(configapi.IntervalOption{
		Types: []string{configapi.TypeAlarmsRaw, configapi.TypeAlarmRequests, configapi.TypeSyncHostService},
		Addr:  cfg.Center.Addr,
		Role:  "mallard2-alarm",
		Service: &models.HostService{
			Hostname:       utils.HostName(),
			IP:             utils.LocalIP(),
			ServiceName:    "mallard2-alarm",
			ServiceVersion: version,
			ServiceBuild:   BuildTime,
		},
	})
	go configapi.Intervals(time.Second * 30)

	// set msgg
	msggcall.SetFiles(cfg.Msgg.CommandFile, cfg.Msgg.ActionFile, cfg.Msgg.File)
	msggcall.SetFileway(cfg.MsggFileway.File, cfg.MsggFileway.Layout, cfg.MsggFileway.Expire)
	go msggcall.ScanRequests(time.Second*time.Duration(cfg.Msgg.Ticker), cfg.Msgg.MergeLevel, cfg.Msgg.MergeSize)
	go msggcall.SyncExpvars("mallard2_alarm_msgg_users", time.Hour, cfg.Stats.MsggUserFile)

	// set redis data
	redisdata.SetClient(redisCli, nil)
	redisdata.SetAlarms(cfg.Redis.LowQueues, cfg.Redis.HighQueues, cfg.Redis.SubscribeKey)
	eventCh := make(chan redisdata.EventRecord, 1e4)
	go redisdata.Pop(eventCh, time.Second)

	alertprocess.Register(redisdata.Alert, msggcall.Call)

	if cfg.DbDSN != "" {
		db, err := initDB(cfg.DbDSN)
		if err != nil {
			log.Fatal("init-db-fail", "error", err)
		}
		alertdata.SetDB(db)
		alertdata.ReadProblems(cfg.ProblemsDumpFile)
		go alertdata.StreamAlert()
		alertprocess.Register(alertdata.Alert)

		alertdata.SetStats("mallard2_alarm_stat", cfg.Stats.DumpFile)
		go alertdata.ScanStat(time.Minute, cfg.Stats.MetricDuration, cfg.Stats.MetricFile)
	} else {
		log.Info("db-disabled")
	}

	alertprocess.Process(eventCh)

	go expvar.PrintAlways("mallard2_alarm_perf", cfg.PerfFile, time.Minute)

	if cfg.Debug {
		// run pprof when set debug
		log.Info("start-pprof")
		go http.ListenAndServe("127.0.0.1:49999", nil)
	}

	osutil.Wait()

	redisdata.StopPop()
	redisCli.Close()

	alertdata.DumpProblems(cfg.ProblemsDumpFile)

	log.Sync()
}

func initRedis(address, pwd string, db int, timeout time.Duration) (*redis.Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     pwd,
		DB:           db,
		DialTimeout:  timeout,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	})
	return cli, cli.Ping().Err()
}

func initDB(dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	d, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = d.Ping(); err != nil {
		return nil, err
	}
	return d, nil
}
