package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"tykky"
)

func main() {
	logd := flag.String("log-dir", os.Getenv("$TMPDIR"), "Where to place logfiles")
	c := flag.String("config", "", "Configuration for installing conda")
	flag.Parse()
	tykky.LogDir = *logd
	f, err := os.Open(*c)
	if err != nil {
		panic("Failed to open config for installing conda")
	}
	byteVal, _ := ioutil.ReadAll(f)
	var config tykky.CondaInstallerConfig
	json.Unmarshal(byteVal, &config)
	config.SetupConda()
}
