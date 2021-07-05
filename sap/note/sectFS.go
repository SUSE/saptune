package note

import (
	"github.com/SUSE/saptune/system"
	"strings"
)

// section [filesystem]

// GetFSVal initialise the file system management structure with the current
// system settings
func GetFSVal(key, cfgval string) (string, string) {
	val := ""
	info := ""
	switch {
	case system.IsXFSOption.MatchString(key):
		// cfgval empty, prefix -, prefix +
		if cfgval == "" {
			// empty
			return val, info
		}
		// no prefix or prefix +
		mustExist := true
		if strings.HasPrefix(cfgval, "-") {
			// prefix -
			mustExist = false
		}
		opt := strings.TrimLeft(cfgval, "+-")
		// Find out mount options
		mountOk, mountNok := system.GetMountOpts(mustExist, "xfs", opt)
		if mustExist {
			val = "+" + opt
			if len(mountNok) != 0 {
				// we have mount points missing the option
				val = "-" + opt
				info = "'" + opt + "' for FS type 'xfs' not explicit set on: " + strings.Join(mountNok, ", ")
			}
		} else {
			val = "-" + opt
			if len(mountOk) != 0 {
				// we have mount points containing the option
				val = "+" + opt
				info = "'" + opt + "' for FS type 'xfs' still explicit set on: " + strings.Join(mountOk, ", ")
			}
		}
		if len(mountOk) == 0 && len(mountNok) == 0 {
			val = "NA"
		}
	}
	return val, info
}

// OptFSVal returns the value from the configuration file
func OptFSVal(key, cfgval string) string {
	// nothing to do, only checking for 'verify'
	if cfgval != "" && !strings.HasPrefix(cfgval, "-") && !strings.HasPrefix(cfgval, "+") {
		cfgval = "+" + cfgval
	}
	return cfgval
}

// SetFSVal nothing to do, only checking for 'verify'
func SetFSVal(key, value string) error {
	// nothing to do, only checking for 'verify'
	return nil
}
