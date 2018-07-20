package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/compute/multijudge"
	"github.com/baishancloud/mallard/componentlib/dataflow/judgestore"
	"github.com/baishancloud/mallard/componentlib/dataflow/judgestore/filter"
	"github.com/baishancloud/mallard/componentlib/httphandler/judgehandler"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	version       = "2.5.0"
	configFile    = "config.json"
	filterCfgFile = "filter.json"
	cfg           = defaultConfig()
	log           = zaplog.Zap("judge")
)

func prepare() {
	osutil.Flags(version, BuildTime, cfg)
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	if err := utils.ReadConfigFile(configFile, &cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
	log.SetDebug(cfg.Debug)
}

func main() {
	prepare()

	queue := make(chan []*models.Metric, 1e5)

	judgestore.SetDir(cfg.StoreDir)
	go filter.SyncFile(filterCfgFile, func(f filter.ForMetrics) {
		judgestore.SetFilters(f)
	}, time.Minute)
	judgestore.RunClean()

	judgehandler.SetQueue(queue)
	go httputil.Listen(cfg.HTTPAddr, judgehandler.Create())

	multijudge.SetCachedEventsFile("cache_events.dump")
	multijudge.RegisterFn(judgestore.WriteMetrics, multijudge.Judge)
	go multijudge.Process(queue)
	go multijudge.ScanForEvents(time.Second * 20)
	go multijudge.ScanStrategies("multi_strategies.json")

	go expvar.PrintAlways("mallard2_judge_perf", cfg.PerfFile, time.Minute)

	osutil.Wait()

	httputil.Close()
	judgestore.Close()
	log.Sync()
}
