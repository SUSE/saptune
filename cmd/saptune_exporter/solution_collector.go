package main

import (
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const subsystem = "solution"

// SolutionCollector is the saptune solution collector
type SolutionCollector struct {
	DefaultCollector
}

// NewSolutionCollector creates a new solution saptune collector
func NewSolutionCollector() (*SolutionCollector, error) {
	c := &SolutionCollector{
		NewDefaultCollector(subsystem),
	}
	// this metric are set by  setSolutionEnabledMetric
	c.SetDescriptor("bobj", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("sap_ase", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("maxdb", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("netweaver", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("hana", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("netweaver_hana", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("s4hana_appserver", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("s4hana_dbserver", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	c.SetDescriptor("s4hana_app_db", "Status of saptune solution. 1 means the solution is enabled on node, 0 otherwise", nil)
	return c, nil
}

// Collect various metrics for saptune solution
func (c *SolutionCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting saptune solution metrics...")
	c.setSolutionEnabledMetric(ch)
}

func (c *SolutionCollector) setSolutionEnabledMetric(ch chan<- prometheus.Metric) {
	// by default all solution are disable, even if saptune error out
	solutions := make(map[string]int)
	solutions["bobj"] = 0
	solutions["sap_ase"] = 0
	solutions["maxdb"] = 0
	solutions["netweaver"] = 0
	solutions["hana"] = 0
	solutions["netweaver_hana"] = 0
	solutions["s4hana_appserver"] = 0
	solutions["s4hana_dbserver"] = 0
	solutions["s4hana_app_db"] = 0

	solutionName, err := exec.Command("saptune", "solution", "enabled").CombinedOutput()
	if err != nil {
		log.Warnf("%v - Failed to run saptune solution enabled command n %s ", err, string(solutionName))
		return
	}
	// set active solution accordingly to ouput of saptune
	switch string(solutionName) {
	case "BOBJ":
		solutions["bobj"] = 1
	case "HANA":
		solutions["hana"] = 1
	case "MAXDB":
		solutions["maxdb"] = 1
	case "NETWEAVER":
		solutions["netweaver"] = 1
	case "NETWEAVER+HANA":
		solutions["netweaver_hana"] = 1
	case "S4HANA-APP+DB":
		solutions["s4hana_app_db"] = 1
	case "S4HANA-APPSERVER":
		solutions["s4hana_appserver"] = 1
	case "S4HANA-DBSERVER":
		solutions["s4hana_dbserver"] = 1
	case "SAP-ASE":
		solutions["sap_ase"] = 1

	default:
		log.Warnf("Unrecognized saptune solution name %s", solutionName)
	}
	for solutionName, status := range solutions {
		ch <- c.MakeGaugeMetric(solutionName, float64(status))
	}

}
