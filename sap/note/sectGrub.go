package note

import (
	"github.com/SUSE/saptune/system"
	"strings"
)

// section [grub]

// GetGrubVal initialise the grub structure with the current system settings
func GetGrubVal(key string) string {
	keyFields := strings.Split(key, ":")
	val := system.ParseCmdline("/proc/cmdline", keyFields[1])
	return val
}

// OptGrubVal returns the value from the configuration file
func OptGrubVal(key, cfgval string) string {
	// nothing to do, only checking for 'verify'
	return cfgval
}

// SetGrubVal nothing to do, only checking for 'verify'
func SetGrubVal(value string) error {
	// nothing to do, only checking for 'verify'
	return nil
}
