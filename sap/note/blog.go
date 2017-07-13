package note

import (
	"github.com/HouzuoGuo/saptune/sap"
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
	newST.VMNumberHugePages, _ = system.GetSysctlUint64(system.SysctlNumberHugepages)
	newST.VMSwappiness, _ = system.GetSysctlUint64(system.SysctlSwappines)
	newST.VMVfsCachePressure, _ = system.GetSysctlUint64(system.SysctlVFSCachePressure)

	newST.VMOvercommitMemory, _ = system.GetSysctlUint64(system.SysctlOvercommitMemory)
	newST.VMOvercommitRatio, _ = system.GetSysctlUint64(system.SysctlOvercommitRatio)
	newST.VMDirtyRatio, _ = system.GetSysctlUint64(system.SysctlDirtyRatio)

	newST.VMDirtyBackgroundRatio, _ = system.GetSysctlUint64(system.SysctlDirtyBackgroundRatio)
	newBlkSchedulers, err := newST.BlockDeviceSchedulers.Inspect()
	if err != nil {
		newST.BlockDeviceSchedulers = newBlkSchedulers.(param.BlockDeviceSchedulers)
	}
	return newST, nil
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
	errs := make([]error, 0, 0)
	errs = append(errs, system.SetSysctlUint64(system.SysctlNumberHugepages, st.VMNumberHugePages))
	errs = append(errs, system.SetSysctlUint64(system.SysctlSwappines, st.VMSwappiness))
	errs = append(errs, system.SetSysctlUint64(system.SysctlVFSCachePressure, st.VMVfsCachePressure))

	errs = append(errs, system.SetSysctlUint64(system.SysctlOvercommitMemory, st.VMOvercommitMemory))
	errs = append(errs, system.SetSysctlUint64(system.SysctlOvercommitRatio, st.VMOvercommitRatio))
	errs = append(errs, system.SetSysctlUint64(system.SysctlDirtyRatio, st.VMDirtyRatio))
	errs = append(errs, system.SetSysctlUint64(system.SysctlDirtyBackgroundRatio, st.VMDirtyBackgroundRatio))

	errs = append(errs, st.BlockDeviceSchedulers.Apply())

	err := sap.PrintErrors(errs)
	return err
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
	newST.NetCoreWmemMax, _ = system.GetSysctlUint64(system.SysctlNetWriteMemMax)
	newST.NetCoreRmemMax, _ = system.GetSysctlUint64(system.SysctlNetReadMemMax)
	newST.NetCoreNetdevMaxBacklog, _ = system.GetSysctlUint64(system.SysctlNetMaxBacklog)
	newST.NetCoreSoMaxConn, _ = system.GetSysctlUint64(system.SysctlNetMaxconn)
	newST.NetIpv4TcpRmem, _ = system.GetSysctlUint64Field(system.SysctlTCPReadMem, 2)
	newST.NetIpv4TcpWmem, _ = system.GetSysctlUint64Field(system.SysctlTCPWriteMem, 2)
	newST.NetIpv4TcpTimestamps, _ = system.GetSysctlUint64(system.SysctlTCPTimestamps)
	newST.NetIpv4TcpSack, _ = system.GetSysctlUint64(system.SysctlTCPSack)
	newST.NetIpv4TcpFack, _ = system.GetSysctlUint64(system.SysctlTCPFack)
	newST.NetIpv4TcpDsack, _ = system.GetSysctlUint64(system.SysctlTCPDsack)
	newST.NetIpv4IpfragLowThres, _ = system.GetSysctlUint64(system.SysctlTCPFragLowThreshold)
	newST.NetIpv4IpfragHighThres, _ = system.GetSysctlUint64(system.SysctlTCPFragHighThreshold)
	newST.NetIpv4TcpMaxSynBacklog, _ = system.GetSysctlUint64(system.SysctlTCPMaxSynBacklog)
	newST.NetIpv4TcpSynackRetries, _ = system.GetSysctlUint64(system.SysctlTCPSynackRetries)
	newST.NetIpv4TcpRetries2, _ = system.GetSysctlUint64(system.SysctpTCPRetries2)
	newST.NetTcpKeepaliveTime, _ = system.GetSysctlUint64(system.SysctlTCPKeepaliveTime)
	newST.NetTcpKeepaliveProbes, _ = system.GetSysctlUint64(system.SysctlTCPKeepaliveProbes)
	newST.NetTcpKeepaliveIntvl, _ = system.GetSysctlUint64(system.SysctlTCPKeepaliveInterval)
	newST.NetTcpTwRecycle, _ = system.GetSysctlUint64(system.SysctlTCPTWRecycle)
	newST.NetTcpTwReuse, _ = system.GetSysctlUint64(system.SysctlTCPTWReuse)
	newST.NetTcpFinTimeout, _ = system.GetSysctlUint64(system.SysctlTCPFinTimeout)
	newST.NetTcpMtuProbing, _ = system.GetSysctlUint64(system.SysctlTCPMTUProbing)

	// Section "Basic TCP/IP Optimization for SLES
	newST.NetIpv4TcpSyncookies, _ = system.GetSysctlUint64(system.SysctlTCPSynCookies)
	newST.NetIpv4ConfAllAcceptSourceRoute, _ = system.GetSysctlUint64(system.SysctlIPAcceptSourceRoute)
	newST.NetIpv4ConfAllAcceptRedirects, _ = system.GetSysctlUint64(system.SysctlIPAcceptRedirects)
	newST.NetIpv4ConfAllRPFilter, _ = system.GetSysctlUint64(system.SysctlIPRPFilter)
	newST.NetIpv4IcmpEchoIgnoreBroadcasts, _ = system.GetSysctlUint64(system.SysctlIPIgnoreICMPBroadcasts)
	newST.NetIpv4IcmpIgnoreBogusErrorResponses, _ = system.GetSysctlUint64(system.SysctlIPIgnoreICMPBogusError)
	newST.NetIpv4ConfAllLogMartians, _ = system.GetSysctlUint64(system.SysctlIPLogMartians)
	newST.KernelRandomizeVASpace, _ = system.GetSysctlUint64(system.SysctlRandomizeVASpace)
	newST.KernelKptrRestrict, _ = system.GetSysctlUint64(system.SysctlKptrRestrict)
	newST.FSProtectedHardlinks, _ = system.GetSysctlUint64(system.SysctlProtectHardlinks)
	newST.FSProtectedSymlinks, _ = system.GetSysctlUint64(system.SysctlProtectSymlinks)
	newST.KernelSchedChildRunsFirst, _ = system.GetSysctlUint64(system.SysctlRunChildFirst)
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
	errs := make([]error, 0, 0)
	errs = append(errs, system.SetSysctlUint64(system.SysctlNetWriteMemMax, st.NetCoreWmemMax))
	errs = append(errs, system.SetSysctlUint64(system.SysctlNetReadMemMax, st.NetCoreRmemMax))

	errs = append(errs, system.SetSysctlUint64(system.SysctlNetMaxBacklog, st.NetCoreNetdevMaxBacklog))
	errs = append(errs, system.SetSysctlUint64(system.SysctlNetMaxconn, st.NetCoreSoMaxConn))

	errs = append(errs, system.SetSysctlUint64Field(system.SysctlTCPReadMem, 2, st.NetIpv4TcpRmem))
	errs = append(errs, system.SetSysctlUint64Field(system.SysctlTCPWriteMem, 2, st.NetIpv4TcpWmem))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPTimestamps, st.NetIpv4TcpTimestamps))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPSack, st.NetIpv4TcpSack))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPFack, st.NetIpv4TcpFack))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPDsack, st.NetIpv4TcpDsack))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPFragLowThreshold, st.NetIpv4IpfragLowThres))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPFragHighThreshold, st.NetIpv4IpfragHighThres))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPMaxSynBacklog, st.NetIpv4TcpMaxSynBacklog))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPSynackRetries, st.NetIpv4TcpSynackRetries))
	errs = append(errs, system.SetSysctlUint64(system.SysctpTCPRetries2, st.NetIpv4TcpRetries2))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPKeepaliveTime, st.NetTcpKeepaliveTime))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPKeepaliveProbes, st.NetTcpKeepaliveProbes))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPKeepaliveInterval, st.NetTcpKeepaliveIntvl))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPTWRecycle, st.NetTcpTwRecycle))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPTWReuse, st.NetTcpTwReuse))
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPFinTimeout, st.NetTcpFinTimeout))

	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPMTUProbing, st.NetTcpMtuProbing))

	// Section "Basic TCP/IP Optimization for SLES
	errs = append(errs, system.SetSysctlUint64(system.SysctlTCPSynCookies, st.NetIpv4TcpSyncookies))
	errs = append(errs, system.SetSysctlUint64(system.SysctlIPAcceptSourceRoute, st.NetIpv4ConfAllAcceptSourceRoute))
	errs = append(errs, system.SetSysctlUint64(system.SysctlIPAcceptRedirects, st.NetIpv4ConfAllAcceptRedirects))
	errs = append(errs, system.SetSysctlUint64(system.SysctlIPRPFilter, st.NetIpv4ConfAllRPFilter))

	errs = append(errs, system.SetSysctlUint64(system.SysctlIPIgnoreICMPBroadcasts, st.NetIpv4IcmpEchoIgnoreBroadcasts))
	errs = append(errs, system.SetSysctlUint64(system.SysctlIPIgnoreICMPBogusError, st.NetIpv4IcmpIgnoreBogusErrorResponses))
	errs = append(errs, system.SetSysctlUint64(system.SysctlIPLogMartians, st.NetIpv4ConfAllLogMartians))

	errs = append(errs, system.SetSysctlUint64(system.SysctlRandomizeVASpace, st.KernelRandomizeVASpace))
	errs = append(errs, system.SetSysctlUint64(system.SysctlKptrRestrict, st.KernelKptrRestrict))
	errs = append(errs, system.SetSysctlUint64(system.SysctlProtectHardlinks, st.FSProtectedHardlinks))
	errs = append(errs, system.SetSysctlUint64(system.SysctlProtectSymlinks, st.FSProtectedSymlinks))

	errs = append(errs, system.SetSysctlUint64(system.SysctlRunChildFirst, st.KernelSchedChildRunsFirst))

	err := sap.PrintErrors(errs)
	return err
}
