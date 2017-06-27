package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

const (
	ARCH_X86       = "amd64"   // GOARCH for 64-bit X86
	ARCH_PPC       = "ppc64le" // GOARCH for 64-bit PowerPC little endian
	LOGIND_DIR     = "/etc/systemd/logind.conf.d"
	SAP_LOGIN_FILE = "sap.conf"
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
	UserTasksMax bool
}

func (inst AfterInstallation) Name() string {
	return "SUSE LINUX Enterprise Server 12: Installation notes"
}
func (inst AfterInstallation) Initialise() (Note, error) {
	return AfterInstallation{UuiddSocket: system.SystemctlIsRunning("uuidd.socket"), UserTasksMax: CheckSapLogindFile()}, nil
}
func (inst AfterInstallation) Optimise() (Note, error) {
	if CheckSapLogindFile() {
		// print message fork bomb
		log.Print("ATTENTION    UserTasksMax set to infinity. With this setting your system is vulnerable to fork bomb attacks.")
	}
	return AfterInstallation{UuiddSocket: true, UserTasksMax: true}, nil
}
func (inst AfterInstallation) Apply() error {
	var err error
	if CheckSapLogindFile() {
		// print message fork bomb
		log.Print("ATTENTION    UserTasksMax set to infinity. With this setting your system is vulnerable to fork bomb attacks.")
	} else {
		// create directory /etc/systemd/logind.conf.d, if it does not exists
		if err = os.MkdirAll(LOGIND_DIR, 0755); err != nil {
			fmt.Printf("Error: Can't create directory '%s'\n", LOGIND_DIR)
			return err
		}
		// create file /etc/systemd/logind.conf.d/sap.conf
		err = ioutil.WriteFile(path.Join(LOGIND_DIR, SAP_LOGIN_FILE),[]byte("[Login]\nUserTasksMax=infinity\n"), 0644)
		if err != nil {
			fmt.Printf("Error: Can't create file '%s'\n", path.Join(LOGIND_DIR, SAP_LOGIN_FILE))
			return err
		}
		// print reboot
		log.Print("ATTENTION    UserTasksMax is now set to infinity. Please reboot the system for the changes to take effect.")
	}
	if IsVM() {
		//skip uuidd, does not work in VMs
		return err
	}
	if inst.UuiddSocket {
		err = system.SystemctlEnableStart("uuidd.socket")
	} else {
		err = system.SystemctlDisableStop("uuidd.socket")
	}
	return err
}
func CheckSapLogindFile() bool {
	_ , err := os.Stat(path.Join(LOGIND_DIR, SAP_LOGIN_FILE))
	if os.IsNotExist(err) {
		// file does not exists, create it later
		return false
	}
	if err == nil {
		// file does exists, check value of UserTasksMax
		content, err := ioutil.ReadFile(path.Join(LOGIND_DIR, SAP_LOGIN_FILE))
		if err != nil {
			fmt.Printf("Error: Can't read file '%s'. Continue anyway.\n", path.Join(LOGIND_DIR, SAP_LOGIN_FILE))
			return false
		}
		for _, line := range strings.Split(string(content), "\n") {
			matched, _ := regexp.MatchString("^[[:blank:]]*UserTasksMax[[:blank:]]*=[[:blank:]]*infinity", line)
			if matched {
				return true
			}
		}
		// value of UserTasksMax does not match our needs
		err = os.Rename(path.Join(LOGIND_DIR, SAP_LOGIN_FILE), path.Join(LOGIND_DIR, SAP_LOGIN_FILE + ".sav"))
		if err != nil {
			fmt.Printf("Error: Can't move file '%s' to '%s'. Continue anyway.\n", path.Join(LOGIND_DIR, SAP_LOGIN_FILE), path.Join(LOGIND_DIR, SAP_LOGIN_FILE + ".sav"))
		}
		return false
	}
	// another error concerning the file occured
	return false
}
func IsVM() bool {
// true - system is vm, false - system is phys.
	_ , err := os.Stat("/usr/bin/systemd-detect-virt")
	if err == nil {
		//systemd-detect-virt err=0 is VM, err=1 is phys.
		cmd := exec.Command("/usr/bin/systemd-detect-virt")
		_, err := cmd.Output()
		if err == nil {
			return true
		} else {
			return false
		}
	}
	out, err := exec.Command("/usr/sbin/dmidecode", "-s", "system-manufacturer").Output()
	switch strings.TrimSpace(string(out)) {
	case "QEMU", "Xen", "VirtualBox", "VMware, Inc.":
		return true
	}
	return false
}
