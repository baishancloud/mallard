package main

type config struct {
	PortalDSN      string `json:"portal_dsn,omitempty"`
	UicDSN         string `json:"uic_dsn,omitempty"`
	Debug          bool   `json:"debug,omitempty"`
	HTTPAddr       string `json:"http_addr,omitempty"`
	PerfFile       string `json:"perf_file,omitempty"`
	ReloadInterval int    `json:"reload_interval,omitempty"`

	EtcdAddr     []string `json:"etcd_addr,omitempty"`
	EtcdUser     string   `json:"etcd_user,omitempty"`
	EtcdPassword string   `json:"etcd_password,omitempty"`
}

func defaultConfig() config {
	return config{
		Debug:          true,
		HTTPAddr:       "127.0.0.1:10999",
		PerfFile:       "center.json",
		ReloadInterval: 10,
	}
}
