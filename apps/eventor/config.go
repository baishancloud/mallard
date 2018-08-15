package main

type config struct {
	CenterAddr string `json:"center_addr,omitempty"`
	Debug      bool   `json:"debug,omitempty"`
	HTTPAddr   string `json:"http_addr,omitempty"`

	RedisAddr        string `json:"redis_addr,omitempty"`
	RedisQueueDb     int    `json:"redis_queue_db,omitempty"`
	RedisCacheDb     int    `json:"redis_cache_db,omitempty"`
	RedisPassword    string `json:"redis_password,omitempty"`
	RedisQueueLayout string `json:"redis_queue_layout,omitempty"`

	PerfFile string `json:"perf_file,omitempty"`
}

func defaultConfig() config {
	return config{
		CenterAddr:       "http://127.0.0.1:10999",
		Debug:            true,
		HTTPAddr:         "0.0.0.0:10799",
		RedisAddr:        "127.0.0.1:6379",
		RedisQueueDb:     0,
		RedisCacheDb:     1,
		RedisPassword:    "",
		RedisQueueLayout: "md2eventp%d",
		PerfFile:         "perf.log",
	}
}
