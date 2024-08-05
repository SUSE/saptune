package system

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

// BlockDev contains all key-value pairs of current available
// block devices in /sys/block
type BlockDev struct {
	AllBlockDevs    []string
	BlockAttributes map[string]map[string]string
}

// IsSched matches block device scheduler tag
var IsSched = regexp.MustCompile(`^IO_SCHEDULER_\w+\-?\d*$`)

// IsNrreq matches block device nrreq tag
var IsNrreq = regexp.MustCompile(`^NRREQ_\w+\-?\d*$`)

// IsRahead matches block device read_ahead_kb tag
var IsRahead = regexp.MustCompile(`^READ_AHEAD_KB_\w+\-?\d*$`)

// IsMsect matches block device max_sectors_kb tag
var IsMsect = regexp.MustCompile(`^MAX_SECTORS_KB_\w+\-?\d*$`)

var isVD = regexp.MustCompile(`^x?vd\w+$`)

// devices like /dev/nvme0n1 are the NVME storage namespaces: the devices you
// use for actual storage, which will behave essentially as disks.
// The NVMe naming standard describes:
//
//	nvme0: first registered device's device controller
//	nvme0n1: first registered device's first namespace
//	nvme0n1p1: first registered device's first namespace's first partition
var isNvme = regexp.MustCompile(`^nvme\d+n\d+$`)

// BlockDeviceIsDisk checks, if a block device is a disk
// /sys/block/*/device/type (TYPE_DISK / 0x00)
// does not work for virtio and nvme block devices, needs workaround
func BlockDeviceIsDisk(dev string) bool {
	fname := fmt.Sprintf("/sys/block/%s/device/type", dev)
	dtype, err := os.ReadFile(fname)
	if err != nil || strings.TrimSpace(string(dtype)) != "0" {
		if isVD.FindStringSubmatch(dev) == nil && isNvme.FindStringSubmatch(dev) == nil {
			// unsupported device
			return false
		}
	}
	return true
}

// GetBlockDeviceInfo reads content of stored block device information.
// content stored in SaptuneSectionDir (/run/saptune/sections)
// as blockdev.run
// Return the content as BlockDev
func GetBlockDeviceInfo() (*BlockDev, error) {
	bdevFileName := fmt.Sprintf("%s/blockdev.run", SaptuneSectionDir)
	bdevConf := &BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	content, err := os.ReadFile(bdevFileName)
	if err == nil && len(content) != 0 {
		err = json.Unmarshal(content, &bdevConf)
	}
	return bdevConf, err
}

// getValidBlockDevices reads all block devices from /sys/block
// and select the block devices, which are 'real disks' or a multipath
// device (/sys/block/*/dm/uuid starts with 'mpath-')
func getValidBlockDevices() (valDevs []string) {
	var isMpath = regexp.MustCompile(`^mpath-\w+`)
	var isLVM = regexp.MustCompile(`^LVM-\w+`)
	candidates := []string{}
	excludedevs := []string{}

	// List /sys/block and inspect the needed info of each one
	_, sysDevs := ListDir("/sys/block", "the available block devices of the system")
	for _, bdev := range sysDevs {
		dmUUID := fmt.Sprintf("/sys/block/%s/dm/uuid", bdev)
		if _, err := os.Stat(dmUUID); err == nil {
			cont, _ := os.ReadFile(dmUUID)
			if isMpath.MatchString(string(cont)) {
				candidates = append(candidates, bdev)
				_, slaves := ListDir(fmt.Sprintf("/sys/block/%s/slaves", bdev), "dm slaves")
				excludedevs = append(excludedevs, slaves...)
			} else {
				// skip not applicable devices
				if isLVM.MatchString(string(cont)) {
					InfoLog("skipping device '%s' (LVM), not applicable", bdev)
				} else {
					InfoLog("skipping device '%s', not applicable", bdev)
				}
			}
		} else {
			if !BlockDeviceIsDisk(bdev) {
				// skip not applicable devices
				InfoLog("skipping device '%s', not applicable", bdev)
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
				// skip not applicable devices
				InfoLog("skipping device '%s', not applicable for dm slaves", bdev)
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

	for _, bdev := range getValidBlockDevices() {
		// add new block device
		blockMap := make(map[string]string)

		// Remember, GetSysChoice does not accept the leading /sys/
		elev, _ := GetSysChoice(path.Join("block", bdev, "queue", "scheduler"))
		if elev == "" || elev == "NA" || elev == "PNA" {
			elev, _ = GetSysString(path.Join("block", bdev, "queue", "scheduler"))
		}
		blockMap["IO_SCHEDULER"] = elev
		val, err := os.ReadFile(path.Join("/sys/block/", bdev, "/queue/scheduler"))
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

		nrTagsFile := path.Join("block", bdev, "mq", "0", "nr_tags")
		nrtags := ""
		if _, err := os.Stat(path.Join("/sys", nrTagsFile)); err == nil {
			nrtags, _ = GetSysString(nrTagsFile)
		}
		blockMap["NR_TAGS"] = nrtags

		// VENDOR, MODEL e.g. for FUJITSU udev replacement
		vendor := ""
		model := ""
		// virtio block devices do not have useful values.
		if !isVD.MatchString(bdev) {
			vendFile := path.Join("block", bdev, "device", "vendor")
			if _, err := os.Stat(path.Join("/sys", vendFile)); err == nil {
				vendor, _ = GetSysString(vendFile)
			} else {
				InfoLog("missing vendor information for block device '%s', file '%s' does not exist.", bdev, vendFile)
			}
			modelFile := path.Join("block", bdev, "device", "model")
			if _, err := os.Stat(path.Join("/sys", modelFile)); err == nil {
				model, _ = GetSysString(modelFile)
			} else {
				InfoLog("missing model information for block device '%s', file '%s' does not exist.", bdev, modelFile)
			}
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
		return os.WriteFile(bdevFileName, content, 0644)
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
