package note

import (
	"fmt"
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

/*
const (
	INISectionSysctl    = "sysctl"
	INISectionVM        = "vm"
	INISectionBlock     = "block"
	INISectionLimits    = "limits"
	SysKernelTHPEnabled = "kernel/mm/transparent_hugepage/enabled"
	SysKSMRun           = "kernel/mm/ksm/run"
)
*/

// No longer active. Only needed for compatibility reasons
// Revert of notes applied by an older saptune version

// 1275776 - Linux: Preparing SLES for SAP environments
type PrepareForSAPEnvironments struct {
	SysconfigPrefix                                         string
	ShmFileSystemSizeMB                                     int64
	LimitNofileSapsysSoft, LimitNofileSapsysHard            system.SecurityLimitInt
	LimitNofileSdbaSoft, LimitNofileSdbaHard                system.SecurityLimitInt
	LimitNofileDbaSoft, LimitNofileDbaHard                  system.SecurityLimitInt
	KernelShmMax, KernelShmAll, KernelShmMni, VMMaxMapCount uint64
	KernelSemMsl, KernelSemMns, KernelSemOpm, KernelSemMni  uint64
}

func (prepare PrepareForSAPEnvironments) Name() string {
	return "Linux: Preparing SLES for SAP environments"
}
func (prepare PrepareForSAPEnvironments) Initialise() (Note, error) {
	newPrepare := prepare
	return newPrepare, nil
}
func (prepare PrepareForSAPEnvironments) Optimise() (Note, error) {
	newPrepare := prepare
	return newPrepare, nil
}
func (prepare PrepareForSAPEnvironments) Apply() error {
	errs := make([]error, 0, 0)
	// Apply new SHM size
	if prepare.ShmFileSystemSizeMB > 0 {
		if err := system.RemountSHM(uint64(prepare.ShmFileSystemSizeMB)); err != nil {
			return err
		}
	} else {
		log.Print("PrepareForSAPEnvironments.Apply: /dev/shm is not a valid mount point, will not adjust its size.")
	}
	// Apply new file descriptor limits
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return err
	}
	secLimits.Set("@sapsys", "soft", "nofile", prepare.LimitNofileSapsysSoft.String())
	secLimits.Set("@sapsys", "hard", "nofile", prepare.LimitNofileSapsysSoft.String())
	secLimits.Set("@sdba", "soft", "nofile", prepare.LimitNofileSdbaSoft.String())
	secLimits.Set("@sdba", "hard", "nofile", prepare.LimitNofileSdbaHard.String())
	secLimits.Set("@dba", "soft", "nofile", prepare.LimitNofileDbaSoft.String())
	secLimits.Set("@dba", "hard", "nofile", prepare.LimitNofileDbaHard.String())
	if err := secLimits.Apply(); err != nil {
		return err
	}
	// Apply shared memory limits
	errs = append(errs, system.SetSysctlUint64(system.SysctlShmax, prepare.KernelShmMax))
	errs = append(errs, system.SetSysctlUint64(system.SysctlShmall, prepare.KernelShmAll))
	errs = append(errs, system.SetSysctlUint64(system.SysctlShmni, prepare.KernelShmMni))
	errs = append(errs, system.SetSysctlUint64(system.SysctlMaxMapCount, prepare.VMMaxMapCount))
	// Apply semaphore limits
	errs = append(errs, system.SetSysctlString(system.SysctlSem, fmt.Sprintf("%d %d %d %d", prepare.KernelSemMsl, prepare.KernelSemMns, prepare.KernelSemOpm, prepare.KernelSemMni)))

	err = sap.PrintErrors(errs)
	return nil
}

// 1984787 - SUSE LINUX Enterprise Server 12: Installation notes
type AfterInstallation struct {
	UuiddSocketStatus bool // UuiddSocketStatus is the status of systemd unit called "uuidd.socket"
	LogindConfigured  bool // LogindConfigured is true if SAP's logind customisation file is in-place
}

func (inst AfterInstallation) Name() string {
	return "SUSE LINUX Enterprise Server 12: Installation notes"
}
func (inst AfterInstallation) Initialise() (Note, error) {
	return AfterInstallation{UuiddSocketStatus: true, LogindConfigured: true}, nil
}
func (inst AfterInstallation) Optimise() (Note, error) {
	return AfterInstallation{UuiddSocketStatus: true, LogindConfigured: true}, nil
}
func (inst AfterInstallation) Apply() error {
	// Set UUID socket status
	var err error
	if inst.UuiddSocketStatus {
		err = system.SystemctlEnableStart("uuidd.socket")
	}
	if err != nil {
		return err
	}
	// Prepare logind config file
	if err := os.MkdirAll(LogindConfDir, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(LogindConfDir, LogindSAPConfFile), []byte(LogindSAPConfContent), 0644); err != nil {
		return err
	}
	if inst.LogindConfigured {
		log.Print("Be aware: system-wide UserTasksMax is now set to infinity according to SAP recommendations.\n" +
			"This opens up entire system to fork-bomb style attacks. Please reboot the system for the changes to take effect.")
	}
	return nil
}

// 2161991 - VMware vSphere (guest) configuration guidelines
type VmwareGuestIOElevator struct {
	BlockDeviceSchedulers param.BlockDeviceSchedulers
}

func (vmio VmwareGuestIOElevator) Name() string {
	return "VMware vSphere (guest) configuration guidelines"
}
func (vmio VmwareGuestIOElevator) Initialise() (Note, error) {
	inspectedParam, err := vmio.BlockDeviceSchedulers.Inspect()
	return VmwareGuestIOElevator{
		BlockDeviceSchedulers: inspectedParam.(param.BlockDeviceSchedulers),
	}, err
}
func (vmio VmwareGuestIOElevator) Optimise() (Note, error) {
	// SAP recommends noop for Vmware guests
	optimisedParam, err := vmio.BlockDeviceSchedulers.Optimise("noop")
	return VmwareGuestIOElevator{
		BlockDeviceSchedulers: optimisedParam.(param.BlockDeviceSchedulers),
	}, err
}
func (vmio VmwareGuestIOElevator) Apply() error {
	return vmio.BlockDeviceSchedulers.Apply()
}

/*
2205917 - SAP HANA DB: Recommended OS settings for SLES 12 / SLES for SAP Applications 12
Disable kernel memory management features that will introduce additional overhead.
*/
type HANARecommendedOSSettings struct {
	KernelMMTransparentHugepage string
	KernelMMKsm                 bool
	KernelNumaBalancing         bool
}

func (hana HANARecommendedOSSettings) Name() string {
	return "SAP HANA DB: Recommended OS settings for SLES 12 / SLES for SAP Applications 12"
}
func (hana HANARecommendedOSSettings) Initialise() (Note, error) {
	ret := HANARecommendedOSSettings{}
	return ret, nil
}
func (hana HANARecommendedOSSettings) Optimise() (Note, error) {
	ret := HANARecommendedOSSettings{}
	return ret, nil
}
func (hana HANARecommendedOSSettings) Apply() error {
	errs := make([]error, 0, 0)
	errs = append(errs, system.SetSysString(SysKernelTHPEnabled, hana.KernelMMTransparentHugepage))
	if hana.KernelMMKsm {
		errs = append(errs, system.SetSysInt(SysKSMRun, 1))
	} else {
		errs = append(errs, system.SetSysInt(SysKSMRun, 0))
	}
	if hana.KernelNumaBalancing {
		errs = append(errs, system.SetSysctlInt(system.SysctlNumaBalancing, 1))
	} else {
		errs = append(errs, system.SetSysctlInt(system.SysctlNumaBalancing, 0))
	}
	err := sap.PrintErrors(errs)
	return err
}

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
	return newST, nil
}
func (st SUSESysOptimisation) Optimise() (Note, error) {
	newST := st
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
	return newST, nil
}
func (st SUSENetCPUOptimisation) Optimise() (Note, error) {
	newST := st
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




// Tuning options composed by a third party vendor.

// section [block]
//type BlockDeviceQueue struct {
	//BlockDeviceSchedulers param.BlockDeviceSchedulers
	//BlockDeviceNrRequests param.BlockDeviceNrRequests
//}

func CmpSetBlkVal(key, value string) error {
	var err error

	switch key {
	case "IO_SCHEDULER":
		setIOQ, err := BlockDeviceQueue{}.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return err
		}

		for _, entry := range strings.Fields(value) {
			fields := strings.Split(entry, "@")
			setIOQ.(param.BlockDeviceSchedulers).SchedulerChoice[fields[0]] = fields[1]
		}
		err = setIOQ.(param.BlockDeviceSchedulers).Apply()
		if err != nil {
			return err
		}
	case "NRREQ":
		setNrR, err := BlockDeviceQueue{}.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return err
		}

		for _, entry := range strings.Fields(value) {
			fields := strings.Split(entry, "@")
			file := path.Join("block", fields[0], "queue", "nr_requests")
			tst_err := system.TestSysString(file, fields[1])
			if tst_err != nil {
				fmt.Printf("Write error on file '%s'.\nCan't set nr_request to '%s', seems to large for the device. Leaving untouched.\n", file, fields[1])
			} else {
				NrR, _ := strconv.Atoi(fields[1])
				setNrR.(param.BlockDeviceNrRequests).NrRequests[fields[0]] = NrR
			}
		}
		err = setNrR.(param.BlockDeviceNrRequests).Apply()
		if err != nil {
			return err
		}
	}
	return err
}

// section [limits]
func CmpSetLimitsVal(key, value string) error {
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return err
	}
	switch key {
	case "MEMLOCK_HARD":
		secLimits.Set("sybase", "hard", "memlock", value)
	case "MEMLOCK_SOFT":
		secLimits.Set("sybase", "soft", "memlock", value)
	}
	err = secLimits.Apply()
	return err
}

// section [vm]
// Manipulate /sys/kernel/mm switches.
// Tuning options composed by a third party vendor.
type CMPTSettings struct {
	ConfFilePath    string            // Full path to the 3rd party vendor's tuning configuration file
	ID              string            // ID portion of the tuning configuration
	DescriptiveName string            // Descriptive name portion of the tuning configuration
	SysctlParams    map[string]string // Sysctl parameter values from the computer system
}

func (cmptvend CMPTSettings) Name() string {
	return cmptvend.DescriptiveName
}

func (cmptvend CMPTSettings) Initialise() (Note, error) {
	cmptvend.SysctlParams = make(map[string]string)
	return cmptvend, nil
}

func (cmptvend CMPTSettings) Optimise() (Note, error) {
	cmptvend.SysctlParams = make(map[string]string)
	return cmptvend, nil
}

func (cmptvend CMPTSettings) Apply() error {
	errs := make([]error, 0, 0)
	// Parse the configuration file
	//ini, err := txtparser.ParseINIFile(cmptvend.ConfFilePath, false)
	// ANGI - TODO
	parseFile := fmt.Sprintf("/var/lib/saptune/saved_conf/%s_n2c.conf", cmptvend.ID)
	if _, err := os.Stat(parseFile); err != nil {
		parseFile = cmptvend.ConfFilePath
	}
	ini, err := txtparser.ParseINIFile(parseFile, false)
	if err != nil {
		return err
	}
	for _, param := range ini.AllValues {
		switch param.Section {
		case INISectionSysctl:
			// Apply sysctl parameters
			errs = append(errs, system.SetSysctlString(param.Key, cmptvend.SysctlParams[param.Key]))
		case INISectionVM:
			errs = append(errs, system.SetSysString(SysKernelTHPEnabled, cmptvend.SysctlParams[param.Key]))
		case INISectionBlock:
			errs = append(errs, CmpSetBlkVal(param.Key, cmptvend.SysctlParams[param.Key]))
		case INISectionLimits:
			errs = append(errs, CmpSetLimitsVal(param.Key, cmptvend.SysctlParams[param.Key]))
		case INISectionReminder:
			continue
		default:
			// saptune does not yet understand settings outside of [sysctl] section
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", cmptvend.ConfFilePath, param.Section)
			continue
		}
	}
	err = sap.PrintErrors(errs)
	return err
}

