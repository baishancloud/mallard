package sqldata

import (
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/models"
)

var (
	selectGroupPluginsSQL = "SELECT group_id, dir FROM plugin_dir ORDER BY group_id ASC"
)

// ReadGroupPlugins reads plugins dirs for each group
// return as map[groupID][dir1,dir2,dir3]
func ReadGroupPlugins() (map[int][]string, error) {
	rows, err := portalDB.Query(selectGroupPluginsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		count   int
		plugins = make(map[int][]string)
	)
	for rows.Next() {
		var groupID int
		var pluginDir string
		if err = rows.Scan(&groupID, &pluginDir); err != nil {
			continue
		}
		plugins[groupID] = append(plugins[groupID], pluginDir)
		count++
	}
	return plugins, nil
}

var (
	selectGroupHostsSQL = "SELECT group_id, host_id FROM group_host ORDER BY group_id ASC, host_id ASC"
)

// ReadGroupHosts query hosts in host-group
func ReadGroupHosts() (map[int][]int, error) {
	rows, err := portalDB.Query(selectGroupHostsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var (
		hosts = make(map[int][]int, 1e3)
		count int
	)
	for rows.Next() {
		var groupID int
		var hostID int
		if err = rows.Scan(&groupID, &hostID); err != nil {
			continue
		}
		hosts[hostID] = append(hosts[hostID], groupID)
		count++
	}
	return hosts, err
}

var (
	selectHostsNamesSQL = "SELECT id,hostname,ip,agent_version,plugin_version,remote_ip,maintain_begin,maintain_end,live_at,sertypes FROM host ORDER BY id ASC"
)

// ReadHosts query host's id-name pairs
func ReadHosts() (map[int]string, map[string]string, map[string]int64, map[string][]interface{}, error) {
	rows, err := portalDB.Query(selectHostsNamesSQL)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer rows.Close()
	var (
		hostName     = make(map[int]string)
		hostInfo     = make(map[string]string)
		hostMaintain = make(map[string]int64)
		hostLiveInfo = make(map[string][]interface{})
		count        int
		now          = time.Now().Unix()
	)
	for rows.Next() {
		var (
			hostID             int
			name               string
			agentVersion       string
			agentIP            string
			agentPluginVersion string
			remoteIP           string
			maintainBegin      int64
			maintainEnd        int64
			liveAt             int64
			sertypes           string
		)
		if err = rows.Scan(&hostID, &name, &agentIP, &agentVersion, &agentPluginVersion, &remoteIP, &maintainBegin, &maintainEnd, &liveAt, &sertypes); err != nil {
			continue
		}
		if name != "" {
			hostInfo[name] = strings.TrimSpace(agentIP + agentVersion + agentPluginVersion + remoteIP)
			hostLiveInfo[name] = []interface{}{
				agentIP, agentVersion, agentPluginVersion, remoteIP, liveAt, sertypes,
			}
		}
		hostName[hostID] = name
		if maintainBegin > 0 && maintainEnd > 0 && now > maintainBegin && now < maintainEnd {
			hostMaintain[name] = maintainEnd
			hostLiveInfo[name] = append(hostLiveInfo[name], maintainEnd)
		}
		count++
	}
	return hostName, hostInfo, hostMaintain, hostLiveInfo, nil
}

var (
	selectHostServiceSQL = "SELECT hostname,service_name,service_version,service_build,created_at,updated_at FROM host_service ORDER BY service_name ASC"
)

// ReadHostServices query hosts services
func ReadHostServices() (map[string]*models.HostService, error) {
	rows, err := portalDB.Queryx(selectHostServiceSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	svcList := make(map[string]*models.HostService)
	for rows.Next() {
		svc := new(models.HostService)
		if err = rows.StructScan(svc); err != nil {
			continue
		}
		key := svc.Key()
		if key != "" {
			svcList[key] = svc
		}
	}
	return svcList, nil
}
