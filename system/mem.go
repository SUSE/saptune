package system

// Gather information about system memory and swap memory.

import (
	"os"
	"strconv"
	"strings"
)

// string definitions for parsing /proc/meminfo output
const (
	MemMainTotalKey = "MemTotal"
	MemSwapTotalKey = "SwapTotal"
)

// ProcMeminfo is the path to the /proc/meminfo file
var ProcMeminfo = "/proc/meminfo"

// ReadMeminfo parses ProcMeminfo into key(string) - value(uint64) pairs.
func ReadMeminfo() (map[string]uint64, error) {
	DebugLog("ReadMeminfo")
	memInfo, err := os.ReadFile(ProcMeminfo)
	if err != nil {
		return nil, ErrorLog("failed to read %s: %v", ProcMeminfo, err)
	}
	infoMap := make(map[string]uint64)
	for _, line := range strings.Split(string(memInfo), "\n") {
		fields := consecutiveSpaces.Split(strings.TrimSpace(line), -1)
		if len(fields) <= 1 {
			continue
		}
		// First field is name with a suffix colon
		name := fields[0]
		if !strings.HasSuffix(name, ":") {
			continue
		}
		// The second field is an integer value
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		infoMap[name[0:len(name)-1]] = value
	}
	DebugLog("ReadMeminfo - returns infoMap as '%v'", infoMap)
	return infoMap, nil
}

// GetMeminfoEntry returns the value of a specific key from ProcMeminfo.
func GetMeminfoEntry(key string) (uint64, error) {
	DebugLog("GetMeminfoEntry - key is '%s'", key)
	infoMap, err := ReadMeminfo()
	if err != nil {
		return 0, err
	}
	val, ok := infoMap[key]
	if !ok {
		return 0, ErrorLog("key '%s' not found in %s", key, ProcMeminfo)
	}
	DebugLog("GetMeminfoEntry - returns val as '%d'", val)
	return val, nil
}

// ParseMeminfo parse /proc/meminfo into key(string) - value(int) pairs.
// Exit on error.
func ParseMeminfo() map[string]uint64 {
	DebugLog("ParseMeminfo")
	infoMap, err := ReadMeminfo()
	if err != nil {
		ErrorExit("ParseMeminfo failed: %v", err)
	}
	DebugLog("ParseMeminfo - returns infoMap as '%v'", infoMap)
	return infoMap
}

// GetMainMemSizeMB return size of system main memory, excluding swap.
// Exit on error.
func GetMainMemSizeMB() uint64 {
	InfoLog("GetMainMemSizeMB - MemTotal is '%+v'\n", ParseMeminfo()[MemMainTotalKey])
	return ParseMeminfo()[MemMainTotalKey] / 1024
}

// GetTotalMemSizeMB return size of system main memory plus swap.
// Exit on error.
func GetTotalMemSizeMB() uint64 {
	InfoLog("GetTotalMemSizeMB - MemTotal is '%+v', SwapTotal is '%+v'\n", ParseMeminfo()[MemMainTotalKey], ParseMeminfo()[MemSwapTotalKey])
	return (ParseMeminfo()[MemMainTotalKey] + ParseMeminfo()[MemSwapTotalKey]) / 1024
}

// GetTotalMemSizePages return size of system main memory plus swap, in pages.
// Exit on error.
func GetTotalMemSizePages() uint64 {
	InfoLog("GetTotalMemSizePages - MemTotal is '%+v', SwapTotal is '%+v', pagesize is '%+v'('%+v')\n", ParseMeminfo()[MemMainTotalKey], ParseMeminfo()[MemSwapTotalKey], os.Getpagesize(), uint64(os.Getpagesize()))
	return (ParseMeminfo()[MemMainTotalKey] + ParseMeminfo()[MemSwapTotalKey]) / uint64(os.Getpagesize())
}

// GetSemaphoreLimits return kernel semaphore limits. Panic on error.
func GetSemaphoreLimits() (msl, mns, opm, mni uint64, err error) {
	msl = 0
	mns = 0
	opm = 0
	mni = 0
	field, err := GetSysctlString("kernel.sem")
	if err != nil {
		ErrorLog("failed to read kernel.sem values")
		return
	}
	fields := consecutiveSpaces.Split(field, -1)
	if len(fields) < 4 {
		err = ErrorLog("failed to read kernel.sem values: %v", fields)
		return
	}
	for i, val := range []*uint64{&msl, &mns, &opm, &mni} {
		if *val, err = strconv.ParseUint(fields[i], 10, 64); err != nil {
			ErrorLog("failed to parse kernel.sem values: %v", err)
			return
		}
	}
	return
}
