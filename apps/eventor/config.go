package main

type centerConfig struct {
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
}

type redisConfig struct {
	Addr        string `json:"addr"`
	Password    string `json:"password"`
	QueueDB     int    `json:"queue_db"`
	CacheDB     int    `json:"cache_db"`
	QueueLayout string `json:"queue_layout"`
}

type config struct {
	Center   centerConfig `json:"center"`
	Redis    redisConfig  `json:"redis"`
	HTTPAddr string       `json:"http_addr"`
	Debug    bool         `json:"debug"`
	PerfFile string       `json:"perf_file"`
}

func defaultConfig() config {
	return config{
		Center: centerConfig{
			Addr:     "http://127.0.0.1:10999",
			Interval: 30,
		},
		Redis: redisConfig{
			Addr:        "127.0.0.1:6379",
			Password:    "",
			QueueDB:     0,
			CacheDB:     1,
			QueueLayout: "event%d",
		},
		Debug:    true,
		HTTPAddr: "0.0.0.0:10799",
		PerfFile: "performance.json",
	}
}
