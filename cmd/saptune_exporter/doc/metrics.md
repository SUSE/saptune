# Metrics specification

This document describes the metrics exposed by `saptune_exporter`.

General notes:
- All the metrics are _namespaced_ with the prefix `saptune`, which is followed by a _subsystem_, and both are in turn composed into a _Fully Qualified Name_ (FQN) of each metrics.
- All the metrics and labels _names_ are in snake_case, as conventional with Prometheus. That said, as much as we'll try to keep this consistent throughout the project, the label _values_ may not actually follow this convention, though (e.g. value is a hostname).


These are the currently implemented subsystems.

1. [Solution](#solution)
2. [Corosync](#corosync)
3. [SBD](#sbd)
4. [DRBD](#drbd)
5. [Scrape](#scrape)


# Solution:
This metrics are related to the saptune solution collector.

- `saptune_solution_hana_enabled` 

 Status of hanadb solution. 1 means the solution is enabled on node, 0 is disabled.

- `saptune_solution_bobj_enabled`

 Status of bobj solution. 1 means the solution is enabled on node, 0 is disabled.

- `saptune_solution_maxdb_enabled`


