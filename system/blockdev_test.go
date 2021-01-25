package system

import (
	"os"
	"path"
	"reflect"
	"testing"
)

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

func TestGetAvailBlockInfo(t *testing.T) {
	bdevConf := BlockDev{
		AllBlockDevs:    make([]string, 0, 6),
		BlockAttributes: make(map[string]map[string]string),
	}
	bdevConf.AllBlockDevs = []string{"sda", "sdb", "sdc"}
	bdevConf.BlockAttributes["sda"] = map[string]string{"IO_SCHEDULER": "mq-deadline", "NRREQ": "32", "READ_AHEAD_KB": "512", "VALID_SCHEDS": "[none] mq-deadline kyber bfq", "MODEL": "QEMU HARDDISK", "VENDOR": "QEMU"}
	bdevConf.BlockAttributes["sdb"] = map[string]string{"IO_SCHEDULER": "bfq", "NRREQ": "64", "READ_AHEAD_KB": "1024", "VALID_SCHEDS": "[none] mq-deadline kyber bfq", "MODEL": "QEMU HARDDISK", "VENDOR": "ATA"}
	bdevConf.BlockAttributes["sdc"] = map[string]string{"IO_SCHEDULER": "none", "NRREQ": "4", "READ_AHEAD_KB": "128", "VALID_SCHEDS": "[none] mq-deadline kyber bfq", "MODEL": "", "VENDOR": "TESTER"}

	err := storeBlockDeviceInfo(bdevConf)
	if err != nil {
		t.Error("storing block device info failed")
	}

	info := "MODEL"
	tag := ".*QEMU HARDDISK.*"
	val := []string{"sda", "sdb"}
	bdev := GetAvailBlockInfo(info, tag)
	eq := reflect.DeepEqual(bdev, val)
	if !eq {
		t.Errorf("expected:'%+v' - actual:'%+v'\n", val, bdev)
	}

	info = "VENDOR"
	tag = ".*ATA.*"
	val = []string{"sdb"}
	bdev = GetAvailBlockInfo(info, tag)
	eq = reflect.DeepEqual(bdev, val)
	if !eq {
		t.Errorf("expected:'%+v' - actual:'%+v'\n", val, bdev)
	}

	info = "HUGO"
	tag = ".*TESTER.*"
	val = []string{}
	bdev = GetAvailBlockInfo(info, tag)
	eq = reflect.DeepEqual(bdev, val)
	if !eq {
		t.Errorf("expected:'%+v' - actual:'%+v'\n", val, bdev)
	}

	bdevFile := path.Join(SaptuneSectionDir, "/blockdev.run")
	_ = os.Remove(bdevFile)
}
