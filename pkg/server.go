package main

import (
	"fmt"
	"strings"
	"sort"
	"time"

	"github.com/golang/glog"
	"github.com/kolo/xmlrpc"
)

/*
New bugs today
How many bugs were resolved from the query (ie. closed, verified, or moved off the team)
 */


type Config struct {
	ComponentOwners   map[string][]string
	Components []string
	DataDir string
	User string
	Password string
}

type Server struct {
	Config Config
}

// initial struct borrowed from github.com/dmacvicar/gorgojo
type Bug struct {
	Id int `xmlrpc:"id"`
	Summary string `xmlrpc:"summary"`
	CreationTime time.Time `xmlrpc:"creation_time"`
	AssignedTo string `xmlrpc:"assigned_to"`
	Component []string `xmlrpc:"component"`
	LastChangeTime time.Time `xmlrpc:"last_change_time"`
	Severity string `xmlrpc:"severity"`
	Status string `xmlrpc:"status"`
	Keywords []string `xmlrpc:"keywords"`
	Version []string `xmlrpc:"version"`
}

var (
	products = []string{"OpenShift Container Platform", "OpenShift Online", "OpenShift Origin"}
	severity = []string{"unspecified", "urgent", "high", "medium"}
	status = []string{"NEW", "ASSIGNED", "POST", "MODIFIED", "ON_DEV"}
)

func (s *Server) RunQueries() {
	if s.Config.User == "" || s.Config.Password == "" {
		glog.Errorf("invalid username/password")
		return
	}

	ret := struct {
		Bugs []Bug `xmlrpc:"bugs"`
	}{}

	attrs := map[string]interface{}{
		"Bugzilla_login": s.Config.User,
		"Bugzilla_password": s.Config.Password,

		"product": products,
		"component": s.Config.Components,
		"severity": severity,
		"status": status,

	}

	client, err := xmlrpc.NewClient("https://bugzilla.redhat.com/xmlrpc.cgi", nil)
	if err != nil {
		glog.Errorf("error creating xmlrpc client: %v", err)
		return
	}

	if err := client.Call("Bug.search", attrs, &ret); err != nil {
		glog.Errorf("%#v", err)
	}

	filteredBugs := []Bug{}

	for i, bug := range ret.Bugs {
		if !hasUpcomingRelease(bug.Keywords) && hasVersion3(bug.Version) {
			filteredBugs = append(filteredBugs, ret.Bugs[i])
		}
	}

	componentCounts := map[string]int{}
	for _, bug := range filteredBugs {
		for _, component := range bug.Component {
			componentCounts[component] = componentCounts[component] + 1
		}
	}
	teamCounts := map[string]int{}
	for component, count := range componentCounts {
		team := s.getTeamForComponent(component)
		teamCounts[team] = teamCounts[team] + count
	}

	sortedTeams := sortMapKeys(teamCounts)
	sortedComponents := sortMapKeys(componentCounts)



	headers := []string{"total"}
	headers = append(headers, sortedTeams...)
	headers = append(headers, sortedComponents...)

	// setup the data
	data := []string{fmt.Sprintf("%d", len(filteredBugs))}

	for _, team := range sortedTeams {
		data = append(data, fmt.Sprintf("%d", teamCounts[team]))
	}
	for _, c := range sortedComponents {
		data = append(data, fmt.Sprintf("%d", componentCounts[c]))
	}



	fmt.Printf("%v\n", strings.Join(headers, "\t"))
	fmt.Printf("%v\n", strings.Join(data, "\t"))

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
		if strings.HasPrefix(v, "3."){
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

// read json
// run queries on a schedule and save data
// format data for charting and save data
// run web server
// display pages
