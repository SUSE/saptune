package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"os"
	"strconv"
	"strings"
)

// section [limits]

// GetLimitsVal initialise the security limit structure with the current
// system settings
func GetLimitsVal(value string) (string, string, error) {
	info := ""
	// Find out current limits
	limit := value
	if limit != "" && limit != "NA" {
		lim := strings.Fields(limit)
		// dom=[0], type=[1], item=[2], value=[3]
		// no check, that the syntax/order of the entry in the config file is
		// a valid limits entry

		// /etc/security/limits.d/saptune-<domain>-<item>-<type>.conf
		if len(lim) < 3 {
			return "", info, fmt.Errorf("Wrong format of limit entry in Note definition file. Please check")
		} else if len(lim) == 3 {
			lim = append(lim, "")
		} else if len(lim) > 3 {
			info = getLimitInfo(lim[3])
		}
		dropInFile := fmt.Sprintf("/etc/security/limits.d/saptune-%s-%s-%s.conf", lim[0], lim[2], lim[1])
		secLimits, err := system.ParseSecLimitsFile(dropInFile)
		if err != nil {
			//ANGI TODO - check, if other files in /etc/security/limits.d contain a value for the touple "<domain>-<item>-<type>"
			return "NA", info, err
		}
		lim[3], _ = secLimits.Get(lim[0], lim[1], lim[2])
		if lim[3] == "" {
			lim[3] = "NA"
		}
		// current limit found
		limit = strings.Join(lim, " ")
	}
	return limit, info, nil
}

// OptLimitsVal optimises the security limit structure with the settings
// from the configuration file or with a calculation
func OptLimitsVal(actval, cfgval string) string {
	cfgval = strings.Join(strings.Fields(strings.TrimSpace(cfgval)), " ")
	return cfgval
}

// SetLimitsVal applies the settings to the system
func SetLimitsVal(key, noteID, value string, revert bool) error {
	var err error
	var secLimits *system.SecLimits
	limit := value
	if limit != "" && limit != "NA" {
		lim := strings.Fields(limit)
		// dom=[0], type=[1], item=[2], value=[3]

		// /etc/security/limits.d/saptune-<domain>-<item>-<type>.conf
		dropInFile := fmt.Sprintf("/etc/security/limits.d/saptune-%s-%s-%s.conf", lim[0], lim[2], lim[1])

		if revert && IsLastNoteOfParameter(key) {
			// revert - remove limits drop-in file
			os.Remove(dropInFile)
			return nil
		}

		secLimits, err = system.ParseSecLimitsFile(dropInFile)
		if err != nil {
			return err
		}

		if lim[3] != "" && lim[3] != "NA" {
			// revert with value from another former applied note
			// or
			// apply - Prepare limits drop-in file
			secLimits.Set(lim[0], lim[1], lim[2], lim[3])

			//err = secLimits.Apply()
			err = secLimits.ApplyDropIn(lim, noteID)
		}
	}
	return err
}

// getLimitInfo returns info, if config value exceeds system limit for nofile
func getLimitInfo(val string) string {
	info := ""
	if val != "" && val != "NA" && val != "infinity" && val != "unlimited" && val != "-1" {
		nrOpen, _ := system.GetSysctlInt("fs.nr_open")
		ilim, _ := strconv.Atoi(val)
		if ilim > nrOpen {
			info = "limit_exceeded"
		}
	}
	return info
}
