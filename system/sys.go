package system

// Manipulate /sys/ switches.

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// GetSysString read a /sys/ key and return the string value.
func GetSysString(parameter string) (string, error) {
	val, err := ioutil.ReadFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		WarningLog("failed to read sys string key '%s': %v", parameter, err)
		return "PNA", err
	}
	return strings.TrimSpace(string(val)), nil
}

// GetSysChoice read a /sys/ key that comes with current value and alternative
// choices, return the current choice or empty string.
func GetSysChoice(parameter string) (string, error) {
	val, err := ioutil.ReadFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		WarningLog("failed to read sys key of choices '%s': %v", parameter, err)
		return "PNA", err
	}
	// Split up the choices
	allChoices := consecutiveSpaces.Split(string(val), -1)
	for _, choice := range allChoices {
		if len(choice) > 2 && choice[0] == '[' && choice[len(choice)-1] == ']' {
			return choice[1 : len(choice)-1], nil
		}
	}
	return "", nil
}

// GetSysInt read an integer /sys/ key.
func GetSysInt(parameter string) (int, error) {
	value, err := GetSysString(parameter)
	if err != nil {
		WarningLog("failed to read integer sys key '%s': %v", parameter, err)
		return 0, err
	}
	return strconv.Atoi(value)
}

// SetSysString write a string /sys/ value.
func SetSysString(parameter, value string) error {
	if value == "PNA" {
		WarningLog("value is '%s', so sys key '%s' is/was not supported by os, skipping.", value, parameter)
		return nil
	}
	err := ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 0644)
	if os.IsNotExist(err) {
		WarningLog("sys key '%s' is not supported by os, skipping.", parameter)
	} else if err != nil {
		WarningLog("failed to set sys key '%s' to string '%s': %v", parameter, value, err)
		return err
	}
	return nil
}

// SetSysInt write an integer /sys/ value.
func SetSysInt(parameter string, value int) error {
	return SetSysString(parameter, strconv.Itoa(value))
}

// TestSysString Test writing a string /sys/ value.
func TestSysString(parameter, value string) error {
	save, err := GetSysString(parameter)
	if err != nil {
		WarningLog("failed to get sys key '%s': %v", parameter, err)
		return err
	}
	if err = ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 0644); err == nil {
		// set key back to previous value, because this was only a test
		err = ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(save), 0644)
	}
	return err
}

// GetSysSearchParam returns the search pattern for a given sys key
// and the conterpart section
func GetSysSearchParam(syskey string) (string, string) {
	searchParam := ""
	sect := ""
	// blkdev
	sched := regexp.MustCompile(`block.*queue\.scheduler$`)
	nrreq := regexp.MustCompile(`block.*queue\.nr_requests$`)
	rakb := regexp.MustCompile(`block.*queue\.read_ahead_kb$`)
	mskb := regexp.MustCompile(`block.*queue\.max_sectors_kb$`)
	dev := regexp.MustCompile(`block\.(.*)\.queue\..*$`)
	d := dev.FindStringSubmatch(syskey)
	bdev := ""
	if len(d) > 0 {
		bdev = d[1]
	} else {
		dev = regexp.MustCompile(`.*_(\w+)$`)
		d = dev.FindStringSubmatch(syskey)
		if len(d) > 0 {
			bdev = d[1]
		}
	}

	switch {
	case syskey == "THP":
		searchParam = "sys:" + SysKernelTHPEnabled
		sect = "sys"
	case syskey == "sys:"+SysKernelTHPEnabled:
		searchParam = "THP"
		sect = "vm"
	case syskey == "KSM":
		searchParam = "sys:" + SysKSMRun
		sect = "sys"
	case syskey == "sys:"+SysKSMRun:
		searchParam = "KSM"
		sect = "vm"
	case IsSched.MatchString(syskey):
		searchParam = "sys:block." + bdev + ".queue.scheduler"
		sect = "sys"
	case sched.MatchString(syskey):
		searchParam = "IO_SCHEDULER_" + bdev
		sect = "block"
	case IsNrreq.MatchString(syskey):
		searchParam = "sys:block." + bdev + ".queue.nr_requests"
		sect = "sys"
	case nrreq.MatchString(syskey):
		searchParam = "NRREQ_" + bdev
		sect = "block"
	case IsRahead.MatchString(syskey):
		searchParam = "sys:block." + bdev + ".queue.read_ahead_kb"
		sect = "sys"
	case rakb.MatchString(syskey):
		searchParam = "READ_AHEAD_KB_" + bdev
		sect = "block"
	case IsMsect.MatchString(syskey):
		searchParam = "sys:block." + bdev + ".queue.max_sectors_kb"
		sect = "sys"
	case mskb.MatchString(syskey):
		searchParam = "MAX_SECTORS_KB_" + bdev
		sect = "block"
	}
	return searchParam, sect
}

// GetNrTags returns the value from /sys/block/<bdev>/mq/0/nr_tags and the
// related scheduler
func GetNrTags(key string) (int, string, string) {
	nrtags := 0
	elev := ""
	disk := ""
	dname := regexp.MustCompile(`^NRREQ_(\w+\-?\d*)$`)
	bdev := dname.FindStringSubmatch(key)
	if len(bdev) > 0 {
		nrTagsFile := path.Join("block", bdev[1], "mq", "0", "nr_tags")
		if _, err := os.Stat(path.Join("/sys", nrTagsFile)); err == nil {
			nrtags, _ = GetSysInt(nrTagsFile)
		}
		elev, _ = GetSysChoice(path.Join("block", bdev[1], "queue", "scheduler"))
		disk = bdev[1]
	}
	return nrtags, elev, disk
}
