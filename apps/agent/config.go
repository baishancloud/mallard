package main

import (
	"errors"
	"sort"
	"strings"

	"github.com/baishancloud/mallard/corelib/utils"
)

type (
	collector struct {
		Interval int    `json:"interval"`
		Prefix   string `json:"prefix"`
	}
	transferConfig struct {
		URLs           []string            `json:"urls"`
		APIs           map[string]string   `json:"apis"`
		Addon          map[string][]string `json:"addon"`
		ConfigInterval int                 `json:"config_interval"`
	}
	plugin struct {
		Dir    string `json:"dir"`
		LogDir string `json:"log_dir"`
		Reload int    `json:"reload"`
	}
	logopt struct {
		ReadInterval int    `json:"read_interval"`
		ReadDir      string `json:"read_dir"`
		WriteFile    string `json:"write_file"`
		CleanDays    int    `json:"clean_days"`
		GzipDays     int    `json:"gzip_days"`
	}
	server struct {
		Addr string `json:"addr"`
	}
	config struct {
		Debug        bool           `json:"debug"`
		Endpoint     string         `json:"endpoint"`
		Core         int            `json:"core"`
		Server       server         `json:"server"`
		Transfer     transferConfig `json:"transfer,omitempty"`
		Collector    collector      `json:"collector"`
		Plugin       plugin         `json:"plugin"`
		DisableJudge bool           `json:"disable_judge"`
		Logutil      logopt         `json:"logutil"`
		PerfFile     string         `json:"perf_file"`
		UseAllConf   bool           `json:"use_allconf"`
	}
)

func (tfr transferConfig) FullURLs(endpoint string) []string {
	u := make([]string, len(tfr.URLs))
	copy(u, tfr.URLs)
	for k, list := range tfr.Addon {
		klist := strings.Split(k, ",")
		for _, k := range klist {
			if k == "" {
				continue
			}
			if strings.HasPrefix(endpoint, k) {
				u = append(u, list...)
			}
		}
	}
	u = utils.StringSliceUnique(u)
	sort.Sort(sort.StringSlice(u))
	return u
}

func defaultConfig() *config {
	return &config{
		Debug:    true,
		Endpoint: "",
		Core:     4,
		Server: server{
			Addr: "127.0.0.1:10699",
		},
		Transfer: transferConfig{
			APIs: map[string]string{
				"metric": "/api/metric",
				"event":  "/api/event",
				"config": "/api/config",
				"self":   "/api/selfinfo",
			},
			Addon: make(map[string][]string),
			URLs: []string{
				"http://127.0.0.1:10899",
			},
			ConfigInterval: 30,
		},
		Collector: collector{
			Interval: 60,
			Prefix:   "sys",
		},
		Plugin: plugin{
			Dir:    "./plugins",
			LogDir: "./plugins_log",
			Reload: 30,
		},
		Logutil: logopt{
			ReadInterval: 5,
			ReadDir:      "./datalogs",
			WriteFile:    "./var/metrics_%s.json",
			CleanDays:    4,
			GzipDays:     2,
		},
		DisableJudge: false,
		UseAllConf:   false,
		PerfFile:     "./datalogs/mallard2_agent.log",
	}
}

func checkConfig(cfg *config) error {
	if cfg.Server.Addr == "" {
		return errors.New("need-http-addr")
	}
	if cfg.Collector.Interval == 0 {
		return errors.New("need-sys-collect-over-0")
	}
	if len(cfg.Transfer.FullURLs("")) == 0 {
		return errors.New("need-transfer-urls")
	}
	return nil
}
