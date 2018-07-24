package main

type (
	center struct {
		Addr     string `json:"addr"`
		Interval int    `json:"interval"`
	}
	config struct {
		StoreDir string `json:"store_dir"`
		Center   center `json:"center"`
		HTTPAddr string `json:"http_addr"`
		PerfFile string `json:"perf_file"`
		Debug    bool   `json:"debug"`
	}
)

func defaultConfig() config {
	return config{
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
