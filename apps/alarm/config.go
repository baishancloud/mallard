package main

type config struct {
	CenterAddr string `json:"center_addr,omitempty"`
	RedisAddr  string `json:"redis_addr,omitempty"`
	DbDSN      string `json:"db_dsn,omitempty"`

	EtcdAddr     []string `json:"etcd_addr,omitempty"`
	EtcdUser     string   `json:"etcd_user,omitempty"`
	EtcdPassword string   `json:"etcd_password,omitempty"`

	HighQueues []string `json:"high_queues,omitempty"`
	LowQueues  []string `json:"low_queues,omitempty"`

	AlarmsDumpFile    string `json:"alarms_dump_file,omitempty"`
	AlarmSubscribeKey string `json:"alarm_subscribe_key,omitempty"`

	CommandFile string `json:"command_file,omitempty"`
	ActionFile  string `json:"action_file,omitempty"`
	MsggFile    string `json:"msgg_file,omitempty"`

	PerfFile string `json:"perf_file,omitempty"`
	Debug    bool   `json:"debug,omitempty"`

	StatMetricFile string `json:"stat_metric_file,omitempty"`
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
