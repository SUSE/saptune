package txtparser

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"testing"
)

func TestChkOsTags(t *testing.T) {
	tag := "15-*"
	secFields := []string{"rpm", "os=15-*", "arch=amd64"}

	ret := chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	_ = system.CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr12"), "/etc/os-release")
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	_ = system.CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr16.2"), "/etc/os-release")
	tag = "16.2"
	secFields = []string{"rpm", "os=16.2", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "19.0"
	secFields = []string{"rpm", "os=19.0", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[0,2,4-5,7-]"
	secFields = []string{"rpm", "os=16.[0,2,4-5,7-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[-2,4-5,7-]"
	secFields = []string{"rpm", "os=16.[-2,4-5,7-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[0-1,2-]"
	secFields = []string{"rpm", "os=16.[0-1,2-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[0-2,4-5,7-]"
	secFields = []string{"rpm", "os=16.[0-2,4-5,7-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[-6]"
	secFields = []string{"rpm", "os=16.[-6]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[02]"
	secFields = []string{"rpm", "os=16.[02]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "19.[2]"
	secFields = []string{"rpm", "os=19.[2]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[-]"
	secFields = []string{"rpm", "os=16.[-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[4-]"
	secFields = []string{"rpm", "os=16.[4-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[0+2]"
	secFields = []string{"rpm", "os=16.[0+2]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	_ = system.CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestChkHWTags(t *testing.T) {
	system.DmiID = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")
	secFields := []string{"sysctl", "vendor=SUSE", "arch=amd64"}
	info := "vendor"
	tag := "SUSE"
	ret := chkHWTags(info, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content of %s file", tag, info)
	}

	secFields = []string{"sysctl", "model=SUSE saptune", "arch=amd64"}
	info = "model"
	tag = "SUSE saptune"
	ret = chkHWTags(info, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content of %s file", tag, info)
	}

	tag = "SE sap"
	ret = chkHWTags(info, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content of %s file", tag, info)
	}

	tag = "hugo"
	ret = chkHWTags(info, tag, secFields)
	if ret {
		t.Errorf("tag '%s' matches content of %s file, but shouldn't", tag, info)
	}

	os.Rename(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name"), path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name_OrG"))
	tag = "SUSE saptune"
	ret = chkHWTags(info, tag, secFields)
	if ret {
		t.Errorf("tag '%s' matches content of %s file, but shouldn't", tag, info)
	}
	os.Rename(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name_OrG"), path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name"))

	system.DmiID = "/sys/class/dmi/id"
}

func TestChkOtherTags(t *testing.T) {
	system.DmiID = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")
	secFields := []string{"sysctl", "product_name=SUSE", "arch=amd64"}
	file := "product_name"
	tag := "SUSE"

	ret := chkOtherTags(file, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content in '%+s'", tag, file)
	}

	tag = "saptune"
	ret = chkOtherTags(file, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content in '%+s'", tag, file)
	}

	tag = "SE sap"
	ret = chkOtherTags(file, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content in '%+s'", tag, file)
	}

	tag = "hugo"
	ret = chkOtherTags(file, tag, secFields)
	if ret {
		t.Errorf("tag '%s' does match content in '%+s', but shouldn't", tag, file)
	}
	system.DmiID = "/sys/class/dmi/id"
}

func TestIsTagAvail(t *testing.T) {
	secFields := []string{"block", "blkvendor=HUGO", "blkmodel=EGON"}
	tag := "blkvendor"
	if !isTagAvail(tag, secFields) {
		t.Errorf("tag '%s' is expected to be available, but is repoted as not available\n", tag)
	}
	tag = "blkmodel"
	if !isTagAvail(tag, secFields) {
		t.Errorf("tag '%s' is expected to be available, but is repoted as not available\n", tag)
	}
	tag = "blkpat"
	if isTagAvail(tag, secFields) {
		t.Errorf("tag '%s' is expected to be NOT available, but is repoted as available\n", tag)
	}
	tag = "blkvendor"
	secFields = []string{"block", "blkvendor="}
	if !isTagAvail(tag, secFields) {
		t.Error("expected 'true', because of correct syntax, but got 'false'")
	}
	secFields = []string{"block", "blkvendor"}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	secFields = []string{"block", "=EGON"}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	secFields = []string{"block", "="}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	secFields = []string{"block", ""}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	// meminfo tag in isTagAvail
	secFields = []string{"sysctl", "meminfo%MemTotal>=2147483648"}
	if !isTagAvail("meminfo", secFields) {
		t.Error("expected 'true' for isTagAvail('meminfo'), but got 'false'")
	}
	secFields = []string{"block", "meminfo%MemTotal>=2147483648", "blkvendor=HUGO"}
	if !isTagAvail("blkvendor", secFields) {
		t.Error("expected 'true' for isTagAvail('blkvendor') following meminfo tag, but got 'false'")
	}
}

func TestChkMeminfoTags(t *testing.T) {
	origProc := system.ProcMeminfo
	defer func() { system.ProcMeminfo = origProc }()

	tmpFile, err := os.CreateTemp("", "meminfo_test_tag")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `MemTotal:       2147483648 kB
MemFree:         1048576 kB
MemAvailable:   1500000 kB
Active(anon):      50000 kB
HugePages_Total:       4
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	system.ProcMeminfo = tmpFile.Name()

	secFields := []string{"sysctl"}

	// Valid matching cases for all operators
	validMatches := []string{
		"meminfo%MemTotal==2147483648",
		"meminfo%MemTotal!=1000",
		"meminfo%MemTotal>=2147483648",
		"meminfo%MemTotal>=2000000000",
		"meminfo%MemTotal>2000000000",
		"meminfo%MemTotal<=2147483648",
		"meminfo%MemTotal<=3000000000",
		"meminfo%MemTotal<3000000000",
		"meminfo%MemFree<=1048576",
		"meminfo%MemFree<2000000",
		"meminfo%Active(anon)==50000",
		"meminfo%HugePages_Total==4",
		"meminfo%HugePages_Total>0",
		"meminfo%MemTotal >= 2147483648",
		"meminfo%MemTotal == 2147483648",
		"meminfo%MemFree < 2000000",
	}
	for _, tag := range validMatches {
		if !chkMeminfoTags(tag, append(secFields, tag)) {
			t.Errorf("expected tag '%s' to match, but got false", tag)
		}
	}

	// Valid non-matching cases
	nonMatches := []string{
		"meminfo%MemTotal==123",
		"meminfo%MemTotal!=2147483648",
		"meminfo%MemTotal>2147483648",
		"meminfo%MemTotal<2147483648",
		"meminfo%MemFree>=2000000",
		"meminfo%MemFree<=500000",
		"meminfo%HugePages_Total==0",
		"meminfo%HugePages_Total<2",
		"meminfo%MemTotal <= 100",
	}
	for _, tag := range nonMatches {
		if chkMeminfoTags(tag, append(secFields, tag)) {
			t.Errorf("expected tag '%s' to NOT match, but got true", tag)
		}
	}

	// Unknown key
	if chkMeminfoTags("meminfo%NoSuchKey>=100", append(secFields, "meminfo%NoSuchKey>=100")) {
		t.Errorf("expected tag with non-existent key to return false")
	}

	// Invalid syntax cases
	invalidSyntax := []string{
		"meminfo%MemTotal",
		"meminfo%MemTotal=2147483648",
		"meminfo%MemTotal=~2147483648",
		"meminfo%MemTotal>=abc",
		"meminfo%>=100",
		"meminfo%",
		"meminfo%MemTotal>>100",
	}
	for _, tag := range invalidSyntax {
		if chkMeminfoTags(tag, append(secFields, tag)) {
			t.Errorf("expected invalid tag '%s' to return false, but got true", tag)
		}
	}
}

func TestChkSecTagsWithMeminfo(t *testing.T) {
	origProc := system.ProcMeminfo
	defer func() { system.ProcMeminfo = origProc }()

	tmpFile, err := os.CreateTemp("", "meminfo_test_sec")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `MemTotal:       2147483648 kB
MemFree:         1048576 kB
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	system.ProcMeminfo = tmpFile.Name()

	// Section matching meminfo tag
	ok, _ := chkSecTags([]string{"sysctl", "meminfo%MemTotal>=2147483648"}, nil)
	if !ok {
		t.Errorf("expected chkSecTags with matching meminfo tag to be true")
	}

	// Section failing meminfo tag
	ok, _ = chkSecTags([]string{"sysctl", "meminfo%MemTotal<1000"}, nil)
	if ok {
		t.Errorf("expected chkSecTags with non-matching meminfo tag to be false")
	}

	// Section with multiple meminfo tags: one true, one false
	ok, _ = chkSecTags([]string{"sysctl", "meminfo%MemTotal>=2147483648", "meminfo%MemFree<500"}, nil)
	if ok {
		t.Errorf("expected chkSecTags with one failing tag to be false")
	}

	// Section with multiple meminfo tags: both true
	ok, _ = chkSecTags([]string{"sysctl", "meminfo%MemTotal>=2147483648", "meminfo%MemFree>500"}, nil)
	if !ok {
		t.Errorf("expected chkSecTags with both matching tags to be true")
	}
}

func TestParseINIWithMeminfo(t *testing.T) {
	origProc := system.ProcMeminfo
	defer func() { system.ProcMeminfo = origProc }()

	tmpMem, err := os.CreateTemp("", "meminfo_ini_test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpMem.Name())

	contentMem := `MemTotal:       2147483648 kB
MemFree:         1048576 kB
`
	if _, err := tmpMem.WriteString(contentMem); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpMem.Close()

	system.ProcMeminfo = tmpMem.Name()

	iniContent := `
[sysctl:meminfo%MemTotal>=2147483648]
kernel.shmmax = 123456789

[sysctl:meminfo%MemTotal<1000000]
kernel.shmall = 999999
`
	iniData := ParseINI(iniContent)

	sysctlMap, exists := iniData.KeyValue["sysctl"]
	if !exists {
		t.Fatalf("expected 'sysctl' section to exist in parsed INI")
	}

	if entry, ok := sysctlMap["kernel.shmmax"]; !ok || entry.Value != "123456789" {
		t.Errorf("expected kernel.shmmax to be '123456789', got '%+v'", entry)
	}

	if _, ok := sysctlMap["kernel.shmall"]; ok {
		t.Errorf("kernel.shmall should have been skipped due to non-matching meminfo tag")
	}
}
