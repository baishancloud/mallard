package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/transfer/eventsender"
	"github.com/baishancloud/mallard/componentlib/transfer/queues"
	"github.com/baishancloud/mallard/componentlib/transfer/transferhandler"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httptoken"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/baishancloud/mallard/extralib/configapi"
)

var (
	version    = "2.5.0"
	configFile = "config.json"
	cfg        = defaultConfig()
	log        = zaplog.Zap("transfer")
)

func prepare() {
	osutil.Flags(version, BuildTime, cfg)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	if err := utils.ReadConfigFile(configFile, &cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
	log.SetDebug(cfg.Debug)
}

func main() {

	prepare()

	// set center
	configapi.SetForInterval(configapi.IntervalOption{
		Addr: cfg.Center.Addr,
		Types: []string{
			configapi.TypeEndpoints,
			configapi.TypeHostInfos,
			configapi.TypeSyncHeartbeat,
			configapi.TypeSyncHostService,
			configapi.TypeHostConfigs,
		},
		Service: &models.HostService{
			Hostname:       utils.HostName(),
			IP:             utils.LocalIP(),
			ServiceName:    "mallard2-transfer",
			ServiceVersion: version,
			ServiceBuild:   BuildTime,
		},
	})
	go configapi.Intervals(time.Second * time.Duration(cfg.Center.Interval))

	// set token
	go httptoken.SyncVerifier(cfg.TokenFile, time.Minute)

	// prepare queues
	mQueue := queues.NewQueue(1e6, cfg.Store.DumpDir)
	go mQueue.ScanDump(time.Minute, func(res queues.ScanDumpResult) {
		log.Info("read-metrics-dump", "dump", res)
	})
	evtQueue := queues.NewQueue(1e6, cfg.Eventor.DumpDir)
	go evtQueue.ScanDump(time.Minute, func(res queues.ScanDumpResult) {
		log.Info("read-events-dump", "dump", res)
	})

	// init event-sender
	eventsender.SetURLs(cfg.Eventor.Addrs)
	go eventsender.ProcessQueue(evtQueue, 200, time.Second*time.Duration(cfg.Eventor.Timeout))

	// init http server
	log.Info("set-http", "is_public", cfg.IsPublic, "is_authorized", cfg.IsAuthorized)
	transferhandler.SetQueues(mQueue, evtQueue)
	go httputil.Listen(cfg.HTTPAddr, transferhandler.Create(cfg.IsPublic, cfg.IsAuthorized), cfg.CertFile, cfg.KeyFile)

	go expvar.PrintAlways("mallard2_transfer_perf", cfg.PerfFile, time.Minute)

	osutil.Wait()

	httputil.Close()
	eventsender.Stop()
	dump(mQueue, evtQueue)
	log.Sync()
}

func dump(mQueue, evtQueue *queues.Queue) {
	file, count, err := mQueue.Dump(1e6 * 2)
	if err != nil {
		log.Warn("metrics-dump-error", "error", err)
	} else {
		log.Info("metrics-dump", "file", file, "count", count)
	}
	file, count, err = evtQueue.Dump(1e6 * 2)
	if err != nil {
		log.Warn("events-dump-error", "error", err)
	} else {
		log.Info("events-dump", "file", file, "count", count)
	}
}
