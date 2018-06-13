package plugins

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/baishancloud/mallard/componentlib/agent/serverinfo"
	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/models"
	"github.com/baishancloud/mallard/corelib/zaplog"
)

var (
	pluginDir        = "plugins"
	pluginLogDir     = "var/plugins_log"
	pluginRunDirList = []string{"sys"}

	pluginsExecCount = expvar.NewDiff("plugins.exec")
	pluginsCount     = expvar.NewBase("plugins")
	pluginsFailCount = expvar.NewDiff("plugins.fail")
)

func init() {
	expvar.Register(pluginsExecCount, pluginsFailCount, pluginsCount)
}

// SetDir sets directory to run plugins
func SetDir(dir string, logDir string, runningDir []string) {
	if len(runningDir) == 0 {
		runningDir = []string{"sys"}
	}
	pluginDir = dir
	pluginLogDir = logDir
	pluginRunDirList = runningDir
	log.Info("set-dir", "dir", dir, "log_dir", logDir, "running", runningDir)
}

// Version return plugins version number
func Version() string {
	versionFile := filepath.Join(pluginDir, "sys/version")
	data, _ := ioutil.ReadFile(versionFile)
	if len(data) == 0 {
		return ""
	}
	return strings.TrimRight(string(data), "\n")
}

var (
	plugins     = make(map[string]*Plugin)
	pluginsLock sync.RWMutex

	log = zaplog.Zap("plugins")
)

// SyncScan starts scanning sync
func SyncScan(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		ScanFiles()
		<-ticker.C
	}
}

// ScanFiles reads plugins files list to memory
func ScanFiles() {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	now := time.Now().Unix()
	nowDirs := pluginRunDirList
	for _, dir := range nowDirs {
		fullDir := filepath.Join(pluginDir, dir)
		fullDir, err := getCorrectDir(fullDir, serverinfo.Hostname())
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug("stats-not-exist", "dir", dir)
				continue
			}
			log.Warn("stat-fail", "dir", dir, "error", err)
			continue
		}
		filepath.Walk(fullDir, func(fpath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !isAllowFilename(info.Name()) {
				log.Debug("ignore-file", "file", info.Name())
				return nil
			}
			plugin := plugins[fpath]
			if plugin == nil {
				logFile := filepath.Join(pluginDir, strings.Replace(filepath.ToSlash(fpath), "/", "_", -1)) + ".log"
				var err error
				plugin, err = NewPlugin(fpath, logFile, info)
				if err != nil {
					log.Warn("new-plugin-error", "file", fpath, "error", err)
					return nil
				}
				plugins[fpath] = plugin
				log.Debug("new-plugin", "file", fpath)
			}
			plugin.ReloadTime = now
			plugin.FileModTime = info.ModTime().Unix()
			return nil
		})
		log.Debug("reload-dir", "dir", fullDir)
	}
	for key, plg := range plugins {
		if now-plg.ReloadTime > 180 {
			delete(plugins, key)
			log.Info("delete-plugin", "file", key)
		}
	}
	log.Info("reload", "dir", pluginRunDirList, "plugins", len(plugins))
	pluginsCount.Set(int64(len(plugins)))
}

func getCorrectDir(dir, endpoint string) (string, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", errors.New("directory is file")
	}
	secondDir := filepath.Join(dir, endpoint)
	if stat, _ := os.Stat(secondDir); stat != nil && stat.IsDir() {
		return secondDir, nil
	}
	return dir, nil
}

func isAllowFilename(filename string) bool {
	if len(filename) == 0 {
		return false
	}
	firstByte := filename[0]
	if firstByte < 49 || firstByte > 57 {
		return false
	}
	ext := filepath.Ext(filename)
	if ext == ".py" || ext == ".sh" || ext == ".exe" {
		return true
	}
	return false
}

// Exec starts plugins execution
func Exec(mCh chan<- []*models.Metric) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		now := <-ticker.C
		go execFiles(mCh, now.Unix())
	}
}

var (
	parsedAsNew = false
)

// SetParsedAsNew sets plugins to parses new metric object
func SetParsedAsNew(flag bool) {
	parsedAsNew = flag
}

func execFiles(mCh chan<- []*models.Metric, now int64) {
	pluginsLock.RLock()
	defer pluginsLock.RUnlock()

	for _, plugin := range plugins {
		if plugin.ShouldExec(now) {
			go func(plugin *Plugin) {
				st := time.Now()
				metrics, err := plugin.Exec(parsedAsNew)
				if err != nil {
					if err == ErrPluginExecBlank {
						log.Debug("exec-0-data", "file", plugin.File)
						return
					}
					log.Warn("exec-error", "file", plugin.File, "error", err)
					pluginsFailCount.Incr(1)
					return
				}
				mLen := len(metrics)
				if mLen == 0 {
					log.Warn("exec-0", "file", plugin.File)
					return
				}
				record := make(map[string]struct{})
				for _, m := range metrics {
					record[m.Name] = struct{}{}
				}
				log.Debug("exec-plugin",
					"file", plugin.File,
					"len", mLen,
					"names", keysOfMap(record),
					"ms", time.Since(st).Nanoseconds()/1e6,
				)
				mCh <- metrics
				pluginsExecCount.Incr(int64(len(metrics)))
			}(plugin)
		}
	}
}

func keysOfMap(m map[string]struct{}) []string {
	if len(m) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// FilesHash returns md5 hash of all plugins files in manager
func FilesHash() map[string]string {
	pluginsLock.RLock()
	defer pluginsLock.RUnlock()

	result := make(map[string]string)
	for _, plugin := range plugins {
		result[plugin.File] = plugin.Hash()
	}
	return result
}
