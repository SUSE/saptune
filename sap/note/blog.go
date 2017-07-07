package note

import (
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"path"
)

/*
SUSE-GUIDE-01 - SLES 11/12 OS Tuning & Optimization Guide – Part 1
https://www.suse.com/communities/blog/sles-1112-os-tuning-optimisation-guide-part-1/
*/
type SUSESysOptimisation struct {
	SysconfigPrefix string

	// Section "SLES Memory Tuning and Optimization"
	VMNumberHugePages, VMSwappiness, VMVfsCachePressure uint64
	VMOvercommitMemory, VMOvercommitRatio               uint64

	// Section "SLES Disk I/O & Storage Tuning Optimization"
	VMDirtyRatio, VMDirtyBackgroundRatio uint64
	BlockDeviceSchedulers                param.BlockDeviceSchedulers
}

func (st SUSESysOptimisation) Name() string {
	// Do not mention SLES 11 here
	return "SLES 12 OS Tuning & Optimization Guide – Part 1"
}
func (st SUSESysOptimisation) Initialise() (Note, error) {
	newST := st
	newST.VMNumberHugePages = system.GetSysctlUint64(system.SysctlNumberHugepages, 0)
	newST.VMSwappiness = system.GetSysctlUint64(system.SysctlSwappines, 0)
	newST.VMVfsCachePressure = system.GetSysctlUint64(system.SysctlVFSCachePressure, 0)

	newST.VMOvercommitMemory = system.GetSysctlUint64(system.SysctlOvercommitMemory, 0)
	newST.VMOvercommitRatio = system.GetSysctlUint64(system.SysctlOvercommitRatio, 0)
	newST.VMDirtyRatio = system.GetSysctlUint64(system.SysctlDirtyRatio, 0)

	newST.VMDirtyBackgroundRatio = system.GetSysctlUint64(system.SysctlDirtyBackgroundRatio, 0)
	newBlkSchedulers, err := newST.BlockDeviceSchedulers.Inspect()
	if err != nil {
		newST.BlockDeviceSchedulers = newBlkSchedulers.(param.BlockDeviceSchedulers)
	}
	return newST, err
}
func (st SUSESysOptimisation) Optimise() (Note, error) {
	newST := st
	// Parse the switches
	conf, err := txtparser.ParseSysconfigFile(path.Join(newST.SysconfigPrefix, "/etc/sysconfig/saptune-note-SUSE-GUIDE-01"), false)
	if err != nil {
		return nil, err
	}
	if conf.GetBool("TUNE_NUMBER_HUGEPAGES", false) {
		newST.VMNumberHugePages = param.MaxU64(newST.VMNumberHugePages, 128)
	}
	if conf.GetBool("TUNE_SWAPPINESS", false) {
		newST.VMSwappiness = param.MinU64(newST.VMSwappiness, 25)
	}
	if conf.GetBool("TUNE_VFS_CACHE_PRESSURE", false) {
		newST.VMVfsCachePressure = param.MinU64(newST.VMVfsCachePressure, 50)
	}
	if conf.GetBool("TUNE_OVERCOMMIT", false) {
		newST.VMOvercommitMemory = 1
		newST.VMOvercommitRatio = param.MaxU64(newST.VMOvercommitRatio, 70)
	}
	if conf.GetBool("TUNE_DIRTY_RATIO", false) {
		newST.VMDirtyRatio = param.MinU64(newST.VMDirtyRatio, 10)
		newST.VMDirtyBackgroundRatio = param.MinU64(newST.VMDirtyRatio, 5)
	}
	if conf.GetBool("TUNE_IO_SCHEDULER", false) {
		for blk := range newST.BlockDeviceSchedulers.SchedulerChoice {
			newST.BlockDeviceSchedulers.SchedulerChoice[blk] = "noop"
		}
	}
	return newST, nil
}
func (st SUSESysOptimisation) Apply() error {
	system.SetSysctlUint64(system.SysctlNumberHugepages, st.VMNumberHugePages)
	system.SetSysctlUint64(system.SysctlSwappines, st.VMSwappiness)
	system.SetSysctlUint64(system.SysctlVFSCachePressure, st.VMVfsCachePressure)

	system.SetSysctlUint64(system.SysctlOvercommitMemory, st.VMOvercommitMemory)
	system.SetSysctlUint64(system.SysctlOvercommitRatio, st.VMOvercommitRatio)
	system.SetSysctlUint64(system.SysctlDirtyRatio, st.VMDirtyRatio)

	system.SetSysctlUint64(system.SysctlDirtyBackgroundRatio, st.VMDirtyBackgroundRatio)
	return st.BlockDeviceSchedulers.Apply()
}

/*
SUSE-GUIDE-01 - SLES 11/12: Network, CPU Tuning and Optimization – Part 2
https://www.suse.com/communities/blog/sles-1112-network-cpu-tuning-optimization-part-2/
*/
type SUSENetCPUOptimisation struct {
	SysconfigPrefix string

	// Section "SLES11/12 Network Tuning & Optimization"
	NetCoreWmemMax, NetCoreRmemMax                                   uint64
	NetCoreNetdevMaxBacklog, NetCoreSoMaxConn                        uint64
	NetIpv4TcpRmem, NetIpv4TcpWmem                                   uint64
	NetIpv4TcpTimestamps, NetIpv4TcpSack                             uint64
	NetIpv4TcpFack, NetIpv4TcpDsack                                  uint64
	NetIpv4IpfragLowThres, NetIpv4IpfragHighThres                    uint64
	NetIpv4TcpMaxSynBacklog, NetIpv4TcpSynackRetries                 uint64
	NetIpv4TcpRetries2                                               uint64
	NetTcpKeepaliveTime, NetTcpKeepaliveProbes, NetTcpKeepaliveIntvl uint64
	NetTcpTwRecycle, NetTcpTwReuse, NetTcpFinTimeout                 uint64
	NetTcpMtuProbing                                                 uint64
	// Section "Basic TCP/IP Optimization for SLES
	NetIpv4TcpSyncookies, NetIpv4ConfAllAcceptSourceRoute                 uint64
	NetIpv4ConfAllAcceptRedirects, NetIpv4ConfAllRPFilter                 uint64
	NetIpv4IcmpEchoIgnoreBroadcasts, NetIpv4IcmpIgnoreBogusErrorResponses uint64
	NetIpv4ConfAllLogMartians                                             uint64
	KernelRandomizeVASpace, KernelKptrRestrict                            uint64
	FSProtectedHardlinks, FSProtectedSymlinks                             uint64
	KernelSchedChildRunsFirst                                             uint64
}

func (st SUSENetCPUOptimisation) Name() string {
	// Do not mention SLES 11 here
	return "SLES 12: Network, CPU Tuning and Optimization – Part 2"
}
func (st SUSENetCPUOptimisation) Initialise() (Note, error) {
	newST := st
	// Section "SLES11/12 Network Tuning & Optimization"
	newST.NetCoreWmemMax = system.GetSysctlUint64(system.SysctlNetWriteMemMax, 0)
	newST.NetCoreRmemMax = system.GetSysctlUint64(system.SysctlNetReadMemMax, 0)
	newST.NetCoreNetdevMaxBacklog = system.GetSysctlUint64(system.SysctlNetMaxBacklog, 0)
	newST.NetCoreSoMaxConn = system.GetSysctlUint64(system.SysctlNetMaxconn, 0)
	newST.NetIpv4TcpRmem = system.GetSysctlUint64Field(system.SysctlTCPReadMem, 2, 0)
	newST.NetIpv4TcpWmem = system.GetSysctlUint64Field(system.SysctlTCPWriteMem, 2, 0)
	newST.NetIpv4TcpTimestamps = system.GetSysctlUint64(system.SysctlTCPTimestamps, 0)
	newST.NetIpv4TcpSack = system.GetSysctlUint64(system.SysctlTCPSack, 0)
	newST.NetIpv4TcpFack = system.GetSysctlUint64(system.SysctlTCPFack, 0)
	newST.NetIpv4TcpDsack = system.GetSysctlUint64(system.SysctlTCPDsack, 0)
	newST.NetIpv4IpfragLowThres = system.GetSysctlUint64(system.SysctlTCPFragLowThreshold, 0)
	newST.NetIpv4IpfragHighThres = system.GetSysctlUint64(system.SysctlTCPFragHighThreshold, 0)
	newST.NetIpv4TcpMaxSynBacklog = system.GetSysctlUint64(system.SysctlTCPMaxSynBacklog, 0)
	newST.NetIpv4TcpSynackRetries = system.GetSysctlUint64(system.SysctlTCPSynackRetries, 0)
	newST.NetIpv4TcpRetries2 = system.GetSysctlUint64(system.SysctpTCPRetries2, 0)
	newST.NetTcpKeepaliveTime = system.GetSysctlUint64(system.SysctlTCPKeepaliveTime, 0)
	newST.NetTcpKeepaliveProbes = system.GetSysctlUint64(system.SysctlTCPKeepaliveProbes, 0)
	newST.NetTcpKeepaliveIntvl = system.GetSysctlUint64(system.SysctlTCPKeepaliveInterval, 0)
	newST.NetTcpTwRecycle = system.GetSysctlUint64(system.SysctlTCPTWRecycle, 0)
	newST.NetTcpTwReuse = system.GetSysctlUint64(system.SysctlTCPTWReuse, 0)
	newST.NetTcpFinTimeout = system.GetSysctlUint64(system.SysctlTCPFinTimeout, 0)
	newST.NetTcpMtuProbing = system.GetSysctlUint64(system.SysctlTCPMTUProbing, 0)

	// Section "Basic TCP/IP Optimization for SLES
	newST.NetIpv4TcpSyncookies = system.GetSysctlUint64(system.SysctlTCPSynCookies, 0)
	newST.NetIpv4ConfAllAcceptSourceRoute = system.GetSysctlUint64(system.SysctlIPAcceptSourceRoute, 0)
	newST.NetIpv4ConfAllAcceptRedirects = system.GetSysctlUint64(system.SysctlIPAcceptRedirects, 0)
	newST.NetIpv4ConfAllRPFilter = system.GetSysctlUint64(system.SysctlIPRPFilter, 0)
	newST.NetIpv4IcmpEchoIgnoreBroadcasts = system.GetSysctlUint64(system.SysctlIPIgnoreICMPBroadcasts, 0)
	newST.NetIpv4IcmpIgnoreBogusErrorResponses = system.GetSysctlUint64(system.SysctlIPIgnoreICMPBogusError, 0)
	newST.NetIpv4ConfAllLogMartians = system.GetSysctlUint64(system.SysctlIPLogMartians, 0)
	newST.KernelRandomizeVASpace = system.GetSysctlUint64(system.SysctlRandomizeVASpace, 0)
	newST.KernelKptrRestrict = system.GetSysctlUint64(system.SysctlKptrRestrict, 0)
	newST.FSProtectedHardlinks = system.GetSysctlUint64(system.SysctlProtectHardlinks, 0)
	newST.FSProtectedSymlinks = system.GetSysctlUint64(system.SysctlProtectSymlinks, 0)
	newST.KernelSchedChildRunsFirst = system.GetSysctlUint64(system.SysctlRunChildFirst, 0)
	return newST, nil
}
func (st SUSENetCPUOptimisation) Optimise() (Note, error) {
	newST := st
	conf, err := txtparser.ParseSysconfigFile(path.Join(newST.SysconfigPrefix, "/etc/sysconfig/saptune-note-SUSE-GUIDE-02"), false)
	if err != nil {
		return nil, err
	}
	// Section "SLES11/12 Network Tuning & Optimization"
	if conf.GetBool("TUNE_NET_RESERVED_SOCKETS", false) {
		newST.NetCoreWmemMax = param.MaxU64(newST.NetCoreWmemMax, 12582912)
		newST.NetCoreRmemMax = param.MaxU64(newST.NetCoreRmemMax, 12582912)
	}
	if conf.GetBool("TUNE_NET_QUEUE_SIZE", false) {
		newST.NetCoreNetdevMaxBacklog = param.MaxU64(newST.NetCoreNetdevMaxBacklog, 9000)
		newST.NetCoreSoMaxConn = param.MaxU64(newST.NetCoreSoMaxConn, 512)
	}
	if conf.GetBool("TUNE_TCP_BUFFER_SIZE", false) {
		newST.NetIpv4TcpRmem = param.MaxU64(newST.NetIpv4TcpRmem, 9437184)
		newST.NetIpv4TcpWmem = param.MaxU64(newST.NetIpv4TcpWmem, 9437184)
	}
	if conf.GetBool("TUNE_TCP_TIMESTAMPS", false) {
		newST.NetIpv4TcpTimestamps = 0
	}
	if conf.GetBool("TUNE_TCP_ACK_BEHAVIOUR", false) {
		newST.NetIpv4TcpSack = 0
		newST.NetIpv4TcpDsack = 0
		newST.NetIpv4TcpFack = 0
	}
	if conf.GetBool("TUNE_IP_FRAGMENTATION", false) {
		newST.NetIpv4IpfragHighThres = param.MaxU64(newST.NetIpv4IpfragHighThres, 544288)
		newST.NetIpv4IpfragLowThres = param.MaxU64(newST.NetIpv4IpfragLowThres, 393216)
	}
	if conf.GetBool("TUNE_TCP_SYN_QUEUE", false) {
		newST.NetIpv4TcpMaxSynBacklog = param.MaxU64(newST.NetIpv4TcpMaxSynBacklog, 8192)
	}
	if conf.GetBool("TUNE_TCP_RETRY_BEHAVIOUR", false) {
		newST.NetIpv4TcpSynackRetries = param.MinU64(newST.NetIpv4TcpSynackRetries, 3)
		newST.NetIpv4TcpRetries2 = param.MinU64(newST.NetIpv4TcpRetries2, 6)
	}
	if conf.GetBool("TUNE_TCP_KEEPALIVE_BEHAVIOUR", false) {
		newST.NetTcpKeepaliveTime = param.MinU64(newST.NetTcpKeepaliveTime, 1000)
		newST.NetTcpKeepaliveProbes = param.MinU64(newST.NetTcpKeepaliveProbes, 4)
		newST.NetTcpKeepaliveIntvl = param.MinU64(newST.NetTcpKeepaliveIntvl, 20)
	}
	if conf.GetBool("TUNE_TCP_TIME_WAIT_BEHAVIOUR", false) {
		newST.NetTcpTwRecycle = 1
		newST.NetTcpTwReuse = 1
	}
	if conf.GetBool("TUNE_TCP_FIN_TIMEOUT", false) {
		newST.NetTcpFinTimeout = param.MinU64(newST.NetTcpFinTimeout, 30)
	}
	if conf.GetBool("TUNE_JUMBO_FRAME_MTU_PROBING", false) {
		newST.NetTcpMtuProbing = 1
	}
	// Section "Basic TCP/IP Optimization for SLES
	if conf.GetBool("TUNE_SECURITY", false) {
		newST.NetIpv4TcpSyncookies = 1
		newST.NetIpv4ConfAllAcceptSourceRoute = 0
		newST.NetIpv4ConfAllAcceptRedirects = 0
		newST.NetIpv4ConfAllRPFilter = 1
		newST.NetIpv4IcmpEchoIgnoreBroadcasts = 1
		newST.NetIpv4IcmpIgnoreBogusErrorResponses = 1
		newST.NetIpv4ConfAllLogMartians = 1
		newST.KernelRandomizeVASpace = 2
		newST.KernelKptrRestrict = 1
		newST.FSProtectedHardlinks = 1
		newST.FSProtectedSymlinks = 1
	}
	if conf.GetBool("TUNE_PROCESS_SCHEDULER", false) {
		newST.KernelSchedChildRunsFirst = 1
	}
	return newST, nil
}
func (st SUSENetCPUOptimisation) Apply() error {
	// Section "SLES11/12 Network Tuning & Optimization"
	system.SetSysctlUint64(system.SysctlNetWriteMemMax, st.NetCoreWmemMax)
	system.SetSysctlUint64(system.SysctlNetReadMemMax, st.NetCoreRmemMax)

	system.SetSysctlUint64(system.SysctlNetMaxBacklog, st.NetCoreNetdevMaxBacklog)
	system.SetSysctlUint64(system.SysctlNetMaxconn, st.NetCoreSoMaxConn)

	system.SetSysctlUint64Field(system.SysctlTCPReadMem, 2, st.NetIpv4TcpRmem)
	system.SetSysctlUint64Field(system.SysctlTCPWriteMem, 2, st.NetIpv4TcpWmem)

	system.SetSysctlUint64(system.SysctlTCPTimestamps, st.NetIpv4TcpTimestamps)

	system.SetSysctlUint64(system.SysctlTCPSack, st.NetIpv4TcpSack)
	system.SetSysctlUint64(system.SysctlTCPFack, st.NetIpv4TcpFack)
	system.SetSysctlUint64(system.SysctlTCPDsack, st.NetIpv4TcpDsack)

	system.SetSysctlUint64(system.SysctlTCPFragLowThreshold, st.NetIpv4IpfragLowThres)
	system.SetSysctlUint64(system.SysctlTCPFragHighThreshold, st.NetIpv4IpfragHighThres)

	system.SetSysctlUint64(system.SysctlTCPMaxSynBacklog, st.NetIpv4TcpMaxSynBacklog)
	system.SetSysctlUint64(system.SysctlTCPSynackRetries, st.NetIpv4TcpSynackRetries)
	system.SetSysctlUint64(system.SysctpTCPRetries2, st.NetIpv4TcpRetries2)

	system.SetSysctlUint64(system.SysctlTCPKeepaliveTime, st.NetTcpKeepaliveTime)
	system.SetSysctlUint64(system.SysctlTCPKeepaliveProbes, st.NetTcpKeepaliveProbes)
	system.SetSysctlUint64(system.SysctlTCPKeepaliveInterval, st.NetTcpKeepaliveIntvl)

	system.SetSysctlUint64(system.SysctlTCPTWRecycle, st.NetTcpTwRecycle)
	system.SetSysctlUint64(system.SysctlTCPTWReuse, st.NetTcpTwReuse)
	system.SetSysctlUint64(system.SysctlTCPFinTimeout, st.NetTcpFinTimeout)

	system.SetSysctlUint64(system.SysctlTCPMTUProbing, st.NetTcpMtuProbing)

	// Section "Basic TCP/IP Optimization for SLES
	system.SetSysctlUint64(system.SysctlTCPSynCookies, st.NetIpv4TcpSyncookies)
	system.SetSysctlUint64(system.SysctlIPAcceptSourceRoute, st.NetIpv4ConfAllAcceptSourceRoute)
	system.SetSysctlUint64(system.SysctlIPAcceptRedirects, st.NetIpv4ConfAllAcceptRedirects)
	system.SetSysctlUint64(system.SysctlIPRPFilter, st.NetIpv4ConfAllRPFilter)

	system.SetSysctlUint64(system.SysctlIPIgnoreICMPBroadcasts, st.NetIpv4IcmpEchoIgnoreBroadcasts)
	system.SetSysctlUint64(system.SysctlIPIgnoreICMPBogusError, st.NetIpv4IcmpIgnoreBogusErrorResponses)
	system.SetSysctlUint64(system.SysctlIPLogMartians, st.NetIpv4ConfAllLogMartians)

	system.SetSysctlUint64(system.SysctlRandomizeVASpace, st.KernelRandomizeVASpace)
	system.SetSysctlUint64(system.SysctlKptrRestrict, st.KernelKptrRestrict)
	system.SetSysctlUint64(system.SysctlProtectHardlinks, st.FSProtectedHardlinks)
	system.SetSysctlUint64(system.SysctlProtectSymlinks, st.FSProtectedSymlinks)

	system.SetSysctlUint64(system.SysctlRunChildFirst, st.KernelSchedChildRunsFirst)
	return nil
}
