package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/sap"
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const (
	// LoginConfDir is the path to systemd's logind configuration directory under /etc.
	LogindConfDir = "/etc/systemd/logind.conf.d"
	// LogindSAPConfFile is a configuration file full of SAP-specific settings for logind.
	LogindSAPConfFile = "sap.conf"
	// LogindSAAPConfContent is the verbatim content of SAP-specific logind settings file.
	LogindSAPConfContent = `
[Login]
UserTasksMax=infinity
`
)

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
	newPrepare.LimitNofileSapsysSoft = secLimits.GetOr0("@sapsys", "soft", "nofile")
	newPrepare.LimitNofileSapsysHard = secLimits.GetOr0("@sapsys", "hard", "nofile")
	newPrepare.LimitNofileSdbaSoft = secLimits.GetOr0("@sdba", "soft", "nofile")
	newPrepare.LimitNofileSdbaHard = secLimits.GetOr0("@sdba", "hard", "nofile")
	newPrepare.LimitNofileDbaSoft = secLimits.GetOr0("@dba", "soft", "nofile")
	newPrepare.LimitNofileDbaHard = secLimits.GetOr0("@dba", "hard", "nofile")
	// Find out shared memory limits
	newPrepare.KernelShmMax, _ = system.GetSysctlUint64(system.SysctlShmax)
	newPrepare.KernelShmAll, _ = system.GetSysctlUint64(system.SysctlShmall)
	newPrepare.KernelShmMni, _ = system.GetSysctlUint64(system.SysctlShmni)
	newPrepare.VMMaxMapCount, _ = system.GetSysctlUint64(system.SysctlMaxMapCount)
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
	for _, val := range []*system.SecurityLimitInt{&newPrepare.LimitNofileSapsysSoft, &newPrepare.LimitNofileSapsysHard, &newPrepare.LimitNofileSdbaSoft, &newPrepare.LimitNofileSdbaHard, &newPrepare.LimitNofileDbaSoft, &newPrepare.LimitNofileDbaHard} {
		switch *val {
		case system.SecurityLimitUnlimitedValue:
			// nothing to do, value remain untouched
		default:
			if *val < 32800 {
				*val = 32800
			}
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
	conf, err := txtparser.ParseSysconfigFile(path.Join(newPrepare.SysconfigPrefix, "/etc/sysconfig/saptune-note-1275776"), false)
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
	logindContent, err := ioutil.ReadFile(path.Join(LogindConfDir, LogindSAPConfFile))
	if err != nil && !os.IsNotExist(err) {
		return AfterInstallation{}, err
	}
	return AfterInstallation{
		UuiddSocketStatus: system.SystemctlIsRunning("uuidd.socket"),
		LogindConfigured:  string(logindContent) == LogindSAPConfContent,
	}, nil
}
func (inst AfterInstallation) Optimise() (Note, error) {
	return AfterInstallation{UuiddSocketStatus: true, LogindConfigured: true}, nil
}
func (inst AfterInstallation) Apply() error {
	// Set UUID socket status
	var err error
	if inst.UuiddSocketStatus {
		err = system.SystemctlEnableStart("uuidd.socket")
	} else {
		err = system.SystemctlDisableStop("uuidd.socket")
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
