package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/dataflow/influxdb"
	"github.com/baishancloud/mallard/componentlib/dataflow/puller"
	"github.com/baishancloud/mallard/corelib/container"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"

	"net/http"
	_ "net/http/pprof"
)

var (
	version                     = "2.5.0"
	configFile                  = "config.json"
	transferConfigFile          = "config_transfer.json"
	influxConfigFile            = "config_influxdb.json"
	cfg, transferCfg, influxCfg = defaultConfig()
	log                         = zaplog.Zap("store")
)

func prepare() {
	osutil.Flags(version, BuildTime, nil)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	//utils.WriteConfigFile(configFile, cfg)
	if err := utils.ReadConfigFile(configFile, &cfg); err != nil {
		log.Fatal("config-error", "error", err, "file", configFile)
	}
	//utils.WriteConfigFile(transferConfigFile, transferCfg)
	if err := utils.ReadConfigFile(transferConfigFile, &transferCfg); err != nil {
		log.Fatal("config-error", "error", err, "file", transferConfigFile)
	}
	//utils.WriteConfigFile(influxConfigFile, influxCfg)
	if err := utils.ReadConfigFile(influxConfigFile, &influxCfg); err != nil {
		log.Fatal("config-error", "error", err, "file", influxConfigFile)
	}

	log.SetDebug(cfg.Debug)
}

func main() {
	prepare()

	cOpt := make(map[string]influxdb.GroupOption, len(influxCfg))
	for key, opt := range influxCfg {
		cOpt[key] = influxdb.GroupOption{
			URLs:      opt.URLs,
			Db:        opt.Db,
			User:      opt.User,
			Password:  opt.Password,
			Blacklist: opt.Blacklist,
			WhiteList: opt.WhiteList,
		}
	}

	go http.ListenAndServe("127.0.0.1:49999", nil)

	queue := container.NewLimitedList(1e7)

	influxdb.SetCluster(cOpt)
	go influxdb.Process(queue)
	go influxdb.SyncExpvars(time.Minute, cfg.StatInfluxdbFile)

	puller.SetQueue(queue)
	puller.SetURLs(transferCfg.URLs, transferCfg.PullConcurrent)
	go puller.SyncExpvars(time.Minute, cfg.StatPullerFile)

	go expvar.PrintAlways("mallard2_store_perf", cfg.PerfFile, time.Minute)

	osutil.Wait()

	puller.Stop()
	influxdb.Stop()

	log.Sync()

}
