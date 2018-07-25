package models

import (
	"encoding/json"

	"github.com/baishancloud/mallard/corelib/utils"
)

// EndpointData is object to recieve transfer's endpoint data
type EndpointData struct {
	Config *EndpointConfig `json:"config"`
	Hash   string          `json:"hash"`
	Time   int64           `json:"tfr_time"`
}

// EndpointConfig is config data for one endpoint
type EndpointConfig struct {
	// Groups  []int    `json:"-"`
	// Templates  []*Template     `json:"-"`
	Plugins    []string         `json:"plgs,omitempty"`
	Strategies []*Strategy      `json:"ss,omitempty"`
	Builtin    *EndpointBuiltin `json:"bt,omitempty"`
	hashCode   string
}

// Len is data length of en endpoint,
// groups + plugins + templates + strategies
func (ec *EndpointConfig) Len() int {
	return len(ec.Strategies)
}

// Hash return config hash code
func (ec *EndpointConfig) Hash() string {
	if ec.hashCode == "" {
		b, _ := json.Marshal(ec)
		ec.hashCode = utils.MD5HashBytes(b)
	}
	return ec.hashCode
}

// IsUsingStrategy checks the strategy id is using for this endpoint
func (ec *EndpointConfig) IsUsingStrategy(id int) bool {
	for _, st := range ec.Strategies {
		if st.ID == id {
			return true
		}
	}
	return false
}

// EndpointBuiltin is config for agent builtin service
type EndpointBuiltin struct {
	Ports []int64 `json:"ports,omitempty"`
}

// EndpointHeartbeat is endpoint heartbeat item
type EndpointHeartbeat struct {
	Version       string `json:"v,omitempty"`
	PluginVersion string `json:"pv,omitempty"`
	IP            string `json:"ip,omitempty"`
	Remote        string `json:"rmt,omitempty"`
	Endpoint      string `json:"ep,omitempty"`
}

// HostService is service info
type HostService struct {
	Hostname       string `json:"hostname,omitempty" db:"hostname"`
	ServiceName    string `json:"service_name,omitempty" db:"service_name"`
	ServiceVersion string `json:"service_version,omitempty" db:"service_version"`
	ServiceBuild   string `json:"service_build,omitempty" db:"service_build"`
	CreatedAt      int64  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      int64  `json:"updated_at,omitempty" db:"updated_at"`
}

// Key returns unique key for the service
func (hs HostService) Key() string {
	return hs.Hostname + "." + hs.ServiceName
}
