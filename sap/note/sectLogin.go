package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

// section [login]

// GetLoginVal initialise the systemd login structure with the current
// system settings
func GetLoginVal(key string) (string, error) {
	var val string
	var utmPat = regexp.MustCompile(`UserTasksMax=(.*)`)
	switch key {
	case "UserTasksMax":
		logindContent, err := ioutil.ReadFile(path.Join(LogindConfDir, LogindSAPConfFile))
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		matches := utmPat.FindStringSubmatch(string(logindContent))
		if len(matches) != 0 {
			val = matches[1]
		} else {
			val = "NA"
		}
	}
	return val, nil
}

// OptLoginVal returns the value from the configuration file
func OptLoginVal(cfgval string) string {
	return strings.ToLower(cfgval)
}

// SetLoginVal applies the settings to the system
func SetLoginVal(key, value string, revert bool) error {
	switch key {
	case "UserTasksMax":
		// set limit per active user (for both - revert and apply)
		if value != "" && value != "NA" {
			for _, userID := range system.GetCurrentLogins() {
				if err := system.SetTasksMax(userID, value); err != nil {
					return err
				}
			}
		}
		// handle drop-in file
		if revert && IsLastNoteOfParameter(key) {
			// revert - remove logind drop-in file
			os.Remove(path.Join(LogindConfDir, LogindSAPConfFile))
			// reload-or-try-restart systemd-logind.service
			err := system.SystemctlReloadTryRestart("systemd-logind.service")
			return err
		}
		if value != "" && value != "NA" {
			// revert with value from another former applied note
			// or
			// apply - Prepare logind drop-in file
			// LogindSAPConfContent is the verbatim content of
			// SAP-specific logind settings file.
			LogindSAPConfContent := fmt.Sprintf("[Login]\nUserTasksMax=%s\n", value)
			if err := os.MkdirAll(LogindConfDir, 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path.Join(LogindConfDir, LogindSAPConfFile), []byte(LogindSAPConfContent), 0644); err != nil {
				return err
			}
			// reload-or-try-restart systemd-logind.service
			if err := system.SystemctlReloadTryRestart("systemd-logind.service"); err != nil {
				return err
			}
			if value == "infinity" {
				system.WarningLog("Be aware: system-wide UserTasksMax is now set to infinity according to SAP recommendations.\n" +
					"This opens up entire system to fork-bomb style attacks.")
			}
		}
	}
	return nil
}
