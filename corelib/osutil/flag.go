package osutil

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/baishancloud/mallard/corelib/utils"
)

// Flags is a simple flag interrupter to print value and load correct config file
func Flags(version string, buildTime string, cfg interface{}) {
	c := flag.String("c", "", "")
	dc := flag.Bool("dc", false, "show default config")
	v := flag.Bool("v", false, "show version")
	vt := flag.Bool("vt", false, "show version and built time")
	flag.Parse()
	if *c != "" {
		// ignore it
	}
	if *dc {
		b, _ := json.MarshalIndent(cfg, "", "\t")
		fmt.Println(string(b))
		os.Exit(0)
	}
	if *v {
		fmt.Println(version)
		os.Exit(0)
	}
	if *vt {
		exeFile, _ := os.Executable()
		hash, _ := utils.MD5File(exeFile)
		fmt.Println("version : " + version)
		fmt.Println("build : " + buildTime)
		fmt.Println("hash : " + hash)
		fmt.Println("go : " + runtime.Version())
		fmt.Println("os : " + runtime.GOOS + "/" + runtime.GOARCH)
		os.Exit(0)
	}
}
