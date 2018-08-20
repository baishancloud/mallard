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
		SertypesConf    string `json:"sertypes_conf,omitempty"`
		Cachegroup      string `json:"cachegroup,omitempty"`
		StorageGroup    string `json:"storage_group,omitempty"`
		IP              string `json:"ip,omitempty"`
		HostnameOS      string `json:"hostname_os,omitempty"`
		HostnameFirst   string `json:"hostname_first,omitempty"`
		HostnameAllConf string `json:"hostname_allconf,omitempty"`
		HostIDAllConf   string `json:"hostid_allconf,omitempty"`
		useAllConf      bool
	}
)

var (
	svrData = &Data{}
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
func Scan(defaultEp string, useAllConf bool) {
	Read(defaultEp, useAllConf)
	log.Info("read", "info", svrData, "use_allconf", useAllConf)

	go utils.TickerThen(time.Minute, func() {
		Read(defaultEp, useAllConf)
		log.Info("read", "info", svrData, "use_allconf", useAllConf)
	})
}

// Read reads current serverinfo
func Read(defaultEp string, useAllConf bool) (*Data, error) {
	newData := &Data{
		IP:            strings.Join(utils.LocalIPs(), ","),
		HostnameOS:    utils.HostName(),
		HostnameFirst: defaultEp,
		useAllConf:    useAllConf,
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

// Hostname returns server hostname
func Hostname() string {
	if svrData.HostnameFirst != "" {
		return svrData.HostnameFirst
	}
	if svrData.useAllConf && svrData.HostnameAllConf != "" {
		return svrData.HostnameAllConf
	}
	return svrData.HostnameOS
}

// Sertypes returns server type string
func Sertypes() string {
	if svrData.Sertypes == "" {
		return svrData.SertypesConf
	}
	return svrData.Sertypes
}

// SetSertypes sets sertypes
func SetSertypes(sertypes string) {
	if svrData != nil {
		svrData.SertypesConf = sertypes
	}
}

// Cachegroup returns cache group name
func Cachegroup() string {
	return svrData.Cachegroup
}

// StorageGroup returns storage group name
func StorageGroup() string {
	return svrData.StorageGroup
}

// IP returns ip
func IP() string {
	return svrData.IP
}

// Cached returns cached serverinfo data
func Cached() *Data {
	return svrData
}
