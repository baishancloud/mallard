package main

type config struct {
	Debug       bool              `json:"debug,omitempty"`
	CenterAddr  string            `json:"center_addr,omitempty"`
	HTTPAddr    string            `json:"http_addr,omitempty"`
	TokenFile   string            `json:"token_file,omitempty"`
	IsPublic    bool              `json:"is_public,omitempty"`
	EventorAddr map[string]string `json:"eventor_addr,omitempty"`
	PerfFile    string            `json:"perf_file,omitempty"`
}

func defaultConfig() config {
	return config{
		Debug:       true,
		CenterAddr:  "http://127.0.0.1:10999",
		EventorAddr: nil,
		HTTPAddr:    "0.0.0.0:10899",
		TokenFile:   "tokens.json",
		IsPublic:    false,
	}
}
