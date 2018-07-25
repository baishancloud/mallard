package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/httpserver"
	"github.com/baishancloud/mallard/componentlib/agent/judger"
	"github.com/baishancloud/mallard/componentlib/agent/logutil"
	"github.com/baishancloud/mallard/componentlib/agent/plugins"
	"github.com/baishancloud/mallard/componentlib/agent/processor"
	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/componentlib/agent/syscollector"
	"github.com/baishancloud/mallard/componentlib/agent/transfer"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	version    = "2.5.0"
	configFile = "config.json"
	cfg        = defaultConfig()
	log        = zaplog.Zap("agent")
)

func prepare() {
	osutil.Flags(version, BuildTime, cfg)
	runtime.GOMAXPROCS(cfg.Core)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	if err := utils.ReadConfigFile(configFile, cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
	if err := checkConfig(cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
	log.SetDebug(cfg.Debug)
}

func main() {
	prepare()

	serverinfo.UseAllConf = cfg.UseAllConf
	serverinfo.Scan(cfg.Endpoint)

	metricsQueue := make(chan []*models.Metric, 2e4)
	eventsQueue := make(chan []*models.Event, 2e4)
	errorQueue := make(chan error, 2e3)

	transfer.SetURLs(cfg.Transfer.FullURLs(serverinfo.Hostname()), cfg.Transfer.APIs)
	configSyncOpt := transfer.SyncOption{
		Interval:  time.Second * time.Duration(cfg.Transfer.ConfigInterval),
		Version:   version,
		BuildTime: BuildTime,
	}
	configSyncOpt.Func = func(epData *models.EndpointData, isUpdate bool) {
		if isUpdate && epData.Config != nil {
			if !cfg.DisableJudge {
				judger.SetStrategyData(epData.Config.Strategies)
			}
			plugins.SetDir(cfg.Plugin.Dir, cfg.Plugin.LogDir, epData.Config.Plugins)
		}
		if epData.Time > 0 {
			syscollector.SetSystime(epData.Time)
			log.Info("set-systime", "time", epData.Time, "now", time.Now().Unix())
		}
	}
	go transfer.SyncConfig(configSyncOpt)
	go transfer.SyncSelfInfo(cfg)

	var judgeFn = func(metrics []*models.Metric) {
		if cfg.DisableJudge {
			return
		}
		events := judger.Judge(metrics)
		if len(events) > 0 {
			eventsQueue <- events
		}
	}

	processor.Register(transfer.Metrics, judgeFn, logutil.Write)
	processor.RegisterEvent(transfer.Events)

	go processor.Process(metricsQueue, eventsQueue, errorQueue)

	go syscollector.Collect(cfg.Collector.Prefix,
		time.Second*time.Duration(cfg.Collector.Interval),
		metricsQueue, errorQueue)

	go plugins.Exec(metricsQueue)
	go plugins.SyncScan(time.Minute)

	httpserver.SetQueue(metricsQueue)
	go httpserver.Listen(cfg.Server.Addr)

	logutil.SetReadDir(cfg.Logutil.ReadDir)
	logutil.SetWriteFile(cfg.Logutil.WriteFile, cfg.Logutil.CleanDays, cfg.Logutil.GzipDays)
	go logutil.ReadInterval(time.Second*time.Duration(cfg.Logutil.ReadInterval), metricsQueue)

	go expvar.PrintAlways("mallard2_agent_perf", cfg.PerfFile, time.Minute*2)

	osutil.Wait()

	syscollector.StopCollect()
	transfer.Stop()
	logutil.Stop()
	log.Sync()
}
