package system

import "testing"

func TestReadSys(t *testing.T) {
	if value, _ := GetSysString("kernel.vmcoreinfo"); len(value) < 3 {
		t.Fatal(value)
	}
	if value, _ := GetSysString("kernel/vmcoreinfo"); len(value) < 3 {
		t.Fatal(value)
	}
	if value, _ := GetSysString("kernel.not_avail"); value != "" {
		t.Fatal(value)
	}
	if value, _ := GetSysString("kernel/not_avail"); value != "" {
		t.Fatal(value)
	}
	GetSysInt("kernel/mm/ksm/run") // must not panic
	GetSysInt("kernel.mm.ksm.run") // must not panic
	if value, _ := GetSysInt("kernel/not_avail"); value != 0 {
		t.Fatal(value)
	}
	if choice, _ := GetSysChoice("kernel/mm/transparent_hugepage/enabled"); choice != "always" && choice != "madvise" && choice != "never" {
		t.Fatal(choice)
	}
	if choice, _ := GetSysChoice("kernel/not_avail"); choice != "" {
		t.Fatal(choice)
	}
}
