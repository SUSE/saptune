package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

// BlockDev contains all key-value pairs of current avaliable
// block devices in /sys/block
type BlockDev struct {
	AllBlockDevs    []string
	BlockAttributes map[string]map[string]string
}

var isVD = regexp.MustCompile(`^vd\w+$`)

// devices like /dev/nvme0n1 are the NVME storage namespaces: the devices you
// use for actual storage, which will behave essentially as disks.
// The NVMe naming standard describes:
//    nvme0: first registered device's device controller
//    nvme0n1: first registered device's first namespace
//    nvme0n1p1: first registered device's first namespace's first partition
var isNvme = regexp.MustCompile(`^nvme\d+n\d+$`)

// BlockDeviceIsDisk checks, if a block device is a disk
// /sys/block/*/device/type (TYPE_DISK / 0x00)
// does not work for virtio and nvme block devices, needs workaround
func BlockDeviceIsDisk(dev string) bool {
	fname := fmt.Sprintf("/sys/block/%s/device/type", dev)
	dtype, err := ioutil.ReadFile(fname)
	if err != nil || strings.TrimSpace(string(dtype)) != "0" {
		if strings.Join(isVD.FindStringSubmatch(dev), "") == "" && strings.Join(isNvme.FindStringSubmatch(dev), "") == "" {
			// unsupported device
			return false
		}
	}
	return true
}

// GetBlockDeviceInfo reads content of stored block device information.
// content stored in SaptuneSectionDir (/var/lib/saptune/sections)
// as blockdev.run
// Return the content as BlockDev
func GetBlockDeviceInfo() (*BlockDev, error) {
	bdevFileName := fmt.Sprintf("%s/blockdev.run", SaptuneSectionDir)
	bdevConf := &BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	content, err := ioutil.ReadFile(bdevFileName)
	if err == nil && len(content) != 0 {
		err = json.Unmarshal(content, &bdevConf)
	}
	return bdevConf, err
}

// getValidBlockDevices reads all block devices from /sys/block
// and select the block devices, which are 'real disks' or a multipath
// device (/sys/block/*/dm/uuid starts with 'mpath-'
func getValidBlockDevices() (valDevs []string) {
	var isMpath = regexp.MustCompile(`^mpath-\w+`)
	var isMpathPart = regexp.MustCompile(`^part.*-mpath-\w+`)
	var isLVM = regexp.MustCompile(`^LVM-\w+`)
	candidates := []string{}
	excludedevs := []string{}

	// List /sys/block and inspect the needed info of each one
	_, sysDevs := ListDir("/sys/block", "the available block devices of the system")
	for _, bdev := range sysDevs {
		dmUUID := fmt.Sprintf("/sys/block/%s/dm/uuid", bdev)
		if _, err := os.Stat(dmUUID); err == nil {
			cont, _ := ioutil.ReadFile(dmUUID)
			if isMpath.MatchString(string(cont)) {
				candidates = append(candidates, bdev)
			}
			_, slaves := ListDir(fmt.Sprintf("/sys/block/%s/slaves", bdev), "dm slaves")
			if len(slaves) != 0 && (isMpath.MatchString(string(cont)) || isLVM.MatchString(string(cont))) && !isMpathPart.MatchString(string(cont)) {
				excludedevs = append(excludedevs, slaves...)
			}
		} else {
			if !BlockDeviceIsDisk(bdev) {
				// skip unsupported devices
				WarningLog("skipping device '%s', unsupported", bdev)
				continue
			}
			candidates = append(candidates, bdev)
		}
	}
	if len(excludedevs) == 0 {
		return candidates
	}
	for _, bdev := range candidates {
		exclude := false
		for _, edev := range excludedevs {
			if bdev == edev {
				// skip unsupported devices
				WarningLog("skipping device '%s', md slaves unsupported", bdev)
				exclude = true
				break
			}
		}
		if !exclude {
			valDevs = append(valDevs, bdev)
		}
	}
	return valDevs
}

// CollectBlockDeviceInfo collects all needed information about
// block devices from /sys/block
// write info to /var/lib/saptune/sections/block.run
func CollectBlockDeviceInfo() []string {
	bdevConf := BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}
	blockMap := make(map[string]string)

	for _, bdev := range getValidBlockDevices() {
		// add new block device
		blockMap = make(map[string]string)

		// Remember, GetSysChoice does not accept the leading /sys/
		elev, _ := GetSysChoice(path.Join("block", bdev, "queue", "scheduler"))
		blockMap["IO_SCHEDULER"] = elev
		val, err := ioutil.ReadFile(path.Join("/sys/block/", bdev, "/queue/scheduler"))
		sched := ""
		if err == nil {
			sched = string(val)
		}
		blockMap["VALID_SCHEDS"] = sched

		// Remember, GetSysString does not accept the leading /sys/
		nrreq, _ := GetSysString(path.Join("block", bdev, "queue", "nr_requests"))
		blockMap["NRREQ"] = nrreq

		readahead, _ := GetSysString(path.Join("block", bdev, "queue", "read_ahead_kb"))
		blockMap["READ_AHEAD_KB"] = readahead

		maxsectkb, _ := GetSysString(path.Join("block", bdev, "queue", "max_sectors_kb"))
		blockMap["MAX_SECTORS_KB"] = maxsectkb

		// VENDOR, MODEL e.g. for FUJITSU udev replacement
		vendor := ""
		model := ""
		// virtio block devices do not have usefull values.
		if !isVD.MatchString(bdev) {
			vendor, _ = GetSysString(path.Join("block", bdev, "device", "vendor"))
			model, _ = GetSysString(path.Join("block", bdev, "device", "model"))
		}
		blockMap["VENDOR"] = vendor
		blockMap["MODEL"] = model
		// ... more to come

		// end of sys/block loop
		// save block info
		bdevConf.BlockAttributes[bdev] = blockMap
		bdevConf.AllBlockDevs = append(bdevConf.AllBlockDevs, bdev)
	}

	err := storeBlockDeviceInfo(bdevConf)
	if err != nil {
		ErrorLog("could not store block device information - err: %v", err)
	}
	return bdevConf.AllBlockDevs
}

// storeBlockDeviceInfo stores block device information to file blockdev.run
// only used in txtparser
// storeSectionInfo stores INIFile section information to section directory
func storeBlockDeviceInfo(obj BlockDev) error {
	overwriteExisting := true
	bdevFileName := fmt.Sprintf("%s/blockdev.run", SaptuneSectionDir)

	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(SaptuneSectionDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(bdevFileName); os.IsNotExist(err) || overwriteExisting {
		return ioutil.WriteFile(bdevFileName, content, 0644)
	}
	return nil
}

// GetAvailBlockInfo returns a list of all block devices matching a special
// tag regarding block device info like VENDOR or MODEL
func GetAvailBlockInfo(info, tag string) []string {
	var blkDevConf *BlockDev
	ret := []string{}
	inf := ""
	if blkDevConf == nil || (len(blkDevConf.AllBlockDevs) == 0 && len(blkDevConf.BlockAttributes) == 0) {
		blkDevConf, _ = GetBlockDeviceInfo()
	}
	for _, entry := range blkDevConf.AllBlockDevs {
		if info == "pat" {
			inf = entry
		} else {
			inf = blkDevConf.BlockAttributes[entry][info]
		}
		if inf == "" {
			continue
		}
		match, _ := regexp.MatchString(tag, inf)
		if match {
			ret = append(ret, entry)
		}
	}
	return ret
}
