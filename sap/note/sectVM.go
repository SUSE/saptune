package note

import (
	"github.com/SUSE/saptune/system"
	"strconv"
	"strings"
)

// section [vm]
// Manipulate /sys/kernel/mm switches.

// GetVMVal initialise the memory management structure with the current
// system settings
func GetVMVal(key string) (string, string) {
	var val string
	info := ""
	switch key {
	case "THP":
		val, _ = system.GetSysChoice(system.SysKernelTHPEnabled)
	case "KSM":
		ksmval, _ := system.GetSysInt(system.SysKSMRun)
		val = strconv.Itoa(ksmval)
	}
	return val, info
}

// OptVMVal optimises the memory management structure with the settings
// from the configuration file
func OptVMVal(key, cfgval string) string {
	val := strings.ToLower(cfgval)
	switch key {
	case "THP":
		if val != "always" && val != "madvise" && val != "never" {
			system.WarningLog("wrong selection for THP. Now set to 'never' to disable transarent huge pages")
			val = "never"
		}
	case "KSM":
		if val != "1" && val != "0" {
			system.WarningLog("wrong selection for KSM. Now set to default value '0'")
			val = "0"
		}
	}
	return val
}

// SetVMVal applies the settings to the system
func SetVMVal(key, value string) error {
	var err error
	switch key {
	case "THP":
		err = system.SetSysString(system.SysKernelTHPEnabled, value)
	case "KSM":
		ksmval, _ := strconv.Atoi(value)
		err = system.SetSysInt(system.SysKSMRun, ksmval)
	}
	return err
}
