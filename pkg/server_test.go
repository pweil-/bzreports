package bzreports

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestQueries(t *testing.T) {
	server := Server{}

	config := Config{}
	file, err := ioutil.ReadFile("/home/pweil/codebase/bzreports/src/github.com/pweil-/bzreports/assets/config.json")
	if err != nil {
		t.Fatalf("error reading config: %v", err)
	}

	json.Unmarshal(file, &config)

	server.Config = config
	server.RunReports()
	t.Errorf("foo")
}





