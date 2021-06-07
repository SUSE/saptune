package note

import (
	"github.com/SUSE/saptune/system"
	"strings"
)

// section [rpm]

// GetRpmVal initialise the rpm structure with the current system settings
func GetRpmVal(key string) string {
	keyFields := strings.Split(key, ":")
	instvers := system.GetRpmVers(keyFields[1])
	return instvers
}

// OptRpmVal returns the value from the configuration file
func OptRpmVal(key, cfgval string) string {
	// nothing to do, only checking for 'verify'
	return cfgval
}

// SetRpmVal nothing to do, only checking for 'verify'
func SetRpmVal(value string) error {
	// nothing to do, only checking for 'verify'
	return nil
}
