package system

import (
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
)

var readFileMatchText = `Only a test for read file
`

func TestIsUserRoot(t *testing.T) {
	if !IsUserRoot() {
		t.Log("the test requires root access")
	}
}

func TestCliArg(t *testing.T) {
	os.Args = []string{"saptune", "note", "list"}

	expected := "note"
	actual := CliArg(1)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expected, actual)
	}
	expected = "list"
	actual = CliArg(2)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expected, actual)
	}
	expected = ""
	actual = CliArg(4)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expected, actual)
	}
}

func TestGetSolutionSelector(t *testing.T) {
	solSelector := GetSolutionSelector()
	t.Logf("architecture is '%s'\n", solSelector)
	//if solSelector != "amd64" && solSelector != "amd64_PC" && solSelector != "ppc64le" && solSelector != "ppc64le_PC" && solSelector != "TRAVIS_TODO" {
	if solSelector != "amd64" && solSelector != "amd64_PC" && solSelector != "ppc64le" && solSelector != "ppc64le_PC" {
		t.Errorf("Test failed, solSelector '%s'", solSelector)
	}
}

func TestGetOsName(t *testing.T) {
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
	actualVal := GetOsName()
	//if actualVal != "SLES" && actualVal != "openSUSE Leap" {
	if actualVal != "SLES" {
		t.Logf("OS is '%s' and not 'SLES'\n", actualVal)
		_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr15"), "/etc/os-release")
		actualVal = GetOsName()
		if actualVal != "SLES" {
			t.Errorf("OS is '%s' and not 'SLES'\n", actualVal)
		}
	}
	// test with non existing file
	os.Remove("/etc/os-release")
	actualVal = GetOsName()
	if actualVal != "" {
		t.Errorf("/etc/os-release should not exist, but returns value '%v'\n", actualVal)
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestGetOsVers(t *testing.T) {
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
	actualVal := GetOsVers()
	if actualVal == "" {
		_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr15"), "/etc/os-release")
		actualVal = GetOsVers()
		if actualVal != "15-SP2" {
			t.Errorf("unexpected OS version '%s'\n", actualVal)
		}
		_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr12"), "/etc/os-release")
		actualVal = GetOsVers()
		if actualVal != "12-SP5" {
			t.Errorf("unexpected OS version '%s'\n", actualVal)
		}
	} else {
		switch actualVal {
		case "12", "12-SP1", "12-SP2", "12-SP3", "12-SP4", "12-SP5", "15", "15-SP1", "15-SP2":
			t.Logf("expected OS version '%s' found\n", actualVal)
		default:
			t.Logf("unexpected OS version '%s'\n", actualVal)
		}
	}

	// test with non existing file
	os.Remove("/etc/os-release")
	actualVal = GetOsVers()
	if actualVal != "" {
		t.Errorf("/etc/os-release should not exist, but returns value '%v'\n", actualVal)
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestIsSLE15(t *testing.T) {
	_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr15"), "/etc/os-release")
	if IsSLE15() {
		t.Logf("found SLE15 OS version\n")
		_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
		if IsSLE15() {
			t.Errorf("expected a non SLE15 os version, but OS version is '%s'\n", GetOsVers())
		}
	} else {
		t.Errorf("expected '15-SP2', but OS version is '%s'\n", GetOsVers())
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestIsSLE12(t *testing.T) {
	_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr12"), "/etc/os-release")
	if IsSLE12() {
		t.Logf("found SLE12 OS version\n")
		_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
		if IsSLE12() {
			t.Errorf("expected a non SLE12 os version, but OS version is '%s'\n", GetOsVers())
		}
	} else {
		t.Errorf("expected '12-SP5', but OS version is '%s'\n", GetOsVers())
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestCmdIsAvailable(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/go") {
		t.Error("'/usr/bin/go' not found")
	}
	if CmdIsAvailable("/cmd_not_avail.CnA") {
		t.Error("found '/cmd_not_avail.CnA'")
	}
}

func TestCheckForPattern(t *testing.T) {
	if CheckForPattern("/file_not_available", "Test_Text") {
		t.Error("found '/file_not_available'")
	}
	if CheckForPattern("/sys/module/video/uevent", "Test_Text") {
		t.Error("could read '/sys/module/video/uevent'")
	}
}

func TestGetServiceName(t *testing.T) {
	value := GetServiceName("sysstat")
	if value != "sysstat.service" {
		t.Errorf("found service '%s' instead of 'sysstat.service'\n", value)
	}
	value = GetServiceName("sysstat.service")
	if value != "sysstat.service" {
		t.Errorf("found service '%s' instead of 'sysstat.service'\n", value)
	}
	value = GetServiceName("UnkownService")
	if value != "" {
		t.Errorf("found service '%s' instead of 'UnkownService'\n", value)
	}
	// test with missing command
	cmdName := "/usr/bin/systemctl"
	savName := "/usr/bin/systemctl_SAVE"
	if err := os.Rename(cmdName, savName); err != nil {
		t.Error(err)
	}
	value = GetServiceName("sysstat")
	if value != "" {
		t.Errorf("found service '%s' instead of 'UnkownService'\n", value)
	}
	if err := os.Rename(savName, cmdName); err != nil {
		t.Error(err)
	}
}

func TestReadConfigFile(t *testing.T) {
	content, err := ReadConfigFile("/file_does_not_exist", true)
	if string(content) != "" {
		t.Error(content, err)
	}
	os.Remove("/file_does_not_exist")
	content, err = ReadConfigFile("/file_does_not_exist", false)
	if string(content) != "" || err == nil {
		t.Error(content, err)
	}
	//content, err = ReadConfigFile("/app/testdata/tstfile", false)
	content, err = ReadConfigFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile"), false)
	if string(content) != readFileMatchText || err != nil {
		t.Error(string(content), err)
	}
}

func TestCopyFile(t *testing.T) {
	//src := "/app/testdata/tstfile"
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	dst := "/tmp/saptune_tstfile"
	err := CopyFile(src, dst)
	if err != nil {
		t.Error(err)
	}
	content, err := ReadConfigFile(dst, false)
	if string(content) != readFileMatchText || err != nil {
		t.Error(string(content), err)
	}
	err = CopyFile("/file_does_not_exist", dst)
	if err == nil {
		t.Errorf("copied from non existing file")
	}
	err = CopyFile(src, "/tmp/saptune_test/saptune_tstfile")
	if err == nil {
		t.Errorf("copied from non existing file")
	}
}

func TestBlockDeviceIsDisk(t *testing.T) {
	if !BlockDeviceIsDisk("sda") {
		t.Error("'sda' is wrongly reported as 'NOT a disk'")
	}
	if !BlockDeviceIsDisk("vda") {
		t.Error("'vda' is wrongly reported as 'NOT a disk'")
	}
	if BlockDeviceIsDisk("sr0") {
		t.Error("'sr0' is wrongly reported as 'a disk'")
	}
	//devs := []string{"sda", "vda", "sr0"}
}

func TestCollectBlockDeviceInfo(t *testing.T) {
	_, sysDevs := ListDir("/sys/block", "the available block devices of the system")
	if len(sysDevs) == 0 {
		t.Skipf("skip test, seems there are no block devices avaialble, sysDevs is '%+v'\n", sysDevs)
	}
	collect := CollectBlockDeviceInfo()
	if len(collect) == 0 {
		t.Errorf("seems no block devices collected, collect is %+v'\n", collect)
	}
	for _, dev := range sysDevs {
		found := false
		for _, entry := range collect {
			if dev == entry {
				found = true
				break
			}
		}
		if !found {
			if BlockDeviceIsDisk(dev) {
				t.Errorf("'%s' is a disk, but not returned by 'CollectBlockDeviceInfo'\n", dev)
			}
		}
	}
	bdevFile := path.Join(SaptuneSectionDir, "/blockdev.run")
	if _, err := os.Stat(bdevFile); os.IsNotExist(err) {
		t.Errorf("file '%+s' missing\n", bdevFile)
	}
	_ = os.Remove(bdevFile)
}

func TestGetBlockDeviceInfo(t *testing.T) {
	bdevConf := BlockDev{
		AllBlockDevs:    make([]string, 0, 6),
		BlockAttributes: make(map[string]map[string]string),
	}
	bdevConf.AllBlockDevs = []string{"sda", "sdb", "sdc"}
	bdevConf.BlockAttributes["sda"] = map[string]string{"IO_SCHEDULER": "mq-deadline", "NRREQ": "32", "READ_AHEAD_KB": "512"}
	bdevConf.BlockAttributes["sdb"] = map[string]string{"IO_SCHEDULER": "bfq", "NRREQ": "64", "READ_AHEAD_KB": "1024"}
	bdevConf.BlockAttributes["sdc"] = map[string]string{"IO_SCHEDULER": "none", "NRREQ": "4", "READ_AHEAD_KB": "128"}

	err := storeBlockDeviceInfo(bdevConf)
	if err != nil {
		t.Error("storing block device info failed")
	}

	blkDev, _ := GetBlockDeviceInfo()
	eq := reflect.DeepEqual(bdevConf.AllBlockDevs, blkDev.AllBlockDevs)
	if !eq {
		t.Errorf("stored and read block devices differ- stored:'%+v' - read:'%+v'\n", bdevConf.AllBlockDevs, blkDev.AllBlockDevs)
	}
	for _, entry := range bdevConf.AllBlockDevs {
		eq := reflect.DeepEqual(bdevConf.BlockAttributes[entry], blkDev.BlockAttributes[entry])
		if !eq {
			t.Errorf("stored and read entries differ - stored:'%+v' - read:'%+v'\n", bdevConf.BlockAttributes[entry], blkDev.BlockAttributes[entry])
		}
	}
	bdevFile := path.Join(SaptuneSectionDir, "/blockdev.run")
	_ = os.Remove(bdevFile)
}

func TestCalledFrom(t *testing.T) {
	val := CalledFrom()
	if !strings.Contains(val, "testing.go") {
		t.Fatalf("called from '%s' instead of 'testing.go'\n", val)
	}
}

/*

## fieldType is 'string'
SaptuneVersion = "5"
system.ErrorExit("Wrong saptune version in file '/etc/sysconfig/saptune': %s", SaptuneVersion)
-> ERROR: Wrong saptune version in file '/etc/sysconfig/saptune': 5
-> $? ist 1

## fieldType is 'int'
system.ErrorExit("Wrong saptune version in file '/etc/sysconfig/saptune'.", 4)
-> ERROR: Wrong saptune version in file '/etc/sysconfig/saptune'.
-> $? ist 4

const (
	exitNotYetTuned = 5
)
system.ErrorExit("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.", exitNotYetTuned)
-> ERROR: Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.
-> $? ist 5

_, err := os.Stat("/hugo")
system.ErrorExit("Failed to revert notes: %v", err)
-> ERROR: Failed to revert notes: stat /hugo: no such file or directory
-> $? ist 1

system.ErrorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
-> ERROR: The parameters listed above have deviated from SAP/SUSE recommendations.
-> return 1


_, err := os.Stat("/hugo")
fileName := "/hugo"
system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
-> ERROR: Failed to read file '/hugo' - stat /hugo: no such file or directory
-> return 1

_, err := os.Stat("/hugo")
fileName := "/hugo"
system.ErrorExit("Failed to read file '%s' - %v", fileName, err, 6)
-> ERROR: Failed to read file '/hugo' - stat /hugo: no such file or directory
-> return 6

_, err := os.Stat("/hugo")
fileName := "/hugo"
nfileName := "/egon"
system.ErrorExit("Failed to rename file '%s' to '%s' - %v", fileName, nfileName, err)
-> ERROR: Failed to rename file '/hugo' to '/egon' - stat /hugo: no such file or directory
-> return 1

system.ErrorExit("", 0)
-> Keine Ausgabe
-> return 0

_, err := os.Stat("/hugo")
system.ErrorExit("%v", err)
-> ERROR: stat /hugo: no such file or directory
-> return 1

_, err := os.Stat("/hugo")
system.ErrorExit("%v", err, 8)
-> ERROR: stat /hugo: no such file or directory
-> return 8

*/
