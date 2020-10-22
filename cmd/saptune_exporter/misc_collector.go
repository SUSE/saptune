package main

import (
	"fmt"
	"os"

	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const subsystem_misc = "misc"

// MiscCollector is the saptune collector for general infos
type MiscCollector struct {
	DefaultCollector
}

// NewMiscCollector creates a new solution saptune collector
func NewMiscCollector() (*MiscCollector, error) {
	c := &MiscCollector{
		NewDefaultCollector(subsystem_misc),
	}
	// this metric are set by  setSolutionEnabledMetric
	c.SetDescriptor("version", "Show version of saptune", nil)

	return c, nil
}

// Collect various metrics for saptune solution
func (c *MiscCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting saptune solution metrics...")
	c.setSaptuneVersionMetric(ch)
}

func (c *MiscCollector) setSaptuneVersionMetric(ch chan<- prometheus.Metric) {
	// get saptune version
	sconf, err := txtparser.ParseSysconfigFile("/etc/sysconfig/saptune", true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to read file '/etc/sysconfig/saptune': %v\n", err)
		system.ErrorExit("", 1)
	}
	SaptuneVersion := sconf.GetString("SAPTUNE_VERSION", "")
	// check if this is always major only and no dot
	ch <- c.MakeGaugeMetric("version", float64(string(SaptuneVersion)[1]))

}
