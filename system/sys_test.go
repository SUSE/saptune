package system

import (
	"testing"
)

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
	if choice, _ := GetSysChoice("kernel/mm/ksm/run"); choice != "" {
		t.Error(choice)
	}
}

func TestWriteSys(t *testing.T) {
	value := ""
	key := "kernel/mm/transparent_hugepage/enabled"
	oldVal, _ := GetSysChoice(key)
	if oldVal == "never" {
		value = "always"
	} else {
		value = "never"
	}
	if err := SetSysString(key, value); err != nil {
		t.Fatal(err)
	}
	choice, _ := GetSysChoice(key)
	if choice != value {
		t.Fatal(choice)
	}
	// set test value back
	if err := SetSysString(key, oldVal); err != nil {
		t.Fatal(err)
	}
	ival := 0
	key = "kernel/mm/ksm/run"
	oval, _ := GetSysInt(key)
	if oval == 0 {
		ival = 1
	}
	if err := SetSysInt(key, ival); err != nil {
		t.Fatal(err)
	}
	nval, _ := GetSysInt(key)
	if nval != ival {
		t.Fatal(nval)
	}
	// set test value back
	if err := SetSysInt(key, oval); err != nil {
		t.Fatal(err)
	}
	if err := SetSysString("kernel/not_avail", "1"); err == nil {
		t.Fatal("writing to an non existent sys key")
	}
	if err := SetSysInt("kernel/not_avail", 1); err == nil {
		t.Fatal("writing to an non existent sys key")
	}
}

func TestTestSysString(t *testing.T) {
	if tstErr := TestSysString("kernel/mm/ksm/run", "0"); tstErr == nil {
		t.Log("writing sys key is possible")
	} else {
		t.Log("could not write sys key")
	}
	if tstErr := TestSysString("kernel/not_avail", "0"); tstErr == nil {
		t.Fatal("writing to an non existent sys key")
	}
}

func TestGetSysSearchParam(t *testing.T) {
	skey := "sys:kernel.mm.transparent_hugepage.enabled"
	mtch := "THP"
	msect := "vm"
	searchParam, sect := GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "THP"
	mtch = "sys:kernel.mm.transparent_hugepage.enabled"
	msect = "sys"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "sys:kernel.mm.ksm.run"
	mtch = "KSM"
	msect = "vm"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "KSM"
	mtch = "sys:kernel.mm.ksm.run"
	msect = "sys"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "IO_SCHEDULER_sdc"
	mtch = "sys:block.sdc.queue.scheduler"
	msect = "sys"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "sys:block.sdc.queue.scheduler"
	mtch = "IO_SCHEDULER_sdc"
	msect = "block"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "NRREQ_sdb"
	mtch = "sys:block.sdb.queue.nr_requests"
	msect = "sys"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "sys:block.sdb.queue.nr_requests"
	mtch = "NRREQ_sdb"
	msect = "block"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "READ_AHEAD_KB_sdd"
	mtch = "sys:block.sdd.queue.read_ahead_kb"
	msect = "sys"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "sys:block.sdd.queue.read_ahead_kb"
	mtch = "READ_AHEAD_KB_sdd"
	msect = "block"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "MAX_SECTORS_KB_sde"
	mtch = "sys:block.sde.queue.max_sectors_kb"
	msect = "sys"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}

	skey = "sys:block.sde.queue.max_sectors_kb"
	mtch = "MAX_SECTORS_KB_sde"
	msect = "block"
	searchParam, sect = GetSysSearchParam(skey)
	if searchParam != mtch {
		t.Errorf("expected '%s', got '%s'\n", mtch, searchParam)
	}
	if sect != msect {
		t.Errorf("expected '%s', got '%s'\n", msect, sect)
	}
}
