package influxdb

import (
	"sort"
	"strings"
	"sync"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

type (
	// ClusterOption is cluster of groups
	ClusterOption map[string]GroupOption
	// GroupOption is group for some influxdb
	GroupOption struct {
		URLs      map[string]string `json:"urls,omitempty"`
		Db        string            `json:"db,omitempty"`
		User      string            `json:"user,omitempty"`
		Password  string            `json:"password,omitempty"`
		Blacklist []string          `json:"blacklist,omitempty"`
		WhiteList []string          `json:"white_list,omitempty"`
	}
	// Group is influxdb nodes manager
	Group struct {
		name      string
		nodesLock sync.RWMutex
		nodes     map[string]*Node
		encoder   Encoder
	}
)

var (
	groupLock   sync.RWMutex
	groups      = make(map[string]*Group)
	groupAccept = make(map[string][]string)
)

// SetCluster sets cluster
func SetCluster(cOpt ClusterOption) {
	// prepare metrics and group names
	var totalNames []string
	var metricsList []string
	for name, gOpt := range cOpt {
		totalNames = append(totalNames, name)
		metricsList = append(metricsList, gOpt.WhiteList...)
		metricsList = append(metricsList, gOpt.Blacklist...)
	}

	// prepare accept defaul data
	acceptTemp := make(map[string]map[string]bool)
	for _, metric := range metricsList {
		acceptTemp[metric] = make(map[string]bool)
		for _, name := range totalNames {
			acceptTemp[metric][name] = true
		}
	}
	acceptTemp["*"] = make(map[string]bool)
	for _, name := range totalNames {
		acceptTemp["*"][name] = true
	}
	for name, gOpt := range cOpt {
		groupLock.Lock()
		groups[name] = NewGroup(name, gOpt)
		groupLock.Unlock()
		if len(gOpt.WhiteList) > 0 {
			for metric := range acceptTemp {
				acceptTemp[metric][name] = false
			}
			acceptTemp["*"][name] = false
			for _, metric := range gOpt.WhiteList {
				acceptTemp[metric][name] = true
			}
			continue
		}
		if len(gOpt.Blacklist) > 0 {
			for _, metric := range gOpt.Blacklist {
				acceptTemp[metric][name] = false
			}
		}
	}
	acceptings := make(map[string][]string, len(acceptTemp))
	for metric, groups := range acceptTemp {
		acceptings[metric] = acceptKeys(groups)
	}

	groupLock.Lock()
	groupAccept = acceptings
	log.Info("set-options", "accept", groupAccept)
	groupLock.Unlock()
}

func acceptKeys(groups map[string]bool) []string {
	var list []string
	for name, ok := range groups {
		if ok {
			list = append(list, name)
		}
	}
	sort.Sort(sort.StringSlice(list))
	return list
}

func trimAccepting(accepting []string) []string {
	arr := make([]string, 0, len(accepting))
	for _, v := range accepting {
		if v == "-" || v == "" {
			continue
		}
		arr = append(arr, v)
	}
	return arr
}

// NewGroup creats new influxdb group
func NewGroup(name string, opt GroupOption) *Group {
	urls := make(map[string]string)
	for key, u := range opt.URLs {
		u = strings.TrimSuffix(u, "/")
		u += "/write?db=" + opt.Db + "&precision=s"
		urls[key] = u
	}
	group := &Group{
		name: name,
		encoder: Encoder{
			Db: opt.Db,
		},
	}
	nodes := make(map[string]*Node)
	for key, u := range urls {
		node := NewNode(NodeOption{
			URL:       u,
			User:      opt.User,
			Password:  opt.Password,
			Name:      name + "_" + key,
			GroupName: name,
			Timeout:   time.Second * 10,
		})
		nodes[key] = node
	}
	group.nodesLock.Lock()
	group.nodes = nodes
	group.nodesLock.Unlock()
	log.Info("init-group", "name", name)
	return group
}

const (
	// MaxPointsBytes is max length when sending points to influxdb in once request
	MaxPointsBytes = 256 * 1024
)

// Send sends points to nodes in group
func (g *Group) Send(points []*client.Point) {
	dataLen := int64(len(points))
	data, err := g.encoder.Encode(points)
	if err != nil {
		log.Warn("points-encode-error", "error", err, "points", points)
		return
	}
	if dataLen > 3 && len(data) > MaxPointsBytes {
		idx := len(points) / 2
		p1 := points[:idx]
		p2 := points[idx:]
		log.Debug("split-points", "bytes", len(data), "len", dataLen, "idx", idx)
		go g.Send(p1)
		go g.Send(p2)
		return
	}
	g.nodesLock.RLock()
	defer g.nodesLock.RUnlock()

	for _, node := range g.nodes {
		nd2 := node
		go nd2.Send(data, dataLen, true)
	}
}
