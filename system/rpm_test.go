package system

import (
	"testing"
)

var vers1 = "228-150.22.1"
var vers2 = "228-142.1"
var vers3 = "229-150.22.1"
var vers4 = "228-160.22.1"
var vers5 = "228-150.25.1"
var vers6 = "228-150.19.1"
var vers7 = "228-150.22.4"
var vers8 = "228-150.22.0"

func TestGetRpmVers(t *testing.T) {
	actualVal := GetRpmVers("kernel-default")
	if actualVal == "" {
		t.Log("rpm 'kernel-default' not found")
	}
}

func TestCmpRpmVers(t *testing.T) {
	actualVal := CmpRpmVers(vers1, vers2)
	if !actualVal {
		t.Fatalf("'%s' reported as < '%s'\n", vers1, vers2)
	}
	actualVal = CmpRpmVers(vers2, vers1)
	if actualVal {
		t.Fatalf("'%s' reported as >= '%s'\n", vers2, vers1)
	}
	actualVal = CmpRpmVers(vers1, vers3)
	if actualVal {
		t.Fatalf("'%s' reported as >= '%s'\n", vers1, vers3)
	}
	actualVal = CmpRpmVers(vers1, vers4)
	if actualVal {
		t.Fatalf("'%s' reported as >= '%s'\n", vers1, vers4)
	}
	actualVal = CmpRpmVers(vers1, vers5)
	if actualVal {
		t.Fatalf("'%s' reported as >= '%s'\n", vers1, vers5)
	}
	actualVal = CmpRpmVers(vers1, vers6)
	if !actualVal {
		t.Fatalf("'%s' reported as < '%s'\n", vers1, vers6)
	}
	actualVal = CmpRpmVers(vers1, vers7)
	if actualVal {
		t.Fatalf("'%s' reported as >= '%s'\n", vers1, vers7)
	}
	actualVal = CmpRpmVers(vers1, vers8)
	if !actualVal {
		t.Fatalf("'%s' reported as < '%s'\n", vers1, vers8)
	}
}

func TestCheckRpmVers(t *testing.T) {
	actualVal := CheckRpmVers("228", "228")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("150.22.1", "142.1")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("3.0.2a", "3.0.2a")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("3.0.2a", "3.0.2")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("3.0.2", "3.0.2a")
	if actualVal >= 0 {
		t.Fatal("higher")
	}
	actualVal = CheckRpmVers("5.5p10", "5.5p10")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("5.5p10", "5.5p1")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("5.5p1", "5.5p10")
	if actualVal >= 0 {
		t.Fatal("higher")
	}
	actualVal = CheckRpmVers("1b.fc17", "1b.fc17")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("1.fc17", "1b.fc17")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("1b.fc17", "1.fc17")
	if actualVal >= 0 {
		t.Fatal("higher")
	}
	actualVal = CheckRpmVers("1g.fc17", "1g.fc17")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("1g.fc17", "1.fc17")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("1.fc17", "1g.fc17")
	if actualVal >= 0 {
		t.Fatal("higher")
	}
	actualVal = CheckRpmVers("20101121", "20101121")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("20101122", "20101121")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("20101121", "20101122")
	if actualVal >= 0 {
		t.Fatal("higher")
	}
	actualVal = CheckRpmVers("6.0.rc1", "6.0.rc1")
	if actualVal != 0 {
		t.Fatal("unequal")
	}
	actualVal = CheckRpmVers("6.0.rc1", "6.0")
	if actualVal != 1 {
		t.Fatal("less or equal")
	}
	actualVal = CheckRpmVers("6.0", "6.0.rc1")
	if actualVal >= 0 {
		t.Fatal("higher")
	}
}
