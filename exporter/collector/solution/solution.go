package solution

import (
	"github.com/SUSE/saptune/exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const subsystem = "solution"

type solutionCollector struct {
	collector.DefaultCollector
}

// NewCollector creates a new solution saptune collector
func NewCollector() (*solutionCollector, error) {
	c := &solutionCollector{
		collector.NewDefaultCollector(subsystem),
	}
	c.SetDescriptor("hanadb", "Status of hanadb solution. 1 means the solution is enabled on node, 0 otherwise", []string{})

	return c, nil
}

func (c *solutionCollector) CollectWithError(ch chan<- prometheus.Metric) error {
	log.Debugln("Collecting saptune solution metrics...")

	ch <- c.MakeGaugeMetric("hanadb_enabled", 1)

	return nil
}
