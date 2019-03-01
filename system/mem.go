package system

// Gather information about system memory and swap memory.

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// string definitions for parsing /proc/meminfo output
const (
	MemMainTotalKey = "MemTotal"
	MemSwapTotalKey = "SwapTotal"
)

// ParseMeminfo parse /proc/meminfo into key(string) - value(int) pairs.
// Panic on error.
func ParseMeminfo() (infoMap map[string]uint64) {
	infoMap = make(map[string]uint64)
	memInfo, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		panic(fmt.Errorf("failed to read /proc/meminfo: %v", err))
	}
	for _, line := range strings.Split(string(memInfo), "\n") {
		fields := consecutiveSpaces.Split(strings.TrimSpace(line), -1)
		if len(fields) <= 1 {
			continue
		}
		// First field is name with a suffix colon
		name := fields[0]
		// The second field is an integer value
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse uint64 value from '%s' in /proc/meminfo", line))
		}
		infoMap[name[0:len(name)-1]] = value
	}
	return
}

// GetMainMemSizeMB return size of system main memory, excluding swap.
// Panic on error.
func GetMainMemSizeMB() uint64 {
	return ParseMeminfo()[MemMainTotalKey] / 1024
}

// GetTotalMemSizeMB return size of system main memory plus swap.
// Panic on error.
func GetTotalMemSizeMB() uint64 {
	return (ParseMeminfo()[MemMainTotalKey] + ParseMeminfo()[MemSwapTotalKey]) / 1024
}

// GetTotalMemSizePages return size of system main memory plus swap, in pages.
// Panic on error.
func GetTotalMemSizePages() uint64 {
	return (ParseMeminfo()[MemMainTotalKey] + ParseMeminfo()[MemSwapTotalKey]) / uint64(os.Getpagesize())
}

// GetSemaphoreLimits return kernel semaphore limits. Panic on error.
func GetSemaphoreLimits() (msl, mns, opm, mni uint64) {
	field, err := GetSysctlString("kernel.sem")
	if err != nil {
		fmt.Errorf("failed to read kernel.sem values")
	}
	fields := consecutiveSpaces.Split(field, -1)
	if len(fields) < 4 {
		panic(fmt.Errorf("failed to read kernel.sem values: %v", fields))
	}
	for i, val := range []*uint64{&msl, &mns, &opm, &mni} {
		if *val, err = strconv.ParseUint(fields[i], 10, 64); err != nil {
			panic(fmt.Errorf("failed to parse kernel.sem values: %v", err))
		}
	}
	return
}
