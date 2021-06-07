package note

import (
	"github.com/SUSE/saptune/system"
	"math"
	"strconv"
)

// section [mem]

// GetMemVal initialise the shared memory structure with the current
// system settings
func GetMemVal(key string) string {
	var val string
	switch key {
	case "VSZ_TMPFS_PERCENT", "ShmFileSystemSizeMB":
		// Find out size of SHM
		mount, found := system.ParseProcMounts().GetByMountPoint("/dev/shm")
		if found {
			val = strconv.FormatUint(mount.GetFileSystemSizeMB(), 10)
			if key == "VSZ_TMPFS_PERCENT" {
				// rounded value
				percent := math.Floor(float64(mount.GetFileSystemSizeMB())*100/float64(system.GetTotalMemSizeMB()) + 0.5)
				val = strconv.FormatFloat(percent, 'f', -1, 64)
			}
		} else {
			system.WarningLog("GetMemVal: failed to find /dev/shm mount point")
			val = "-1"
		}
	}
	return val
}

// OptMemVal optimises the shared memory structure with the settings
// from the configuration file or with a calculation
func OptMemVal(key, actval, cfgval, tmpfspercent string) string {
	// tmpfspercent value of VSZ_TMPFS_PERCENT from config or override file
	var size uint64
	var ret string

	switch key {
	case "VSZ_TMPFS_PERCENT":
		ret = cfgval
	case "ShmFileSystemSizeMB":
		if actval == "-1" {
			system.WarningLog("OptMemVal: /dev/shm is not a valid mount point, will not calculate its optimal size.")
			ret = "-1"
		} else if cfgval != "0" {
			ret = cfgval
		} else {
			if tmpfspercent == "0" {
				// Calculate optimal SHM size (TotalMemSizeMB*75/100) (SAP-Note 941735)
				size = uint64(system.GetTotalMemSizeMB()) * 75 / 100
			} else {
				// Calculate optimal SHM size (TotalMemSizeMB*VSZ_TMPFS_PERCENT/100)
				val, _ := strconv.ParseUint(tmpfspercent, 10, 64)
				size = uint64(system.GetTotalMemSizeMB()) * val / 100
			}
			ret = strconv.FormatUint(size, 10)
		}
	}
	return ret
}

// SetMemVal applies the settings to the system
func SetMemVal(key, value string) error {
	var err error
	switch key {
	case "ShmFileSystemSizeMB":
		val, err := strconv.ParseUint(value, 10, 64)
		if val > 0 {
			if err = system.RemountSHM(uint64(val)); err != nil {
				return err
			}
		} else {
			system.WarningLog("SetMemVal: /dev/shm is not a valid mount point, will not adjust its size.")
		}
	}
	return err
}
