// Manipulate sysctl switches.
package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	SysctlPagecacheLimitMB          = "vm.pagecache_limit_mb"
	SysctlPagecacheLimitIgnoreDirty = "vm.pagecache_limit_ignore_dirty"
	SysctlNumaBalancing             = "kernel.numa_balancing"
	SysctlShmall                    = "kernel.shmall"
	SysctlShmax                     = "kernel.shmmax"
	SysctlShmni                     = "kernel.shmmni"
	SysctlMaxMapCount               = "vm.max_map_count"
	SysctlSem                       = "kernel.sem"
	SysctlNumberHugepages           = "vm.nr_hugepages"
	SysctlSwappines                 = "vm.swappiness"
	SysctlVFSCachePressure          = "vm.vfs_cache_pressure"
	SysctlOvercommitMemory          = "vm.overcommit_memory"
	SysctlOvercommitRatio           = "vm.overcommit_ratio"
	SysctlDirtyRatio                = "vm.dirty_ratio"
	SysctlDirtyBackgroundRatio      = "vm.dirty_background_ratio"
	SysctlNetReadMemMax             = "net.core.rmem_max"
	SysctlNetWriteMemMax            = "net.core.wmem_max"
	SysctlNetMaxBacklog             = "net.core.netdev_max_backlog"
	SysctlNetMaxconn                = "net.core.somaxconn"
	SysctlTCPReadMem                = "net.ipv4.tcp_rmem"
	SysctlTCPWriteMem               = "net.ipv4.tcp_wmem"
	SysctlTCPTimestamps             = "net.ipv4.tcp_timestamps"
	SysctlTCPSack                   = "net.ipv4.tcp_sack"
	SysctlTCPDsack                  = "net.ipv4.tcp_dsack"
	SysctlTCPFack                   = "net.ipv4.tcp_fack"
	SysctlTCPFragLowThreshold       = "net.ipv4.ipfrag_low_thresh"
	SysctlTCPFragHighThreshold      = "net.ipv4.ipfrag_high_thresh"
	SysctlTCPMaxSynBacklog          = "net.ipv4.tcp_max_syn_backlog"
	SysctlTCPSynackRetries          = "net.ipv4.tcp_synack_retries"
	SysctpTCPRetries2               = "net.ipv4.tcp_retries2"
	SysctlTCPKeepaliveTime          = "net.ipv4.tcp_keepalive_time"
	SysctlTCPKeepaliveProbes        = "net.ipv4.tcp_keepalive_probes"
	SysctlTCPKeepaliveInterval      = "net.ipv4.tcp_keepalive_intvl"
	SysctlTCPTWRecycle              = "net.ipv4.tcp_tw_recycle"
	SysctlTCPTWReuse                = "net.ipv4.tcp_tw_reuse"
	SysctlTCPFinTimeout             = "net.ipv4.tcp_fin_timeout"
	SysctlTCPMTUProbing             = "net.ipv4.tcp_mtu_probing"
	SysctlTCPSynCookies             = "net.ipv4.tcp_syncookies"
	SysctlIPAcceptSourceRoute       = "net.ipv4.conf.all.accept_source_route"
	SysctlIPAcceptRedirects         = "net.ipv4.conf.all.accept_redirects"
	SysctlIPRPFilter                = "net.ipv4.conf.all.rp_filter"
	SysctlIPIgnoreICMPBroadcasts    = "net.ipv4.icmp_echo_ignore_broadcasts"
	SysctlIPIgnoreICMPBogusError    = "net.ipv4.icmp_ignore_bogus_error_responses"
	SysctlIPLogMartians             = "net.ipv4.conf.all.log_martians"
	SysctlRandomizeVASpace          = "kernel.randomize_va_space"
	SysctlKptrRestrict              = "kernel.kptr_restrict"
	SysctlProtectHardlinks          = "fs.protected_hardlinks"
	SysctlProtectSymlinks           = "fs.protected_symlinks"
	SysctlRunChildFirst             = "kernel.sched_child_runs_first"
)

// Read a sysctl key and return the string value. Panic on error.
func GetSysctlString(parameter string, valIfNotFound string) string {
	val, err := ioutil.ReadFile(path.Join("/proc/sys", strings.Replace(parameter, ".", "/", -1)))
	if os.IsNotExist(err) {
		return valIfNotFound
	}
	if err != nil {
		panic(fmt.Errorf("failed to read sysctl string key '%s': %v", parameter, err))
	}
	return strings.TrimSpace(string(val))
}

// Read an integer sysctl key. Panic on error.
func GetSysctlInt(parameter string, valIfNotFound int) int {
	value, err := strconv.Atoi(GetSysctlString(parameter, strconv.Itoa(valIfNotFound)))
	if err != nil {
		panic(fmt.Errorf("failed to read sysctl int key '%s': %v", parameter, err))
	}
	return value
}

// Read an uint64 sysctl key. Panic on error.
func GetSysctlUint64(parameter string, valIfNotFound uint64) uint64 {
	value, err := strconv.ParseUint(GetSysctlString(parameter, strconv.FormatUint(valIfNotFound, 10)), 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to read sysctl uint64 key '%s': %v", parameter, err))
	}
	return value
}

// Extract a uint64 value from a sysctl key of many fields. Panic on error
func GetSysctlUint64Field(param string, field int, valIfNotFound uint64) uint64 {
	allFields := consecutiveSpaces.Split(GetSysctlString(param, ""), -1)
	if field < len(allFields) {
		value, err := strconv.ParseUint(allFields[field], 10, 64)
		if err != nil {
			panic(fmt.Errorf("failed to read sysctl uint64 key field '%s' %d: %v", param, field, err))
		}
		return value
	}
	return valIfNotFound
}

// Write a string sysctl value. Panic on error.
func SetSysctlString(parameter, value string) {
	if err := ioutil.WriteFile(path.Join("/proc/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 644); err != nil {
		panic(fmt.Errorf("failed to write sysctl string key '%s': %v", parameter, err))
	}
}

// Write an integer sysctl value. Panic on error.
func SetSysctlInt(parameter string, value int) {
	SetSysctlString(parameter, strconv.Itoa(value))
}

// Write an integer sysctl value. Panic on error.
func SetSysctlUint64(parameter string, value uint64) {
	SetSysctlString(parameter, strconv.FormatUint(value, 10))
}

// Write an integer sysctl value into the specified field pf the key. Panic on error
func SetSysctlUint64Field(param string, field int, value uint64) {
	allFields := consecutiveSpaces.Split(GetSysctlString(param, ""), -1)
	if field < len(allFields) {
		allFields[field] = strconv.FormatUint(value, 10)
		SetSysctlString(param, strings.Join(allFields, " "))
	} else {
		panic(fmt.Errorf("failed to write sysctl uint64 field '%s' %d, format error.", param, field))
	}
}
