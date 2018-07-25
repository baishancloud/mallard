package main

import (
	"runtime"
	"time"

	"github.com/baishancloud/mallard/componentlib/center/centerhandler"
	"github.com/baishancloud/mallard/componentlib/center/sqldata"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/httputil"
	"github.com/baishancloud/mallard/corelib/osutil"
	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	version    = "2.5.1"
	configFile = "config.json"
	cfg        = defaultConfig()
	log        = zaplog.Zap("center")
)

func prepare() {
	osutil.Flags(version, BuildTime, cfg)
	log.SetDebug(cfg.Debug)
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	log.Info("init", "core", runtime.GOMAXPROCS(0), "version", version)

	if err := utils.ReadConfigFile(configFile, &cfg); err != nil {
		log.Fatal("config-error", "error", err)
	}
}

func prepareDB(portalDSN, uicDSN string) (*sqlx.DB, *sqlx.DB, error) {
	pdb, err := sqlx.Open("mysql", portalDSN)
	if err != nil {
		return nil, nil, err
	}
	if err = pdb.Ping(); err != nil {
		return nil, nil, err
	}
	cdb, err := sqlx.Open("mysql", uicDSN)
	if err != nil {
		return nil, nil, err
	}
	if err = cdb.Ping(); err != nil {
		return nil, nil, err
	}
	log.Info("init-db")
	return pdb, cdb, nil
}

func main() {
	prepare()

	pdb, cdb, err := prepareDB(cfg.PortalDSN, cfg.UicDSN)
	if err != nil {
		log.Fatal("db-fail", "error", err)
	}

	sqldata.InitExpvars()
	sqldata.SetDB(pdb, cdb)
	go sqldata.Sync(time.Second*time.Duration(cfg.ReloadInterval), nil)

	go httputil.Listen(cfg.HTTPAddr, centerhandler.Handlers())

	go expvar.PrintAlways("mallard2_center_perf", cfg.PerfFile, time.Minute)

	osutil.Wait()

	httputil.Close()
	pdb.Close()
	cdb.Close()
	log.Sync()
}
