package system

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
)

var systemctlCmd = "/usr/bin/systemctl"
var tunedAdmCmd = "/usr/sbin/tuned-adm"
var actTunedProfile = "/etc/tuned/active_profile"

// SystemctlEnable call systemctl enable on thing.
func SystemctlEnable(thing string) error {
	out, err := exec.Command(systemctlCmd, "enable", thing).CombinedOutput()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl enable on %s - %s", err, thing, string(out))
	}
	DebugLog("SystemctlEnable - /usr/bin/systemctl enable '%s' : '%+v %s'", thing, err, string(out))
	return nil
}

// SystemctlStatus call systemctl status on thing.
func SystemctlStatus(thing string) error {
	out, err := exec.Command(systemctlCmd, "status", thing).CombinedOutput()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl status on %s - %s", err, thing, string(out))
	}
	InfoLog("SystemctlStatus - '%+v'\n", string(out))
	return nil
}

// SystemctlDisable call systemctl disable on thing.
func SystemctlDisable(thing string) error {
	out, err := exec.Command(systemctlCmd, "disable", thing).CombinedOutput()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl disable on %s - %s", err, thing, string(out))
	}
	DebugLog("SystemctlDisable - /usr/bin/systemctl disable '%s' : '%+v %s'", thing, err, string(out))
	return nil
}

// SystemctlRestart call systemctl restart on thing.
func SystemctlRestart(thing string) error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl restart on %s", err, thing)
	}
	if running {
		out, err := exec.Command(systemctlCmd, "restart", thing).CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl restart on %s - %s", err, thing, string(out))
		}
		DebugLog("SystemctlRestart( - /usr/bin/systemctl restart '%s' : '%+v %s'", thing, err, string(out))
	}
	return nil
}

// SystemctlReloadTryRestart call systemctl reload on thing.
func SystemctlReloadTryRestart(thing string) error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl reload-or-try-restart on %s", err, thing)
	}
	if running {
		out, err := exec.Command(systemctlCmd, "reload-or-try-restart", thing).CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl reload-or-try-restart on %s - %s", err, thing, string(out))
		}
		DebugLog("SystemctlReloadTryRestart( - /usr/bin/systemctl reload-or-try-restart '%s' : '%+v %s'", thing, err, string(out))
	}
	return nil
}

// SystemctlStart call systemctl start on thing.
func SystemctlStart(thing string) error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl start on %s", err, thing)
	}
	if running {
		out, err := exec.Command(systemctlCmd, "start", thing).CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl start on %s - %s", err, thing, string(out))
		}
		DebugLog("SystemctlStart - /usr/bin/systemctl start '%s' : '%+v %s'", thing, err, string(out))
	}
	return nil
}

// SystemctlStop call systemctl stop on thing.
func SystemctlStop(thing string) error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl stop on %s", err, thing)
	}
	if running {
		out, err := exec.Command(systemctlCmd, "stop", thing).CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl stop on %s - %s", err, thing, string(out))
		}
		DebugLog("SystemctlStop - /usr/bin/systemctl stop '%s' : '%+v %s'", thing, err, string(out))
	}
	return nil
}

// SystemctlEnableStart call systemctl enable and then systemctl start on thing.
func SystemctlEnableStart(thing string) error {
	if err := SystemctlEnable(thing); err != nil {
		return err
	}
	err := SystemctlStart(thing)
	return err
}

// SystemctlDisableStop call systemctl disable and then systemctl stop on thing.
// Panic on error.
func SystemctlDisableStop(thing string) error {
	if err := SystemctlDisable(thing); err != nil {
		return err
	}
	err := SystemctlStop(thing)
	return err
}

// SystemctlIsEnabled return true only if systemctl suggests that the thing is
// enabled.
func SystemctlIsEnabled(thing string) bool {
	if _, err := exec.Command(systemctlCmd, "is-enabled", thing).CombinedOutput(); err == nil {
		return true
	}
	return false
}

// SystemctlIsRunning return true only if systemctl suggests that the thing is
// running.
func SystemctlIsRunning(thing string) bool {
	if _, err := exec.Command(systemctlCmd, "is-active", thing).CombinedOutput(); err == nil {
		return true
	}
	return false
}


// GetSystemState returns the output of 'systemctl is-system-running'
func GetSystemState() (string, error) {
	retval := ""
	out, err := exec.Command(systemctlCmd, "is-system-running").CombinedOutput()
	DebugLog("IsSystemRunning - /usr/bin/systemctl is-system-running : '%+v %s'", err, string(out))
	if len(out) != 0 {
		retval = string(out)
	}
	return retval, err
}

// IsSystemRunning returns true, if 'is-system-running' reports 'running'
// 'degraded' or 'starting'. In all other cases it returns false, which means:
// do not call 'start' or 'restart' to prevent 'Transaction is destructive'
// messages
func IsSystemRunning() (bool, error) {
	match := false
	out, err := exec.Command(systemctlCmd, "is-system-running").CombinedOutput()
	DebugLog("IsSystemRunning - /usr/bin/systemctl is-system-running : '%+v %s'", err, string(out))
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "starting" || strings.TrimSpace(line) == "running" || strings.TrimSpace(line) == "degraded" {
			DebugLog("IsSystemRunning - system is degraded/starting/running, match true")
			match = true
			break
		}
	}
	if !match && err != nil {
		return match, ErrorLog("%v - Failed to call systemctl is-system-running", err)
	}
	return match, nil
}

// IsServiceAvailable checks, if a systemd service is available on the system
func IsServiceAvailable(service string) bool {
	match := false
	cmdArgs := []string{"--no-pager", "list-unit-files", "-t", "service"}
	cmdOut, err := exec.Command(systemctlCmd, cmdArgs...).CombinedOutput()
	if err != nil {
		return match
	}
	for _, line := range strings.Split(string(cmdOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if strings.TrimSpace(fields[0]) == service {
			match = true
			break
		}
		if strings.TrimSpace(fields[0]) == fmt.Sprintf("%s.service", service) {
			match = true
			break
		}
	}
	return match
}

// CmpServiceStates compares the expected service states with the current
// active service states
func CmpServiceStates(actStates, expStates string) bool {
	ret := false
	retStart := ""
	retEnable := ""
	if expStates == "" {
		return true
	}
	for _, state := range strings.Split(expStates, ",") {
		tmpret := ""
		sval := strings.ToLower(strings.TrimSpace(state))
		if sval != "start" && sval != "stop" && sval != "enable" && sval != "disable" {
			continue
		}
		for _, aState := range strings.Split(actStates, ",") {
			aval := strings.ToLower(strings.TrimSpace(aState))
			if sval == aval {
				tmpret = "true"
				break
			} else {
				tmpret = "false"
			}
		}
		if sval == "start" || sval == "stop" {
			if retStart != "true" {
				retStart = tmpret
			}
		} else {
			if retEnable != "true" {
				retEnable = tmpret
			}
		}
	}

	if (retStart == "" || retStart == "true") && (retEnable == "" || retEnable == "true") {
		ret = true
	}
	if retStart == "" && retEnable == "" {
		ret = false
	}
	return ret
}

// WriteTunedAdmProfile write new profile to tuned, used instead of sometimes
// unreliable 'tuned-adm' command
func WriteTunedAdmProfile(profileName string) error {
	err := ioutil.WriteFile(actTunedProfile, []byte(profileName), 0644)
	if err != nil {
		return ErrorLog("Failed to write tuned profile '%s' to '%s': %v", profileName, actTunedProfile, err)
	}
	return nil
}

// GetTunedProfile returns the currently active tuned profile by reading the
// file /etc/tuned/active_profile
// may be unreliable in newer tuned versions, so better use 'tuned-adm active'
// Return empty string if it cannot be determined.
func GetTunedProfile() string {
	content, err := ioutil.ReadFile(actTunedProfile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// TunedAdmOff calls tuned-adm to switch off the active profile.
func TunedAdmOff() error {
	if !SystemctlIsRunning("tuned.service") {
		// 'tuned-adm off' does not work without running tuned
		return nil
	}
	if out, err := exec.Command(tunedAdmCmd, "off").CombinedOutput(); err != nil {
		return ErrorLog("Failed to call tuned-adm to switch off the active profile - %v %s", err, string(out))
	}
	return nil
}

// TunedAdmProfile calls tuned-adm to switch to the specified profile.
// newer versions of tuned seems to be reliable with this command and they
// changed the behaviour/handling of the file /etc/tuned/active_profile
func TunedAdmProfile(profileName string) error {
	if out, err := exec.Command(tunedAdmCmd, "profile", profileName).CombinedOutput(); err != nil {
		return ErrorLog("Failed to call tuned-adm to active profile %s - %v %s", profileName, err, string(out))
	}
	return nil
}

// GetTunedAdmProfile return the currently active tuned profile.
// Return empty string if it cannot be determined.
func GetTunedAdmProfile() string {
	out, err := exec.Command(tunedAdmCmd, "active").CombinedOutput()
	if err != nil {
		_ = ErrorLog("Failed to call tuned-adm to get the active profile - %v %s", err, string(out))
		return ""
	}
	re := regexp.MustCompile(`Current active profile: ([\w-]+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) == 0 {
		return ""
	}
	return matches[1]
}

// IsSapconfActive checks, if sapconf is active
func IsSapconfActive(sapconf string) bool {
	if SystemctlIsEnabled(sapconf) || SystemctlIsRunning(sapconf) || CmdIsAvailable("/var/lib/sapconf/act_profile") || CmdIsAvailable("/run/sapconf/active") {
		return true
	}
	return false
}
