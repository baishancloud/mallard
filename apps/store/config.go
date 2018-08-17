package main

type (
	// InfluxOption is option of one influxdb
	InfluxOption struct {
		URLs      map[string]string `json:"urls"`
		Db        string            `json:"db"`
		User      string            `json:"user"`
		Password  string            `json:"password"`
		Blacklist []string          `json:"blacklist"`
		WhiteList []string          `json:"whitelist"`
	}
	// Influx is cluster of influxdbs
	Influx map[string]InfluxOption
	// Transfer is transfer urls
	Transfer struct {
		URLs           map[string]string `json:"urls"`
		PullConcurrent int               `json:"pull_concurrent"`
	}
	// Config is all config
	Config struct {
		Debug            bool   `json:"debug"`
		PerfFile         string `json:"perf_file"`
		StatPullerFile   string `json:"stat_puller_file"`
		StatInfluxdbFile string `json:"stat_influxdb_file"`
	}
)

func defaultConfig() (Config, Transfer, Influx) {
	return Config{
			PerfFile:         "performance.json",
			StatInfluxdbFile: "stat_influxdb.json",
			StatPullerFile:   "stat_puller.json",
			Debug:            true,
		}, Transfer{
			URLs:           make(map[string]string),
			PullConcurrent: 2,
		}, Influx{}
}
