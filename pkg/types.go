package bzreports

import "time"

type Config struct {
	ComponentOwners map[string][]string
	Components      []string
	DataDir         string
	User            string
	Password        string
}

type Server struct {
	Config Config
}

// initial struct borrowed from github.com/dmacvicar/gorgojo
type Bug struct {
	Id             int       `xmlrpc:"id"`
	Summary        string    `xmlrpc:"summary"`
	CreationTime   time.Time `xmlrpc:"creation_time"`
	AssignedTo     string    `xmlrpc:"assigned_to"`
	Component      []string  `xmlrpc:"component"`
	LastChangeTime time.Time `xmlrpc:"last_change_time"`
	Severity       string    `xmlrpc:"severity"`
	Status         string    `xmlrpc:"status"`
	Keywords       []string  `xmlrpc:"keywords"`
	Version        []string  `xmlrpc:"version"`
}

var (
	products = []string{"OpenShift Container Platform", "OpenShift Online", "OpenShift Origin"}
	severity = []string{"unspecified", "urgent", "high", "medium"}
	status   = []string{"NEW", "ASSIGNED", "POST", "MODIFIED", "ON_DEV"}
)
