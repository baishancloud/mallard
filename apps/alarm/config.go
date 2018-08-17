package main

type (
	centerConfig struct {
		Addr     string `json:"addr"`
		Interval int    `json:"interval"`
	}
	redisConfig struct {
		Addr         string   `json:"addr"`
		LowQueues    []string `json:"low_queues"`
		HighQueues   []string `json:"high_queues"`
		SubscribeKey string   `json:"subscribe_key"`
	}
	msggConfig struct {
		CommandFile string `json:"command_file"`
		ActionFile  string `json:"action_file"`
		File        string `json:"file"`
		MergeSize   int    `json:"merge_size"`
		MergeLevel  int    `json:"merge_level"`
		Ticker      int    `json:"ticker"`
	}
	msggFileway struct {
		File   string `json:"file"`
		Expire int64  `json:"expire"`
		Layout string `json:"layout"`
	}
	statsConfig struct {
		DumpFile       string `json:"dump_file"`
		MetricDuration int64  `json:"metric_duration"`
		MetricFile     string `json:"metric_file"`
		MsggUserFile   string `json:"msgg_user_file"`
	}
)
type config struct {
	Center centerConfig `json:"center"`
	Redis  redisConfig  `json:"redis"`
	DbDSN  string       `json:"db_dsn"`

	ProblemsDumpFile string `json:"problems_dump_file"`

	Msgg        msggConfig  `json:"msgg"`
	MsggFileway msggFileway `json:"msgg_fileway"`

	Stats statsConfig `json:"stats"`

	PerfFile string `json:"perf_file"`
	Debug    bool   `json:"debug"`
}

func defaultConfig() config {
	return config{
		Center: centerConfig{
			Addr:     "http://127.0.0.1:10999",
			Interval: 30,
		},
		Redis: redisConfig{
			Addr:         "127.0.0.1:6379",
			SubscribeKey: "alarms_subscribe",
			LowQueues:    []string{"event5", "event6"},
			HighQueues:   []string{"event1", "event2", "event3", "event4"},
		},
		DbDSN:            "root:@tcp(127.0.0.1:3306)/clerk?timeout=10s",
		ProblemsDumpFile: "datalogs/problems.log",

		Msgg: msggConfig{
			ActionFile:  "action.sh",
			CommandFile: "command.sh",
			File:        "msgg.sh",
			MergeLevel:  3,
			MergeSize:   5,
			Ticker:      5,
		},
		MsggFileway: msggFileway{
			File:   "msgg_fileway.py",
			Layout: "datalogs/alarms_%s",
			Expire: 1800,
		},

		Debug:    true,
		PerfFile: "performance.json",
		Stats: statsConfig{
			DumpFile:       "datalogs/stats.log",
			MetricDuration: 1800,
			MetricFile:     "performance_stats.json",
			MsggUserFile:   "performance_users.json",
		},
	}
}
