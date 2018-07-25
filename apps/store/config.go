package main

type (
	// InfluxOption is option of one influxdb
	InfluxOption struct {
		URLs      map[string]string `json:"urls,omitempty"`
		Db        string            `json:"db,omitempty"`
		User      string            `json:"user,omitempty"`
		Password  string            `json:"password,omitempty"`
		Blacklist []string          `json:"blacklist,omitempty"`
		WhiteList []string          `json:"whitelist,omitempty"`
	}
	// Influx is cluster of influxdbs
	Influx map[string]InfluxOption
	// Transfer is transfer urls
	Transfer struct {
		URLs           map[string]string `json:"urls,omitempty"`
		PullConcurrent int               `json:"pull_concurrent,omitempty"`
	}
	// Config is all config
	Config struct {
		Debug            bool   `json:"debug,omitempty"`
		PerfFile         string `json:"perf_file,omitempty"`
		StatPullerFile   string `json:"stat_puller_file,omitempty"`
		StatInfluxdbFile string `json:"stat_influxdb_file,omitempty"`
	}
)

func defaultConfig() (Config, Transfer, Influx) {
	return Config{
			PerfFile: "performance.json",
			Debug:    true,
		}, Transfer{
			PullConcurrent: 2,
		}, Influx{}
}
