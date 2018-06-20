package main

type config struct {
	StoreDir string `json:"store_dir,omitempty"`
	HTTPAddr string `json:"http_addr,omitempty"`
	PerfFile string `json:"perf_file,omitempty"`
	Debug    bool   `json:"debug,omitempty"`
}

func defaultConfig() config {
	return config{
		StoreDir: "./datastore",
		HTTPAddr: "0.0.0.0:10988",
		PerfFile: "perf.json",
		Debug:    true,
	}
}
