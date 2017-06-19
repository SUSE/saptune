package system

import "testing"

func TestReadSys(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
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
