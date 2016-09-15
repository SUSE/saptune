// Gather information about system memory and swap memory.
package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	MEMINFO_MAIN_TOTAL_KEY = "MemTotal"
	MEMINFO_SWAP_TOTAL_KEY = "SwapTotal"
)

// Parse /proc/meminfo into key(string) - value(int) pairs. Panic on error.
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

// Return size of system main memory, excluding swap. Panic on error.
func GetMainMemSizeMB() uint64 {
	return ParseMeminfo()[MEMINFO_MAIN_TOTAL_KEY] / 1024
}

// Return size of system main memory plus swap. Panic on error.
func GetTotalMemSizeMB() uint64 {
	return (ParseMeminfo()[MEMINFO_MAIN_TOTAL_KEY] + ParseMeminfo()[MEMINFO_SWAP_TOTAL_KEY]) / 1024
}

// Return size of system main memory plus swap, in pages. Panic on error.
func GetTotalMemSizePages() uint64 {
	return (ParseMeminfo()[MEMINFO_MAIN_TOTAL_KEY] + ParseMeminfo()[MEMINFO_SWAP_TOTAL_KEY]) / uint64(os.Getpagesize())
}

// Return kernel semaphore limits. Panic on error.
func GetSemaphoreLimits() (msl, mns, opm, mni uint64) {
	fields := consecutiveSpaces.Split(GetSysctlString("kernel.sem", ""), -1)
	if len(fields) < 4 {
		panic(fmt.Errorf("failed to read kermel.sem values: %v", fields))
	}
	var err error
	for i, val := range []*uint64{&msl, &mns, &opm, &mni} {
		if *val, err = strconv.ParseUint(fields[i], 10, 64); err != nil {
			panic(fmt.Errorf("failed to parse kermel.sem values: %v", err))
		}
	}
	return
}
