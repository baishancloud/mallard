package serverinfo

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/baishancloud/mallard/corelib/utils"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

type (
	// Data is info list from allconf
	Data struct {
		Sertypes        string `json:"sertypes,omitempty"`
		Cachegroup      string `json:"cachegroup,omitempty"`
		StorageGroup    string `json:"storage_group,omitempty"`
		IP              string `json:"ip,omitempty"`
		HostnameOS      string `json:"hostname_os,omitempty"`
		HostnameFirst   string `json:"hostname_first,omitempty"`
		HostnameAllConf string `json:"hostname_allconf,omitempty"`
		HostIDAllConf   string `json:"hostid_allconf,omitempty"`
	}
)

var (
	svrData *Data
	log     = zaplog.Zap("serverinfo")
)

var (
	cachegroupFile   = "/allconf/cachegroup.conf"
	storageGroupFile = "/allconf/storagegroup.conf"
	sertypesFile     = "/allconf/sertypes.conf"
	hostnameFile     = "/allconf/hostname.conf"
	hostIDFile       = "/allconf/hostid.conf"
)

// Scan starts scanning serverinfo
func Scan(defaultEp string) {
	Read(defaultEp)
	log.Info("read", "info", svrData, "use_allconf", UseAllConf)
	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()
		for {
			<-ticker.C
			Read(defaultEp)
		}
	}()
}

// Read reads current serverinfo
func Read(defaultEp string) (*Data, error) {
	newData := &Data{
		IP:            strings.Join(utils.LocalIPs(), ","),
		HostnameOS:    utils.HostName(),
		HostnameFirst: defaultEp,
	}
	dat, err := ioutil.ReadFile(cachegroupFile)
	if err == nil {
		cg := string(dat)
		cg = strings.Split(cg, "\n")[0]
		cqi := strings.Index(cg, "=")
		if cqi > 9 && len(cg) > cqi {
			newData.Cachegroup = cg[cqi+1:]
		}
	}
	dat, err = ioutil.ReadFile(storageGroupFile)
	if err == nil {
		cg := string(dat)
		cg = strings.Split(cg, "\n")[0]
		cqi := strings.Index(cg, "=")
		if cqi > 9 && len(cg) > cqi {
			newData.StorageGroup = cg[cqi+1:]
		}
	}
	dat, err = ioutil.ReadFile(sertypesFile)
	if err == nil {
		st := string(dat)
		st = strings.Split(st, "\n")[0]
		sti := strings.Index(st, "=")
		if sti > 7 && len(st) > 9 {
			st = st[sti+1:]
			sts := strings.Replace(st, ",", "|", -1)
			newData.Sertypes = sts
		}
	}
	dat, err = ioutil.ReadFile(hostnameFile)
	if err == nil {
		str := strings.TrimSuffix(string(dat), "\n")
		sts := strings.Split(str, "=")
		if len(sts) == 2 {
			newData.HostnameAllConf = sts[1]
		}
	}
	dat, err = ioutil.ReadFile(hostIDFile)
	if err == nil {
		str := strings.TrimSuffix(string(dat), "\n")
		sts := strings.Split(str, "=")
		if len(sts) == 2 {
			newData.HostIDAllConf = sts[1]
		}
	}
	svrData = newData
	return svrData, nil
}

var (
	// UseAllConf sets to use hostname from allconf
	UseAllConf = true
)

// Hostname returns server hostname
func Hostname() string {
	if svrData == nil {
		return ""
	}
	if svrData.HostnameFirst != "" {
		return svrData.HostnameFirst
	}
	if UseAllConf && svrData.HostnameAllConf != "" {
		return svrData.HostnameAllConf
	}
	return svrData.HostnameOS
}

// Sertypes returns server type string
func Sertypes() string {
	if svrData == nil {
		return ""
	}
	return svrData.Sertypes
}

// Cachegroup returns cache group name
func Cachegroup() string {
	if svrData == nil {
		return ""
	}
	return svrData.Cachegroup
}

// StorageGroup returns storage group name
func StorageGroup() string {
	if svrData == nil {
		return ""
	}
	return svrData.StorageGroup
}

// IP returns ip
func IP() string {
	if svrData == nil {
		return ""
	}
	return svrData.IP
}

// Cached returns cached serverinfo data
func Cached() *Data {
	return svrData
}
