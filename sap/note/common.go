package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"log"
	"path"
)

const (
	ARCH_X86 = "amd64"   // GOARCH for 64-bit X86
	ARCH_PPC = "ppc64le" // GOARCH for 64-bit PowerPC little endian
)

// 1275776 - Linux: Preparing SLES for SAP environments
type PrepareForSAPEnvironments struct {
	SysconfigPrefix                                         string
	ShmFileSystemSizeMB                                     int64
	LimitNofileSapsysSoft, LimitNofileSapsysHard            int
	LimitNofileSdbaSoft, LimitNofileSdbaHard                int
	LimitNofileDbaSoft, LimitNofileDbaHard                  int
	KernelShmMax, KernelShmAll, KernelShmMni, VMMaxMapCount uint64
	KernelSemMsl, KernelSemMns, KernelSemOpm, KernelSemMni  uint64
}

func (prepare PrepareForSAPEnvironments) Name() string {
	return "Linux: Preparing SLES for SAP environments"
}
func (prepare PrepareForSAPEnvironments) Initialise() (Note, error) {
	newPrepare := prepare
	// Find out size of SHM
	mount, found := system.ParseProcMounts().GetByMountPoint("/dev/shm")
	if found {
		newPrepare.ShmFileSystemSizeMB = int64(mount.GetFileSystemSizeMB())
	} else {
		log.Print("PrepareForSAPEnvironments.Initialise: failed to find /dev/shm mount point")
		newPrepare.ShmFileSystemSizeMB = -1
	}

	// Find out current file descriptor limits
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return nil, err
	}
	newPrepare.LimitNofileSapsysSoft, _ = secLimits.Get("@sapsys", "soft", "nofile")
	newPrepare.LimitNofileSapsysHard, _ = secLimits.Get("@sapsys", "hard", "nofile")
	newPrepare.LimitNofileSdbaSoft, _ = secLimits.Get("@sdba", "soft", "nofile")
	newPrepare.LimitNofileSdbaHard, _ = secLimits.Get("@sdba", "hard", "nofile")
	newPrepare.LimitNofileDbaSoft, _ = secLimits.Get("@dba", "soft", "nofile")
	newPrepare.LimitNofileDbaHard, _ = secLimits.Get("@dba", "hard", "nofile")
	// Find out shared memory limits
	newPrepare.KernelShmMax = system.GetSysctlUint64(system.SYSCTL_SHMMAX, 0)
	newPrepare.KernelShmAll = system.GetSysctlUint64(system.SYSCTL_SHMALL, 0)
	newPrepare.KernelShmMni = system.GetSysctlUint64(system.SYSCTL_SHMMNI, 0)
	newPrepare.VMMaxMapCount = system.GetSysctlUint64(system.SYSCTL_MAX_MAP_COUNT, 0)
	// Find out semaphore limits
	newPrepare.KernelSemMsl, newPrepare.KernelSemMns, newPrepare.KernelSemOpm, newPrepare.KernelSemMni = system.GetSemaphoreLimits()
	return newPrepare, err
}
func (prepare PrepareForSAPEnvironments) Optimise() (Note, error) {
	newPrepare := prepare

	// Calculate optimal SHM size
	if newPrepare.ShmFileSystemSizeMB > 0 {
		newPrepare.ShmFileSystemSizeMB = param.MaxI64(newPrepare.ShmFileSystemSizeMB, int64(system.GetTotalMemSizeMB())*75/100)
	} else {
		log.Print("PrepareForSAPEnvironments.Optimise: /dev/shm is not a valid mount point, will not calculate its optimal size.")
	}
	// Raise maximum file descriptors to at least 32800
	for _, val := range []*int{&newPrepare.LimitNofileSapsysSoft, &newPrepare.LimitNofileSapsysHard, &newPrepare.LimitNofileSdbaSoft, &newPrepare.LimitNofileSdbaHard, &newPrepare.LimitNofileDbaSoft, &newPrepare.LimitNofileDbaHard} {
		if *val < 32800 {
			*val = 32800
		}
	}
	/*
		Calculation of shared memory limits are conducted using combined input from notes:
		- 1275776 - Linux: Preparing SLES for SAP environments:
		- 628131 - SAP MaxDB/liveCache operating system parameters on UNIX
		Regarding ShmMax:
		- "kernel.shmmax is in Bytes; minimum 20GB" $((VSZ*1024*1024*1024))
		- "shmmax >= 1073741824 bytes (= 1 GB)"
		Regarding ShmAll:
		- "kernel.shmall is in 4 KB pages; minimum 20GB" $((VSZ*1024*(1024/PSZ)))
		Regarding ShmMni:
		- "shmseg >= 1024 You can calculate the maximum number of shared memory segments required by SAP MaxDB as follows (for details, see the appendix (section 4)):
			SAP MaxDB/liveCache 7.3 - 7.9: shmseg >= TasksToApps + 50
			Limits the maximum number of shared memory segments per process"
		- "shmseg * 2 (but min. 1024) Defines the number of shared memory identifiers that are available in the system."
	*/
	conf, err := system.ParseSysconfigFile(path.Join(newPrepare.SysconfigPrefix, "/etc/sysconfig/saptune-note-1275776"), false)
	if err != nil {
		return nil, err
	}
	shmCountReferenceValue := conf.GetUint64("SHM_COUNT_REF_VALUE", 0)
	newPrepare.KernelShmMax = param.MaxU64(newPrepare.KernelShmMax, system.GetTotalMemSizeMB()*1049586 /* MB to Bytes */, 20*1024*1024*1024)
	newPrepare.KernelShmAll = param.MaxU64(newPrepare.KernelShmAll, system.GetTotalMemSizePages())
	newPrepare.KernelShmMni = param.MaxU64(newPrepare.KernelShmMni, shmCountReferenceValue, 2048)
	newPrepare.VMMaxMapCount = param.MaxU64(newPrepare.VMMaxMapCount, 2000000)

	/*
		Semaphore limits are set according to 1275776 - Linux: Preparing SLES for SAP environments:
		MSL: 1250
		MNS: 256000
		OPM: 100
		MNI: 8192
	*/
	newPrepare.KernelSemMsl = param.MaxU64(newPrepare.KernelSemMsl, 1250)
	newPrepare.KernelSemMns = param.MaxU64(newPrepare.KernelSemMns, 256000)
	newPrepare.KernelSemOpm = param.MaxU64(newPrepare.KernelSemOpm, 100)
	newPrepare.KernelSemMni = param.MaxU64(newPrepare.KernelSemMni, 8192)
	return newPrepare, nil
}
func (prepare PrepareForSAPEnvironments) Apply() error {
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
	secLimits.Set("@sapsys", "soft", "nofile", prepare.LimitNofileSapsysSoft)
	secLimits.Set("@sapsys", "hard", "nofile", prepare.LimitNofileSapsysSoft)
	secLimits.Set("@sdba", "soft", "nofile", prepare.LimitNofileSdbaSoft)
	secLimits.Set("@sdba", "hard", "nofile", prepare.LimitNofileSdbaHard)
	secLimits.Set("@dba", "soft", "nofile", prepare.LimitNofileDbaSoft)
	secLimits.Set("@dba", "hard", "nofile", prepare.LimitNofileDbaHard)
	if err := secLimits.Apply(); err != nil {
		return err
	}
	// Apply shared memory limits
	system.SetSysctlUint64(system.SYSCTL_SHMMAX, prepare.KernelShmMax)
	system.SetSysctlUint64(system.SYSCTL_SHMALL, prepare.KernelShmAll)
	system.SetSysctlUint64(system.SYSCTL_SHMMNI, prepare.KernelShmMni)
	system.SetSysctlUint64(system.SYSCTL_MAX_MAP_COUNT, prepare.VMMaxMapCount)
	// Apply semaphore limits
	system.SetSysctlString(system.SYSCTL_SEM, fmt.Sprintf("%d %d %d %d", prepare.KernelSemMsl, prepare.KernelSemMns, prepare.KernelSemOpm, prepare.KernelSemMni))
	return nil
}

// 1984787 - SUSE LINUX Enterprise Server 12: Installation notes
type AfterInstallation struct {
	UuiddSocket bool
}

func (inst AfterInstallation) Name() string {
	return "SUSE LINUX Enterprise Server 12: Installation notes"
}
func (inst AfterInstallation) Initialise() (Note, error) {
	return AfterInstallation{UuiddSocket: system.SystemctlIsRunning("uuidd.socket")}, nil
}
func (inst AfterInstallation) Optimise() (Note, error) {
	// Unconditionally enable uuid socket
	return AfterInstallation{UuiddSocket: true}, nil
}
func (inst AfterInstallation) Apply() error {
	if inst.UuiddSocket {
		return system.SystemctlEnableStart("uuidd.socket")
	} else {
		return system.SystemctlDisableStop("uuidd.socket")
	}
}
