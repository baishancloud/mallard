package main

type (
	center struct {
		Addr     string `json:"addr"`
		Interval int    `json:"interval"`
	}
	transferConfig struct {
		URLs []string          `json:"urls"`
		APIs map[string]string `json:"apis"`
	}
	judgeConfig struct {
		FilterFile   string `json:"filter_file"`
		DumpFile     string `json:"dump_file"`
		ScanInterval int    `json:"scan_interval"`
		StoreDir     string `json:"store_dir"`
	}
	config struct {
		Transfer transferConfig `json:"transfer"`
		Center   center         `json:"center"`
		Judge    judgeConfig    `json:"judge"`
		HTTPAddr string         `json:"http_addr"`
		PerfFile string         `json:"perf_file"`
		Debug    bool           `json:"debug"`
	}
)

func defaultConfig() config {
	return config{
		Transfer: transferConfig{
			APIs: map[string]string{
				"metric": "/api/metric",
				"event":  "/api/event",
				"config": "/api/config",
				"self":   "/api/selfinfo",
			},
			URLs: []string{
				"http://127.0.0.1:10899",
			},
		},
		Judge: judgeConfig{
			FilterFile:   "filter.json",
			ScanInterval: 30,
			DumpFile:     "./datalogs/events_dump.log",
			StoreDir:     "./datastore",
		},
		HTTPAddr: "0.0.0.0:10988",
		Center: center{
			Addr:     "http://127.0.0.1:10999",
			Interval: 30,
		},
		PerfFile: "performance.json",
		Debug:    true,
	}
}
