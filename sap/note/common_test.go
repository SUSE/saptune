package note

import (
	"github.com/HouzuoGuo/saptune/system"
	"testing"
)

func TestPrepareForSAPEnvironments(t *testing.T) {
	prepare := PrepareForSAPEnvironments{SysconfigPrefix: SYSCONFIG_SRC_DIR}
	if prepare.Name() == "" {
		t.Fatal(prepare.Name())
	}
	initPrepare, err := prepare.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	optimised, err := initPrepare.(PrepareForSAPEnvironments).Optimise()
	if err != nil || optimised.(PrepareForSAPEnvironments) == initPrepare.(PrepareForSAPEnvironments) {
		t.Fatal(err, optimised, initPrepare)
	}
	// Check attributes from each optimised parameter
	o := optimised.(PrepareForSAPEnvironments)
	if o.ShmFileSystemSizeMB < system.GetTotalMemSizeMB()*75/100 {
		t.Fatalf("%+v", o)
	}
	if o.LimitNofileSapsysSoft < 32000 || o.LimitNofileSapsysHard < 32000 || o.LimitNofileSdbaSoft < 32000 || o.LimitNofileSdbaHard < 32000 ||
		o.LimitNofileDbaSoft < 32000 || o.LimitNofileDbaHard < 32000 {
		t.Fatalf("%+v", o)
	}
	if o.KernelShmMax < 1000 || o.KernelShmAll < 1000 || o.KernelShmMni < 2048 || o.VMMaxMapCount < 2000000 {
		t.Fatalf("%+v", o)
	}
	if o.KernelSemMsl < 1250 || o.KernelSemMns < 256000 || o.KernelSemOpm < 100 || o.KernelSemMni < 8192 {
		t.Fatalf("%+v", o)
	}
}

func TestAfterInstallation(t *testing.T) {
	if system.IsUserOBS() {
		t.Skip("uuid status cannot be inspected on build service")
	}
	inst := AfterInstallation{}
	if inst.Name() == "" {
		t.Fatal(inst.Name())
	}
	initInst, err := inst.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	optimised, err := initInst.(AfterInstallation).Optimise()
	if !optimised.(AfterInstallation).UuiddSocket {
		t.Fatal(optimised)
	}
}
