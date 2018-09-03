package main

type (
	centerConfig struct {
		Addr     string `json:"addr"`
		Interval int    `json:"interval"`
	}
	eventorConfig struct {
		Addrs   map[string]string `json:"addrs"`
		DumpDir string            `json:"dump_dir"`
		Timeout int               `json:"timeout"`
	}
	metricConfig struct {
		DumpDir string `json:"dump_dir"`
	}
	config struct {
		Debug        bool          `json:"debug"`
		Center       centerConfig  `json:"center"`
		Eventor      eventorConfig `json:"eventor"`
		Store        metricConfig  `json:"store"`
		HTTPAddr     string        `json:"http_addr"`
		KeyFile      string        `json:"key_file"`
		CertFile     string        `json:"cert_file"`
		TokenFile    string        `json:"token_file"`
		IsPublic     bool          `json:"is_public"`
		IsAuthorized bool          `json:"is_authorized"`
		PerfFile     string        `json:"perf_file"`
	}
)

func defaultConfig() config {
	return config{
		Debug: true,
		Center: centerConfig{
			Addr:     "http://127.0.0.1:10999",
			Interval: 30,
		},
		Eventor: eventorConfig{
			Addrs:   make(map[string]string),
			DumpDir: "_dump/event",
			Timeout: 20,
		},
		Store: metricConfig{
			DumpDir: "_dump/metric",
		},
		HTTPAddr:     "0.0.0.0:10899",
		TokenFile:    "tokens.json",
		IsPublic:     false,
		IsAuthorized: true,
	}
}
