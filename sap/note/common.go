package note

import (
	"fmt"
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
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
	var err error

	newPrepare.KernelSemMsl, newPrepare.KernelSemMns, newPrepare.KernelSemOpm, newPrepare.KernelSemMni = system.GetSemaphoreLimits()
	return newPrepare, err
}
func (prepare PrepareForSAPEnvironments) Optimise() (Note, error) {
	newPrepare := prepare

	for _, val := range []*system.SecurityLimitInt{&newPrepare.LimitNofileSapsysSoft, &newPrepare.LimitNofileSapsysHard, &newPrepare.LimitNofileSdbaSoft, &newPrepare.LimitNofileSdbaHard, &newPrepare.LimitNofileDbaSoft, &newPrepare.LimitNofileDbaHard} {
		if *val < 32800 {
			*val = 32800
		}
	}
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
	// Apply semaphore limits
	errs = append(errs, system.SetSysctlString(system.SysctlSem, fmt.Sprintf("%d %d %d %d", prepare.KernelSemMsl, prepare.KernelSemMns, prepare.KernelSemOpm, prepare.KernelSemMni)))

	err = sap.PrintErrors(errs)
	return err
}

