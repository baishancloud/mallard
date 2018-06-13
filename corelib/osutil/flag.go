package osutil

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Flags is a simple flag interrupter to print value and load correct config file
func Flags(version string, buildTime string, cfg interface{}) {
	c := flag.Bool("c", false, "show default config")
	v := flag.Bool("v", false, "show version")
	vt := flag.Bool("vt", false, "show version and built time")
	flag.Parse()
	if *c {
		b, _ := json.MarshalIndent(cfg, "", "\t")
		fmt.Println(string(b))
		os.Exit(0)
	}
	if *v {
		fmt.Println(version)
		os.Exit(0)
	}
	if *vt {
		fmt.Println("version : " + version)
		fmt.Println("build : " + buildTime)
		fmt.Println("go : " + runtime.Version())
		fmt.Println("os : " + runtime.GOOS + "/" + runtime.GOARCH)
		os.Exit(0)
	}
}
