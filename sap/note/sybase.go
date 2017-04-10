package note

import (
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"path"
	"runtime"
)

/*
1680803 - SYB: SAP Adaptive Server Enterprise - Best Practice for SAP Business Suite and SAP BW
Disable THP
Set HP
Set BlockDeviceSchedulers
Set BlockDeviceNrRequests
*/
const SYBASE_SYSCONFIG = "/etc/sysconfig/saptune-note-1680803"

type ASERecommendedOSSettings struct {
	SysconfigPrefix string

	KernelTransparentHugepage string
	KernelHugepages           uint64
	BlockDeviceSchedulers     param.BlockDeviceSchedulers
	BlockDeviceNrRequests     param.BlockDeviceNrRequests
}

func (ase ASERecommendedOSSettings) Name() string {
	return "SAP Adaptive Server Enterprise"
}

func (ase ASERecommendedOSSettings) Initialise() (Note, error) {
	actASE := ase
	if runtime.GOARCH == ARCH_X86 {
		actASE.KernelTransparentHugepage = system.GetSysChoice(system.SYS_THP)
		actASE.KernelHugepages = system.GetSysctlUint64(system.SYSCTL_NR_HUGEPAGES, 0)

	}
	newBlkSchedulers, err := actASE.BlockDeviceSchedulers.Inspect()
	if err != nil {
		return nil, err
	}
	actASE.BlockDeviceSchedulers = newBlkSchedulers.(param.BlockDeviceSchedulers)

	newBlkReq, err := actASE.BlockDeviceNrRequests.Inspect()
	if err != nil {
		return nil, err
	}
	actASE.BlockDeviceNrRequests = newBlkReq.(param.BlockDeviceNrRequests)
	return actASE, nil
}
func (ase ASERecommendedOSSettings) Optimise() (Note, error) {
	newASE := ase
	config, err := txtparser.ParseSysconfigFile(path.Join(newASE.SysconfigPrefix, SYBASE_SYSCONFIG), false)
	if err != nil {
		return nil, err
	}

	if runtime.GOARCH == ARCH_X86 {
		sval := config.GetString("SYBASE_THP", "")
		if sval == "yes" {
			newASE.KernelTransparentHugepage = "never"
		}

		ival := config.GetUint64("SYBASE_NUMBER_HUGEPAGES", 0)
		if ival > 0 {
			newASE.KernelHugepages = param.MaxU64(newASE.KernelHugepages, ival)
		} else {
			//TODO calculate
		}
	}

	sval := config.GetString("SYBASE_IO_SCHEDULER", "")
	if sval == "" {
		sval = "noop"
	}
	for blk := range newASE.BlockDeviceSchedulers.SchedulerChoice {
		newASE.BlockDeviceSchedulers.SchedulerChoice[blk] = sval
	}

	ival := config.GetInt("SYBASE_NRREQ", 0)
	if ival == 0 {
		ival = 1024
	}

	for blk := range newASE.BlockDeviceNrRequests.NrRequests {
		newASE.BlockDeviceNrRequests.NrRequests[blk] = ival
	}

	return newASE, nil
}
func (ase ASERecommendedOSSettings) Apply() error {
	if runtime.GOARCH == ARCH_X86 {
		system.SetSysString(system.SYS_THP, ase.KernelTransparentHugepage)
		system.SetSysctlUint64(system.SYSCTL_NR_HUGEPAGES, ase.KernelHugepages)
	}
	err := ase.BlockDeviceSchedulers.Apply()
	if err != nil {
		return err
	}
	err = ase.BlockDeviceNrRequests.Apply()
	return err
}
