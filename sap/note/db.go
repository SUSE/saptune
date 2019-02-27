package note

import (
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"path"
)

// 1557506 - Linux paging improvements
type LinuxPagingImprovements struct {
	SysconfigPrefix string // Used by test cases to specify alternative sysconfig location

	VMPagecacheLimitMB          uint64
	VMPagecacheLimitIgnoreDirty int
	UseAlgorithmForHANA         bool
}

// Name returns the name of the related note
func (paging LinuxPagingImprovements) Name() string {
	return "Linux paging improvements"
}

// Initialise reads the parameter values from current system
func (paging LinuxPagingImprovements) Initialise() (Note, error) {
	vmPagecach, _ := system.GetSysctlUint64(system.SysctlPagecacheLimitMB)
	vmIgnoreDirty, _ := system.GetSysctlInt(system.SysctlPagecacheLimitIgnoreDirty)
	return LinuxPagingImprovements{
		SysconfigPrefix:             paging.SysconfigPrefix,
		VMPagecacheLimitMB:          vmPagecach,
		VMPagecacheLimitIgnoreDirty: vmIgnoreDirty,
		UseAlgorithmForHANA:         true,
	}, nil
}

// Optimise gets the expected pagecache values from the configuration
// or calculates new values
func (paging LinuxPagingImprovements) Optimise() (Note, error) {
	newPaging := paging
	conf, err := txtparser.ParseSysconfigFile(path.Join(newPaging.SysconfigPrefix, "/usr/share/saptune/notes/1557506"), false)
	if err != nil {
		return nil, err
	}
	inputEnable := conf.GetBool("ENABLE_PAGECACHE_LIMIT", false)
	inputOverride := conf.GetInt("OVERRIDE_PAGECACHE_LIMIT_MB", 0)

	// As discussed with SAP and Alliance team, use the HANA formula for
	// Netweaver too.
	// So for HANA and Netweaver: new limit is 2% system memory
	newPaging.VMPagecacheLimitMB = system.GetMainMemSizeMB() * 2 / 100
	if inputOverride != 0 {
		newPaging.VMPagecacheLimitMB = uint64(inputOverride)
	}
	if !inputEnable {
		newPaging.VMPagecacheLimitMB = 0
	}
	newPaging.VMPagecacheLimitIgnoreDirty = conf.GetInt("PAGECACHE_LIMIT_IGNORE_DIRTY", 1)
	return newPaging, err
}

// Apply sets the new values in the system
func (paging LinuxPagingImprovements) Apply() error {
	errs := make([]error, 0, 0)
	errs = append(errs, system.SetSysctlUint64(system.SysctlPagecacheLimitMB, paging.VMPagecacheLimitMB))
	errs = append(errs, system.SetSysctlInt(system.SysctlPagecacheLimitIgnoreDirty, paging.VMPagecacheLimitIgnoreDirty))

	err := sap.PrintErrors(errs)
	return err
}
