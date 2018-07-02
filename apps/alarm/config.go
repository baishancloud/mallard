package main

type config struct {
	CenterAddr string `json:"center_addr"`
	RedisAddr  string `json:"redis_addr"`
	DbDSN      string `json:"db_dsn"`

	EtcdAddr     []string `json:"etcd_addr"`
	EtcdUser     string   `json:"etcd_user"`
	EtcdPassword string   `json:"etcd_password"`

	HighQueues []string `json:"high_queues"`
	LowQueues  []string `json:"low_queues"`

	AlarmsDumpFile    string `json:"alarms_dump_file"`
	AlarmSubscribeKey string `json:"alarm_subscribe_key"`

	CommandFile string `json:"command_file"`
	ActionFile  string `json:"action_file"`
	MsggFile    string `json:"msgg_file"`

	PerfFile string `json:"perf_file"`
	Debug    bool   `json:"debug"`

	StatMetricFile string `json:"stat_metric_file"`
}

func defaultConfig() config {
	return config{
		CenterAddr:        "http://127.0.0.1:10999",
		RedisAddr:         "127.0.0.1:6379",
		DbDSN:             "",
		AlarmsDumpFile:    "problems.log",
		AlarmSubscribeKey: "",
		LowQueues:         []string{},
		HighQueues:        []string{},
		CommandFile:       "",
		ActionFile:        "",
		MsggFile:          "",
		Debug:             true,
	}
}
