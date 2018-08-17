package main

type config struct {
	PortalDSN      string `json:"portal_dsn,omitempty"`
	UicDSN         string `json:"uic_dsn,omitempty"`
	Debug          bool   `json:"debug,omitempty"`
	HTTPAddr       string `json:"http_addr,omitempty"`
	PerfFile       string `json:"perf_file,omitempty"`
	ReloadInterval int    `json:"reload_interval,omitempty"`
}

func defaultConfig() config {
	return config{
		Debug:          true,
		PortalDSN:      "root:@tcp(127.0.0.1:3306)/portal",
		UicDSN:         "root:@tcp(127.0.0.1:3306)/uic",
		HTTPAddr:       "127.0.0.1:10998",
		PerfFile:       "performance.json",
		ReloadInterval: 30,
	}
}
