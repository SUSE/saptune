package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
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
		logindContent, err := os.ReadFile(path.Join(LogindConfDir, LogindSAPConfFile))
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		system.DebugLog("GetLoginVal - UserTasksMax: content of file '%s' is '%s' - error is '%v'", LogindSAPConfFile, string(logindContent), err)
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
		system.DebugLog("SetLoginVal - key is '%s', value is '%s', revert is '%v'\n", key, value, revert)
		// because of systemd problems during shutting down a node,
		// I need to change the code blocks to get UserTasksMax working
		// properly after a Reboot. To prevent left-over drop-in file
		// which will cause a wrong 'saved_state' value for UserTasksMax
		// which will result in a wrong system value after a revert of
		// a Note containing UserTasksMax setting
		//
		// So first handle drop-in file during revert
		// then set limit per active user
		//
		// Because of the changed order we need 'exitEarly', because
		// in case of removing the drop-in file we do not need to
		// execute the rest of the code

		exitEarly := false
		// handle drop-in file during revert
		if revert && IsLastNoteOfParameter(key) {
			system.DebugLog("SetLoginVal - UserTasksMax: remove drop-in file")
			exitEarly = true
			// revert - remove logind drop-in file
			os.Remove(path.Join(LogindConfDir, LogindSAPConfFile))
			// reload-or-try-restart systemd-logind.service
			if err := system.SystemctlReloadTryRestart("systemd-logind.service"); err != nil {
				return err
			}
		}
		// set limit per active user (for both - revert and apply)
		if value != "" && value != "NA" {
			for _, userID := range system.GetCurrentLogins() {
				system.DebugLog("userID is '%v'\n", userID)
				if err := system.SetTasksMax(userID, value); err != nil {
					system.DebugLog("error is '%v'\n", err)
					return err
				}
			}
		}
		if exitEarly {
			// we are in 'revert' and it's the last Note handling
			// UserTasksMax setting
			// so exit now
			return nil
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
			if err := os.WriteFile(path.Join(LogindConfDir, LogindSAPConfFile), []byte(LogindSAPConfContent), 0644); err != nil {
				return err
			}
			// reload-or-try-restart systemd-logind.service
			if err := system.SystemctlReloadTryRestart("systemd-logind.service"); err != nil {
				return err
			}
			if value == "infinity" {
				system.InfoLog("Be aware: system-wide UserTasksMax is now set to infinity according to SAP recommendations.\n" +
					"This opens up entire system to fork-bomb style attacks.")
			}
		}
	}
	return nil
}
