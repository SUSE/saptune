# Saptune exporter

This is a bespoke Prometheus exporter used to enable the monitoring of Pacemaker based HA clusters.  
`
The exporter run on port `9758` and it is **officially** registered upstream at Prometheus doc: https://github.com/prometheus/prometheus/wiki/Default-port-allocations


## Table of Contents

1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
   1. [Metrics](doc/metrics.md)
4. [Design](#design)

# Features:

* monitor and export saptune solution metric

# Installation:

At this time

# Usage:

You can run the exporter in any of the nodes you have saptune installed.

$ ./saptune_exporter  
INFO[0000] Serving metrics on 0.0.0.0:9664

# Design:

The following project follows convention used by implementation of other exporters like https://github.com/ClusterLabs/ha_cluster_exporter and prometheus upstream conventions.
