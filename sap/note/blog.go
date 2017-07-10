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
	var err error
	newST := st
	newST.VMNumberHugePages, err = system.GetSysctlUint64(system.SysctlNumberHugepages)
	if err != nil {
		return nil, err
	}
	newST.VMSwappiness, err = system.GetSysctlUint64(system.SysctlSwappines)
	if err != nil {
		return nil, err
	}
	newST.VMVfsCachePressure, err = system.GetSysctlUint64(system.SysctlVFSCachePressure)
	if err != nil {
		return nil, err
	}

	newST.VMOvercommitMemory, err = system.GetSysctlUint64(system.SysctlOvercommitMemory)
	if err != nil {
		return nil, err
	}
	newST.VMOvercommitRatio, err = system.GetSysctlUint64(system.SysctlOvercommitRatio)
	if err != nil {
		return nil, err
	}
	newST.VMDirtyRatio, err = system.GetSysctlUint64(system.SysctlDirtyRatio)
	if err != nil {
		return nil, err
	}

	newST.VMDirtyBackgroundRatio, err = system.GetSysctlUint64(system.SysctlDirtyBackgroundRatio)
	if err != nil {
		return nil, err
	}
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
	err := system.SetSysctlUint64(system.SysctlNumberHugepages, st.VMNumberHugePages)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlSwappines, st.VMSwappiness)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlVFSCachePressure, st.VMVfsCachePressure)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlOvercommitMemory, st.VMOvercommitMemory)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlOvercommitRatio, st.VMOvercommitRatio)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlDirtyRatio, st.VMDirtyRatio)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlDirtyBackgroundRatio, st.VMDirtyBackgroundRatio)
	if err != nil {
		return err
	}
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
	var err error
	// Section "SLES11/12 Network Tuning & Optimization"
	newST.NetCoreWmemMax, err = system.GetSysctlUint64(system.SysctlNetWriteMemMax)
	if err != nil {
		return nil, err
	}
	newST.NetCoreRmemMax, err = system.GetSysctlUint64(system.SysctlNetReadMemMax)
	if err != nil {
		return nil, err
	}
	newST.NetCoreNetdevMaxBacklog, err = system.GetSysctlUint64(system.SysctlNetMaxBacklog)
	if err != nil {
		return nil, err
	}
	newST.NetCoreSoMaxConn, err = system.GetSysctlUint64(system.SysctlNetMaxconn)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpRmem, err = system.GetSysctlUint64Field(system.SysctlTCPReadMem, 2)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpWmem, err = system.GetSysctlUint64Field(system.SysctlTCPWriteMem, 2)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpTimestamps, err = system.GetSysctlUint64(system.SysctlTCPTimestamps)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpSack, err = system.GetSysctlUint64(system.SysctlTCPSack)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpFack, err = system.GetSysctlUint64(system.SysctlTCPFack)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpDsack, err = system.GetSysctlUint64(system.SysctlTCPDsack)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4IpfragLowThres, err = system.GetSysctlUint64(system.SysctlTCPFragLowThreshold)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4IpfragHighThres, err = system.GetSysctlUint64(system.SysctlTCPFragHighThreshold)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpMaxSynBacklog, err = system.GetSysctlUint64(system.SysctlTCPMaxSynBacklog)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpSynackRetries, err = system.GetSysctlUint64(system.SysctlTCPSynackRetries)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4TcpRetries2, err = system.GetSysctlUint64(system.SysctpTCPRetries2)
	if err != nil {
		return nil, err
	}
	newST.NetTcpKeepaliveTime, err = system.GetSysctlUint64(system.SysctlTCPKeepaliveTime)
	if err != nil {
		return nil, err
	}
	newST.NetTcpKeepaliveProbes, err = system.GetSysctlUint64(system.SysctlTCPKeepaliveProbes)
	if err != nil {
		return nil, err
	}
	newST.NetTcpKeepaliveIntvl, err = system.GetSysctlUint64(system.SysctlTCPKeepaliveInterval)
	if err != nil {
		return nil, err
	}
	newST.NetTcpTwRecycle, err = system.GetSysctlUint64(system.SysctlTCPTWRecycle)
	if err != nil {
		return nil, err
	}
	newST.NetTcpTwReuse, err = system.GetSysctlUint64(system.SysctlTCPTWReuse)
	if err != nil {
		return nil, err
	}
	newST.NetTcpFinTimeout, err = system.GetSysctlUint64(system.SysctlTCPFinTimeout)
	if err != nil {
		return nil, err
	}
	newST.NetTcpMtuProbing, err = system.GetSysctlUint64(system.SysctlTCPMTUProbing)
	if err != nil {
		return nil, err
	}

	// Section "Basic TCP/IP Optimization for SLES
	newST.NetIpv4TcpSyncookies, err = system.GetSysctlUint64(system.SysctlTCPSynCookies)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4ConfAllAcceptSourceRoute, err = system.GetSysctlUint64(system.SysctlIPAcceptSourceRoute)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4ConfAllAcceptRedirects, err = system.GetSysctlUint64(system.SysctlIPAcceptRedirects)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4ConfAllRPFilter, err = system.GetSysctlUint64(system.SysctlIPRPFilter)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4IcmpEchoIgnoreBroadcasts, err = system.GetSysctlUint64(system.SysctlIPIgnoreICMPBroadcasts)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4IcmpIgnoreBogusErrorResponses, err = system.GetSysctlUint64(system.SysctlIPIgnoreICMPBogusError)
	if err != nil {
		return nil, err
	}
	newST.NetIpv4ConfAllLogMartians, err = system.GetSysctlUint64(system.SysctlIPLogMartians)
	if err != nil {
		return nil, err
	}
	newST.KernelRandomizeVASpace, err = system.GetSysctlUint64(system.SysctlRandomizeVASpace)
	if err != nil {
		return nil, err
	}
	newST.KernelKptrRestrict, err = system.GetSysctlUint64(system.SysctlKptrRestrict)
	if err != nil {
		return nil, err
	}
	newST.FSProtectedHardlinks, err = system.GetSysctlUint64(system.SysctlProtectHardlinks)
	if err != nil {
		return nil, err
	}
	newST.FSProtectedSymlinks, err = system.GetSysctlUint64(system.SysctlProtectSymlinks)
	if err != nil {
		return nil, err
	}
	newST.KernelSchedChildRunsFirst, err = system.GetSysctlUint64(system.SysctlRunChildFirst)
	if err != nil {
		return nil, err
	}
	return newST, err
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
	err := system.SetSysctlUint64(system.SysctlNetWriteMemMax, st.NetCoreWmemMax)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlNetReadMemMax, st.NetCoreRmemMax)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlNetMaxBacklog, st.NetCoreNetdevMaxBacklog)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlNetMaxconn, st.NetCoreSoMaxConn)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64Field(system.SysctlTCPReadMem, 2, st.NetIpv4TcpRmem)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64Field(system.SysctlTCPWriteMem, 2, st.NetIpv4TcpWmem)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPTimestamps, st.NetIpv4TcpTimestamps)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPSack, st.NetIpv4TcpSack)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPFack, st.NetIpv4TcpFack)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPDsack, st.NetIpv4TcpDsack)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPFragLowThreshold, st.NetIpv4IpfragLowThres)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPFragHighThreshold, st.NetIpv4IpfragHighThres)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPMaxSynBacklog, st.NetIpv4TcpMaxSynBacklog)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPSynackRetries, st.NetIpv4TcpSynackRetries)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctpTCPRetries2, st.NetIpv4TcpRetries2)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPKeepaliveTime, st.NetTcpKeepaliveTime)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPKeepaliveProbes, st.NetTcpKeepaliveProbes)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPKeepaliveInterval, st.NetTcpKeepaliveIntvl)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPTWRecycle, st.NetTcpTwRecycle)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPTWReuse, st.NetTcpTwReuse)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlTCPFinTimeout, st.NetTcpFinTimeout)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlTCPMTUProbing, st.NetTcpMtuProbing)
	if err != nil {
		return err
	}

	// Section "Basic TCP/IP Optimization for SLES
	err = system.SetSysctlUint64(system.SysctlTCPSynCookies, st.NetIpv4TcpSyncookies)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlIPAcceptSourceRoute, st.NetIpv4ConfAllAcceptSourceRoute)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlIPAcceptRedirects, st.NetIpv4ConfAllAcceptRedirects)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlIPRPFilter, st.NetIpv4ConfAllRPFilter)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlIPIgnoreICMPBroadcasts, st.NetIpv4IcmpEchoIgnoreBroadcasts)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlIPIgnoreICMPBogusError, st.NetIpv4IcmpIgnoreBogusErrorResponses)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlIPLogMartians, st.NetIpv4ConfAllLogMartians)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlRandomizeVASpace, st.KernelRandomizeVASpace)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlKptrRestrict, st.KernelKptrRestrict)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlProtectHardlinks, st.FSProtectedHardlinks)
	if err != nil {
		return err
	}
	err = system.SetSysctlUint64(system.SysctlProtectSymlinks, st.FSProtectedSymlinks)
	if err != nil {
		return err
	}

	err = system.SetSysctlUint64(system.SysctlRunChildFirst, st.KernelSchedChildRunsFirst)
	return err
}
