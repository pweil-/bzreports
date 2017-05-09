package bzreports

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kolo/xmlrpc"
)

func validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.User == "" || config.Password == "" {
		return fmt.Errorf("invalid username/password")
	}
	return nil
}

func (s *Server) RunReports() error {
	if err := validate(&s.Config); err != nil {
		return err
	}

	ret := struct {
		Bugs []Bug `xmlrpc:"bugs"`
	}{}

	attrs := map[string]interface{}{
		"Bugzilla_login":    s.Config.User,
		"Bugzilla_password": s.Config.Password,

		"product":   products,
		"component": s.Config.Components,
		"severity":  severity,
		"status":    status,
	}

	glog.Infof("creating client...")
	client, err := xmlrpc.NewClient("https://bugzilla.redhat.com/xmlrpc.cgi", nil)
	if err != nil {
		return fmt.Errorf("error creating xmlrpc client: %v", err)

	}
	glog.Infof("client created...")

	glog.Infof("running query...")
	if err := client.Call("Bug.search", attrs, &ret); err != nil {
		return fmt.Errorf("error calling Bug.search: %#v", err)
	}
	glog.Infof("query run...")

	glog.Infof("filtering...")
	filteredBugs := []Bug{}
	for i, bug := range ret.Bugs {
		if !hasUpcomingRelease(bug.Keywords) && hasVersion3(bug.Version) {
			filteredBugs = append(filteredBugs, ret.Bugs[i])
		}
	}
	glog.Infof("bugs filtered...")

	glog.Infof("formatting data...")
	componentCounts := s.makeComponentCountMap()
	for _, bug := range filteredBugs {
		for _, component := range bug.Component {
			componentCounts[component] = componentCounts[component] + 1
		}
	}
	teamCounts := s.makeTeamCountMap()
	for component, count := range componentCounts {
		team := s.getTeamForComponent(component)
		teamCounts[team] = teamCounts[team] + count
	}

	sortedTeams := sortMapKeys(teamCounts)
	sortedComponents := sortMapKeys(componentCounts)

	headers := []string{"date", "total"}
	headers = append(headers, sortedTeams...)
	headers = append(headers, sortedComponents...)

	// setup the data
	data := []string{time.Now().Format(time.RFC3339), fmt.Sprintf("%d", len(filteredBugs))}

	for _, team := range sortedTeams {
		data = append(data, fmt.Sprintf("%d", teamCounts[team]))
	}
	for _, c := range sortedComponents {
		data = append(data, fmt.Sprintf("%d", componentCounts[c]))
	}
	glog.Infof("data formatted...")

	fileName := filepath.Join(s.Config.DataDir, "data.txt")

	glog.Infof("writing to %s...", fileName)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0750)
	newFile := false
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(fileName)
			if err != nil {
				return fmt.Errorf("error creating file: %v", err)
			}
			newFile = true
		} else {
			return fmt.Errorf("error opening file: %v", err)
		}
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	if newFile {
		err = csvWriter.Write(headers)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}
	err = csvWriter.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	csvWriter.Flush()
	glog.Infof("done!")
	return nil
}

func (s *Server) makeTeamCountMap() map[string]int {
	teamCounts := map[string]int{}
	for team, _ := range s.Config.ComponentOwners {
		teamCounts[team] = 0
	}
	return teamCounts
}

func (s *Server) makeComponentCountMap() map[string]int {
	componentCounts := map[string]int{}
	for _, component := range s.Config.Components {
		componentCounts[component] = 0
	}
	return componentCounts
}

func (s *Server) createHeaders() []string {
	headers := []string{"total"}

	for team, _ := range s.Config.ComponentOwners {
		headers = append(headers, team)
	}

	for _, c := range s.Config.Components {
		headers = append(headers, c)
	}
	return headers
}

func (s *Server) getTeamForComponent(component string) string {
	for team, components := range s.Config.ComponentOwners {
		for _, c := range components {
			if c == component {
				return team
			}
		}
	}
	glog.Errorf("unknown component for team: %s", component)
	return "unknown"
}

func hasUpcomingRelease(keywords []string) bool {
	for _, kw := range keywords {
		if kw == "UpcomingRelease" {
			return true
		}
	}
	return false
}
func hasVersion3(versions []string) bool {
	for _, v := range versions {
		if strings.HasPrefix(v, "3.") {
			return true
		}
	}
	return false
}

func sortMapKeys(m map[string]int) []string {
	keys := []string{}
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}