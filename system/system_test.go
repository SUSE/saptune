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
	bdevConf.BlockAttributes["sda"] = map[string]string{"IO_SCHEDULER": "mq-deadline", "NRREQ": "32", "READ_AHEAD_KB" : "512"}
	bdevConf.BlockAttributes["sdb"] = map[string]string{"IO_SCHEDULER": "bfq", "NRREQ": "64", "READ_AHEAD_KB" : "1024"}
	bdevConf.BlockAttributes["sdc"] = map[string]string{"IO_SCHEDULER": "none", "NRREQ": "4", "READ_AHEAD_KB" : "128"}

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
