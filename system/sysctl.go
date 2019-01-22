// Manipulate sysctl switches.
package system

import (
	"fmt"
	"io/ioutil"
	"log"
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

// Read a sysctl key and return the string value.
func GetSysctlString(parameter string) (string, error) {
	val, err := ioutil.ReadFile(path.Join("/proc/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		return "", fmt.Errorf("Failed to read sysctl key '%s': %v", parameter, err)
	}
	return strings.TrimSpace(string(val)), nil
}

// Read an integer sysctl key.
func GetSysctlInt(parameter string) (int, error) {
	value, err := GetSysctlString(parameter)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// Read an uint64 sysctl key.
func GetSysctlUint64(parameter string) (uint64, error) {
	value, err := GetSysctlString(parameter)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(value, 10, 64)
}

// Write a string sysctl value.
func SetSysctlString(parameter, value string) error {
	err := ioutil.WriteFile(path.Join("/proc/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 0644)
	if os.IsNotExist(err) {
		log.Printf("sysctl key '%s' is not supported by os, skipping.", parameter)
	} else if err != nil {
		return fmt.Errorf("Failed to write sysctl key '%s': %v", parameter, err)
	}
	return nil
}

// Write an integer sysctl value.
func SetSysctlInt(parameter string, value int) error {
	err := SetSysctlString(parameter, strconv.Itoa(value))
	return err
}

// Write an integer sysctl value.
func SetSysctlUint64(parameter string, value uint64) error {
	err := SetSysctlString(parameter, strconv.FormatUint(value, 10))
	return err
}

// Write an integer sysctl value into the specified field pf the key.
func SetSysctlUint64Field(param string, field int, value uint64) error {
	fields, err := GetSysctlString(param)
	if err != nil {
		return err
	}
	allFields := consecutiveSpaces.Split(fields, -1)
	if field < len(allFields) {
		allFields[field] = strconv.FormatUint(value, 10)
		err = SetSysctlString(param, strings.Join(allFields, " "))
	} else {
		err = fmt.Errorf("Failed to write sysctl key field '%s' %d: %v", param, field, err)
	}
	return err
}

func IsPagecacheAvailable() bool {
        _, err := ioutil.ReadFile(path.Join("/proc/sys", strings.Replace(SysctlPagecacheLimitMB, ".", "/", -1)))
        if err == nil {
                return true
        }
        return false
}
