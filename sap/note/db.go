package note

import (
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"path"
)

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
	ret.KernelMMTransparentHugepage, _ = system.GetSysChoice(SysKernelTHPEnabled)
	ksmRun, _ := system.GetSysInt(SysKSMRun)
	ret.KernelMMKsm = ksmRun == 1
	nuBa, _ := system.GetSysctlInt(system.SysctlNumaBalancing)
	ret.KernelNumaBalancing = nuBa == 1
	return ret, nil
}
func (hana HANARecommendedOSSettings) Optimise() (Note, error) {
	ret := HANARecommendedOSSettings{
		KernelMMKsm:         false,
		KernelNumaBalancing: false,
	}
	ret.KernelMMTransparentHugepage = "never"
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
	err := system.WriteNoteErrors(errs)
	return err
}

// 1557506 - Linux paging improvements
type LinuxPagingImprovements struct {
	SysconfigPrefix string // Used by test cases to specify alternative sysconfig location

	VMPagecacheLimitMB          uint64
	VMPagecacheLimitIgnoreDirty int
	UseAlgorithmForHANA         bool
}

func (paging LinuxPagingImprovements) Name() string {
	return "Linux paging improvements"
}
func (paging LinuxPagingImprovements) Initialise() (Note, error) {
	vmPagecach, _ := system.GetSysctlUint64(system.SysctlPagecacheLimitMB)
	vmIgnoreDirty, _ := system.GetSysctlInt(system.SysctlPagecacheLimitIgnoreDirty)
	return LinuxPagingImprovements{
		SysconfigPrefix:             paging.SysconfigPrefix,
		VMPagecacheLimitMB:          vmPagecach,
		VMPagecacheLimitIgnoreDirty: vmIgnoreDirty,
	}, nil
}
func (paging LinuxPagingImprovements) Optimise() (Note, error) {
	newPaging := paging
	conf, err := txtparser.ParseSysconfigFile(path.Join(newPaging.SysconfigPrefix, "/etc/sysconfig/saptune-note-1557506"), false)
	if err != nil {
		return nil, err
	}
	inputEnable := conf.GetBool("ENABLE_PAGECACHE_LIMIT", false)
	inputOverride := conf.GetInt("OVERRIDE_PAGECACHE_LIMIT_MB", 0)
	inputIsHANA := conf.GetBool("TUNE_FOR_HANA", false)

	if inputIsHANA {
		// For HANA: new limit is 2% system memory
		newPaging.VMPagecacheLimitMB = system.GetMainMemSizeMB() * 2 / 100
	} else {
		// For NW: new limit is 1/16 of system memory, within range 512 to 4096
		newPaging.VMPagecacheLimitMB = system.GetMainMemSizeMB() / 16
		if newPaging.VMPagecacheLimitMB < 512 {
			newPaging.VMPagecacheLimitMB = 512
		} else if newPaging.VMPagecacheLimitMB > 4096 {
			newPaging.VMPagecacheLimitMB = 4096
		}
	}
	if inputOverride != 0 {
		newPaging.VMPagecacheLimitMB = uint64(inputOverride)
	}
	if !inputEnable {
		newPaging.VMPagecacheLimitMB = 0
	}
	newPaging.VMPagecacheLimitIgnoreDirty = conf.GetInt("PAGECACHE_LIMIT_IGNORE_DIRTY", 1)
	return newPaging, err
}
func (paging LinuxPagingImprovements) Apply() error {
	errs := make([]error, 0, 0)
	errs = append(errs, system.SetSysctlUint64(system.SysctlPagecacheLimitMB, paging.VMPagecacheLimitMB))
	errs = append(errs, system.SetSysctlInt(system.SysctlPagecacheLimitIgnoreDirty, paging.VMPagecacheLimitIgnoreDirty))

	err := system.WriteNoteErrors(errs)
	return err
}
