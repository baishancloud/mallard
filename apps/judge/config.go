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
	config struct {
		Transfer transferConfig `json:"transfer"`
		StoreDir string         `json:"store_dir"`
		Center   center         `json:"center"`
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
		StoreDir: "./datastore",
		HTTPAddr: "0.0.0.0:10988",
		Center: center{
			Addr:     "http://127.0.0.1:10999",
			Interval: 20,
		},
		PerfFile: "performance.json",
		Debug:    true,
	}
}
