package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const subsystem = "solution"

type solutionCollector struct {
	DefaultCollector
}

// NewCollector creates a new solution saptune collector
func NewSolutionCollector() (*solutionCollector, error) {
	c := &solutionCollector{
		NewDefaultCollector(subsystem),
	}
	c.SetDescriptor("hana_enabled", "Status of hanadb solution. 1 means the solution is enabled on node, 0 otherwise", nil)

	return c, nil
}

func (c *solutionCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting saptune solution metrics...")

	ch <- c.MakeGaugeMetric("hana_enabled", 1)
}
