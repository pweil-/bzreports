package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"flag"

	"github.com/golang/glog"

	bzreports "github.com/pweil-/bzreports/pkg"
)

func main() {
	flag.Parse()
	server := bzreports.Server{}

	config := bzreports.Config{}
	file, err := ioutil.ReadFile("/home/pweil/codebase/bzreports/src/github.com/pweil-/bzreports/assets/config.json")
	if err != nil {
		glog.Fatalf("error reading config: %v", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		glog.Errorf("error unmarshalling config: %v", err)
		os.Exit(1)
	}

	server.Config = config
	err = server.RunReports()
	if err != nil {
		glog.Errorf("error running RunQueries: %v", err)
		os.Exit(1)
	}
}
