package note

import (
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// section [block]

// GetBlkVal initialise the block device structure with the current
// system settings
func GetBlkVal(key string, cur *param.BlockDeviceQueue) (string, string, error) {
	newQueue := make(map[string]string)
	newReq := make(map[string]int)
	newRah := make(map[string]int)
	newMse := make(map[string]int)
	retVal := ""
	info := ""

	switch {
	case system.IsSched.MatchString(key):
		newIOQ, err := cur.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return "", info, err
		}
		newQueue = newIOQ.(param.BlockDeviceSchedulers).SchedulerChoice
		retVal = newQueue[strings.TrimPrefix(key, "IO_SCHEDULER_")]
		cur.BlockDeviceSchedulers = newIOQ.(param.BlockDeviceSchedulers)
	case system.IsNrreq.MatchString(key):
		newNrR, err := cur.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return "", info, err
		}
		newReq = newNrR.(param.BlockDeviceNrRequests).NrRequests
		retVal = strconv.Itoa(newReq[strings.TrimPrefix(key, "NRREQ_")])
		cur.BlockDeviceNrRequests = newNrR.(param.BlockDeviceNrRequests)
	case system.IsRahead.MatchString(key):
		newRahead, err := cur.BlockDeviceReadAheadKB.Inspect()
		if err != nil {
			return "", info, err
		}
		newRah = newRahead.(param.BlockDeviceReadAheadKB).ReadAheadKB
		retVal = strconv.Itoa(newRah[strings.TrimPrefix(key, "READ_AHEAD_KB_")])
		cur.BlockDeviceReadAheadKB = newRahead.(param.BlockDeviceReadAheadKB)
	case system.IsMsect.MatchString(key):
		newMsect, err := cur.BlockDeviceMaxSectorsKB.Inspect()
		if err != nil {
			return "", info, err
		}
		newMse = newMsect.(param.BlockDeviceMaxSectorsKB).MaxSectorsKB
		retVal = strconv.Itoa(newMse[strings.TrimPrefix(key, "MAX_SECTORS_KB_")])
		cur.BlockDeviceMaxSectorsKB = newMsect.(param.BlockDeviceMaxSectorsKB)
	}
	return retVal, info, nil
}

// OptBlkVal optimises the block device structure with the settings
// from the configuration file
func OptBlkVal(key, cfgval string, cur *param.BlockDeviceQueue, bOK map[string][]string) (string, string) {
	info := ""
	if cfgval == "" {
		return cfgval, info
	}
	sval := cfgval
	switch {
	case system.IsSched.MatchString(key):
		// ANGI TODO - support different scheduler per device or
		// all devices with same scheduler (oval="all none")
		oval := ""
		sfound := false
		dname := regexp.MustCompile(`^IO_SCHEDULER_(\w+)$`)
		bdev := dname.FindStringSubmatch(key)
		for _, sched := range strings.Split(cfgval, ",") {
			sval = strings.ToLower(strings.TrimSpace(sched))
			if !param.IsValidScheduler(bdev[1], sval) {
				continue
			} else {
				sfound = true
				oval = bdev[1] + " " + sval
				bOK[sval] = append(bOK[sval], bdev[1])
				break
			}
		}
		if !sfound {
			sval = cfgval
			info = "NA"
		} else {
			opt, _ := cur.BlockDeviceSchedulers.Optimise(oval)
			cur.BlockDeviceSchedulers = opt.(param.BlockDeviceSchedulers)
		}
	case system.IsNrreq.MatchString(key):
		if sval == "0" {
			sval = "1024"
		}
		ival, _ := strconv.Atoi(sval)
		opt, _ := cur.BlockDeviceNrRequests.Optimise(ival)
		cur.BlockDeviceNrRequests = opt.(param.BlockDeviceNrRequests)
	case system.IsRahead.MatchString(key):
		if sval == "0" {
			sval = "512"
		}
		ival, _ := strconv.Atoi(sval)
		opt, _ := cur.BlockDeviceReadAheadKB.Optimise(ival)
		cur.BlockDeviceReadAheadKB = opt.(param.BlockDeviceReadAheadKB)
	case system.IsMsect.MatchString(key):
		ival, _ := strconv.Atoi(sval)
		dname := regexp.MustCompile(`^MAX_SECTORS_KB_(\w+)$`)
		bdev := dname.FindStringSubmatch(key)
		maxHWsector, _ := system.GetSysInt(path.Join("block", bdev[1], "queue", "max_hw_sectors_kb"))
		if ival > maxHWsector {
			system.WarningLog("value '%v' for 'max_sectors_kb' for device '%s' is bigger than the value '%v' for 'max_hw_sectors_kb'. Limit to '%v'.", ival, bdev[1], maxHWsector, maxHWsector)
			ival = maxHWsector
			sval = strconv.Itoa(maxHWsector)
			info = "limited"
		}
		opt, _ := cur.BlockDeviceMaxSectorsKB.Optimise(ival)
		cur.BlockDeviceMaxSectorsKB = opt.(param.BlockDeviceMaxSectorsKB)
	}
	return sval, info
}

// SetBlkVal applies the settings to the system
func SetBlkVal(key, value string, cur *param.BlockDeviceQueue, revert bool) error {
	var err error

	switch {
	case system.IsSched.MatchString(key):
		if revert {
			cur.BlockDeviceSchedulers.SchedulerChoice[strings.TrimPrefix(key, "IO_SCHEDULER_")] = value
		}
		err = cur.BlockDeviceSchedulers.Apply(strings.TrimPrefix(key, "IO_SCHEDULER_"))
		if err != nil {
			return err
		}
	case system.IsNrreq.MatchString(key):
		if revert {
			ival, _ := strconv.Atoi(value)
			cur.BlockDeviceNrRequests.NrRequests[strings.TrimPrefix(key, "NRREQ_")] = ival
		}
		err = cur.BlockDeviceNrRequests.Apply(strings.TrimPrefix(key, "NRREQ_"))
		if err != nil {
			return err
		}
	case system.IsRahead.MatchString(key):
		if revert {
			ival, _ := strconv.Atoi(value)
			cur.BlockDeviceReadAheadKB.ReadAheadKB[strings.TrimPrefix(key, "READ_AHEAD_KB_")] = ival
		}
		err = cur.BlockDeviceReadAheadKB.Apply(strings.TrimPrefix(key, "READ_AHEAD_KB_"))
		if err != nil {
			return err
		}
	case system.IsMsect.MatchString(key):
		if revert {
			ival, _ := strconv.Atoi(value)
			cur.BlockDeviceMaxSectorsKB.MaxSectorsKB[strings.TrimPrefix(key, "MAX_SECTORS_KB_")] = ival
		}
		err = cur.BlockDeviceMaxSectorsKB.Apply(strings.TrimPrefix(key, "MAX_SECTORS_KB_"))
		if err != nil {
			return err
		}
	}
	return err
}
