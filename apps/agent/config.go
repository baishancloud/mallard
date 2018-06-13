package main

import (
	"errors"
	"sort"
	"strings"

	"github.com/baishancloud/mallard/corelib/utils"
)

type (
	transferConfig struct {
		URLs  []string            `json:"urls,omitempty"`
		APIs  map[string]string   `json:"apis,omitempty"`
		Addon map[string][]string `json:"addon,omitempty"`
	}
	config struct {
		Debug         bool           `json:"debug,omitempty"`
		Endpoint      string         `json:"endpoint,omitempty"`
		Core          int            `json:"core,omitempty"`
		HTTPAddr      string         `json:"http_addr,omitempty"`
		Transfer      transferConfig `json:"transfer,omitempty"`
		SysCollect    int            `json:"sys_collect,omitempty"`
		SysPrefix     string         `json:"sys_prefix,omitempty"`
		PluginsDir    string         `json:"plugins_dir,omitempty"`
		PluginsLogDir string         `json:"plugins_log_dir,omitempty"`
		DisableJudge  bool           `json:"disable_judge,omitempty"`
		PerfFile      string         `json:"perf_file,omitempty"`
		ReadLogDir    string         `json:"read_log_dir,omitempty"`
		WriteLogFile  string         `json:"write_log_file,omitempty"`
	}
)

func (tfr transferConfig) FullURLs(endpoint string) []string {
	u := make([]string, len(tfr.URLs))
	copy(u, tfr.URLs)
	for k, list := range tfr.Addon {
		if strings.HasPrefix(endpoint, k) {
			u = append(u, list...)
		}
	}
	u = utils.StringSliceUnique(u)
	sort.Sort(sort.StringSlice(u))
	return u
}

func defaultConfig() *config {
	return &config{
		Debug:      true,
		Core:       2,
		HTTPAddr:   "127.0.0.1:10699",
		SysCollect: 60,
		Transfer: transferConfig{
			APIs: map[string]string{
				"metric": "/api/metric",
				"event":  "/api/event",
				"config": "/api/config",
				"self":   "/api/selfinfo",
			},
		},
		DisableJudge:  false,
		PluginsDir:    "plugins",
		PluginsLogDir: "",
		PerfFile:      "./datalogs/mallard2_agent.log",
		ReadLogDir:    "./datalogs",
		WriteLogFile:  "./var/metrics.log",
		SysPrefix:     "",
	}
}

func checkConfig(cfg *config) error {
	if cfg.HTTPAddr == "" {
		return errors.New("need-http-addr")
	}
	if cfg.SysCollect == 0 {
		return errors.New("need-sys-collect-over-0")
	}
	if len(cfg.Transfer.FullURLs("")) == 0 {
		return errors.New("need-transfer-urls")
	}
	return nil
}
