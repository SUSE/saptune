package system

import (
	"os"
	"path"
	"testing"
)

var readFileMatchText = `Only a test for read file
`

func TestIsUserRoot(t *testing.T) {
	if !IsUserRoot() {
		t.Log("the test requires root access")
	}
}

func TestGetOsName(t *testing.T) {
	actualVal := GetOsName()
	if actualVal != "SLES" && actualVal != "openSUSE Leap" {
		t.Errorf("OS is '%s' and not 'SLES'\n", actualVal)
	}
	if actualVal == "" {
		t.Logf("empty value returned for the os Name")
	}
}

func TestGetOsVers(t *testing.T) {
	actualVal := GetOsVers()
	switch actualVal {
	case "12", "12-SP1", "12-SP2", "12-SP3", "12-SP4", "15", "15-SP1":
		t.Logf("expected OS version '%s' found\n", actualVal)
	default:
		t.Logf("unexpected OS version '%s'\n", actualVal)
	}
}

func TestIsSLE15(t *testing.T) {
	if IsSLE15() {
		t.Logf("found SLE15 OS version\n")
	} else {
		t.Logf("OS version is '%s'\n", GetOsVers())
	}
}

func TestIsSLE12(t *testing.T) {
	if IsSLE12() {
		t.Logf("found SLE12 OS version\n")
	} else {
		t.Logf("OS version is '%s'\n", GetOsVers())
	}
}

func TestCmdIsAvailable(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/go") {
		t.Fatal("'/usr/bin/go' not found")
	}
	if CmdIsAvailable("/cmd_not_avail.CnA") {
		t.Fatal("found '/cmd_not_avail.CnA'")
	}
}

func TestCheckForPattern(t *testing.T) {
	if CheckForPattern("/file_not_available", "Test_Text") {
		t.Fatal("found '/file_not_available'")
	}
}

func TestGetServiceName(t *testing.T) {
	value := GetServiceName("sysstat")
	if value != "sysstat.service" {
		t.Fatalf("found service '%s' instead of 'sysstat.service'\n", value)
	}
	value = GetServiceName("sysstat.service")
	if value != "sysstat.service" {
		t.Fatalf("found service '%s' instead of 'sysstat.service'\n", value)
	}
	value = GetServiceName("UnkownService")
	if value != "" {
		t.Fatalf("found service '%s' instead of 'UnkownService'\n", value)
	}
}

func TestReadConfigFile(t *testing.T) {
	content, err := ReadConfigFile("/file_does_not_exist", true)
	if string(content) != "" {
		t.Fatal(content, err)
	}
	os.Remove("/file_does_not_exist")
	content, err = ReadConfigFile("/file_does_not_exist", false)
	if string(content) != "" || err == nil {
		t.Fatal(content, err)
	}
	//content, err = ReadConfigFile("/app/testdata/tstfile", false)
	content, err = ReadConfigFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile"), false)
	if string(content) != readFileMatchText || err != nil {
		t.Fatal(string(content), err)
	}
}

func TestCopyFile(t *testing.T) {
	//src := "/app/testdata/tstfile"
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	dst := "/tmp/saptune_tstfile"
	err := CopyFile(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ReadConfigFile(dst, false)
	if string(content) != readFileMatchText || err != nil {
		t.Fatal(string(content), err)
	}
	err = CopyFile("/file_does_not_exist", dst)
	if err == nil {
		t.Fatalf("copied from non existing file")
	}
	err = CopyFile(src, "/tmp/saptune_test/saptune_tstfile")
	if err == nil {
		t.Fatalf("copied from non existing file")
	}
}

func TestBlockDeviceIsDisk(t *testing.T) {
	devs := []string{"sda", "vda", "sr0"}
	for _, bdev := range devs {
		if !BlockDeviceIsDisk(bdev) {
			t.Logf("device unsupported, '%s' is NOT a disk", bdev)
		} else {
			t.Logf("device supported, '%s' is a disk", bdev)
		}
	}
}
