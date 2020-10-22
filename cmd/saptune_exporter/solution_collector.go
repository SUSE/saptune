package main

import (
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
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
	c.SetDescriptor("hana_enabled", "Status of hanadb solution. 1 means the solution is enabled on node, 0 otherwise", nil)

	return c, nil
}

// Collect various metrics for saptune solution
func (c *SolutionCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting saptune solution metrics...")

	solutionSelector := system.GetSolutionSelector()
	archSolutions, exist := solution.AllSolutions[solutionSelector]
	if !exist {
		log.Warnf("The system architecture (%s) is not supported.", solutionSelector)
		return
	}
	log.Infoln(archSolutions)
	ch <- c.MakeGaugeMetric("hana_enabled", 1)
}
