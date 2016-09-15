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
	SYSCTL_PAGECACHE_LIMIT               = "vm.pagecache_limit_mb"
	SYSCTL_PAGECACHE_IGNORE_DIRTY        = "vm.pagecache_limit_ignore_dirty"
	SYSCTL_NUMA_BALANCING                = "kernel.numa_balancing"
	SYSCTL_SHMALL                        = "kernel.shmall"
	SYSCTL_SHMMAX                        = "kernel.shmmax"
	SYSCTL_SHMMNI                        = "kernel.shmmni"
	SYSCTL_MAX_MAP_COUNT                 = "vm.max_map_count"
	SYSCTL_SEM                           = "kernel.sem"
	SYSCTL_NR_HUGEPAGES                  = "vm.nr_hugepages"
	SYSCTL_SWAPPINESS                    = "vm.swappiness"
	SYSCTL_VFS_CACHE_PRESSURE            = "vm.vfs_cache_pressure"
	SYSCTL_OVERCOMMIT_MEMORY             = "vm.overcommit_memory"
	SYSCTL_OVERCOMMIT_RATIO              = "vm.overcommit_ratio"
	SYSCTL_DIRTY_RATIO                   = "vm.dirty_ratio"
	SYSCTL_DIRTY_BG_RATIO                = "vm.dirty_background_ratio"
	SYSCTL_NET_CORE_READ_MEM_MAX         = "net.core.rmem_max"
	SYSCTL_NET_CORE_WRITE_MEM_MAX        = "net.core.wmem_max"
	SYSCTL_NET_CORE_MAX_BACKLOG          = "net.core.netdev_max_backlog"
	SYSCTL_NET_CORE_SOMAXCONN            = "net.core.somaxconn"
	SYSCTL_NET_IPV4_TCP_READ_MEM         = "net.ipv4.tcp_rmem"
	SYSCTL_NET_IPV4_TCP_WRITE_MEM        = "net.ipv4.tcp_wmem"
	SYSCTL_NET_IPV4_TCP_TIMESTAMPS       = "net.ipv4.tcp_timestamps"
	SYSCTL_IPV4_TCP_SACK                 = "net.ipv4.tcp_sack"
	SYSCTL_IPV4_TCP_DSACK                = "net.ipv4.tcp_dsack"
	SYSCTL_IPV4_TCP_FACK                 = "net.ipv4.tcp_fack"
	SYSCTL_IPFRAG_LOW_THRESH             = "net.ipv4.ipfrag_low_thresh"
	SYSCTL_IPFRAG_HIGH_THRESH            = "net.ipv4.ipfrag_high_thresh"
	SYSCTL_IPV4_TCP_MAX_SYN_BACKLOG      = "net.ipv4.tcp_max_syn_backlog"
	SYSCTL_NET_IPV4_TCP_SYNACK_RETRIES   = "net.ipv4.tcp_synack_retries"
	SYSCTL_NET_IPV4_TCP_RETRIES2         = "net.ipv4.tcp_retries2"
	SYSCTL_NET_IPV4_TCP_KEEPALIVE_TIME   = "net.ipv4.tcp_keepalive_time"
	SYSCTL_NET_IPV4_TCP_KEEPALIVE_PROBES = "net.ipv4.tcp_keepalive_probes"
	SYSCTL_NET_IPV4_TCP_KEEPALIVE_INTVL  = "net.ipv4.tcp_keepalive_intvl"
	SYSCTL_NET_IPV4_TCP_TW_RECYCLE       = "net.ipv4.tcp_tw_recycle"
	SYSCTL_NET_IPV4_TCP_TW_REUSE         = "net.ipv4.tcp_tw_reuse"
	SYSCTL_NET_IPV4_TCP_FIN_TIMEOUT      = "net.ipv4.tcp_fin_timeout"
	SYSCTL_NET_IPV4_TCP_MTU_PROBING      = "net.ipv4.tcp_mtu_probing"
	SYSCTL_IPV4_TCP_SYNCOOKIES           = "net.ipv4.tcp_syncookies"
	SYSCTL_IPV4_ACCEPT_SOURCE_ROUTE      = "net.ipv4.conf.all.accept_source_route"
	SYSCTL_IPV4_ACCEPT_REDIRECTS         = "net.ipv4.conf.all.accept_redirects"
	SYSCTL_IPV4_RP_FILTER                = "net.ipv4.conf.all.rp_filter"
	SYSCTL_IPV4_ICMP_IGNORE_BROADCASTS   = "net.ipv4.icmp_echo_ignore_broadcasts"
	SYSCTL_IPV4_ICMP_IGNORE_BOGUS        = "net.ipv4.icmp_ignore_bogus_error_responses"
	SYSCTL_IPV4_LOG_MARTIANS             = "net.ipv4.conf.all.log_martians"
	SYSCTL_RANDOMISE_VA_SPACE            = "kernel.randomize_va_space"
	SYSCTL_KPTR_RESTRICT                 = "kernel.kptr_restrict"
	SYSCTL_PROTECTED_HARDLINKS           = "fs.protected_hardlinks"
	SYSCTL_PROTECTED_SYMLINKS            = "fs.protected_symlinks"
	SYSCTL_SCHED_CHILD_RUNS_FIRST        = "kernel.sched_child_runs_first"
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
