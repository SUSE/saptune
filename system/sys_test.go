package system

import "testing"

func TestReadSys(t *testing.T) {
	if IsUserOBS() {
		t.Skip("Build service does not have /sys")
	}
	if value := GetSysString("kernel.vmcoreinfo"); len(value) < 3 {
		t.Fatal(value)
	}
	if value := GetSysString("kernel/vmcoreinfo"); len(value) < 3 {
		t.Fatal(value)
	}
	GetSysInt("kernel/mm/ksm/run") // must not panic
	GetSysInt("kernel.mm.ksm.run") // must not panic
	if choice := GetSysChoice("kernel/mm/transparent_hugepage/enabled"); choice != "always" && choice != "madvise" && choice != "never" {
		t.Fatal(choice)
	}
}
