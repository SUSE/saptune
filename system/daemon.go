package system

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var systemddvCmd = "/usr/bin/systemd-detect-virt"
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
	NoticeLog("SystemctlStatus - '%+v'\n", string(out))
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

// SystemdDetectVirt calls systemd-detect-virt.
// option can be '-r' (chroot), -c (container), -v (vm)
// '-r' only returns 0 or 1 without any output
func SystemdDetectVirt(opt string) (bool, string, error) {
	var out []byte
	var err error

	virt := false
	vtype := ""
	if opt == "" {
		out, err = exec.Command(systemddvCmd).CombinedOutput()
	} else {
		out, err = exec.Command(systemddvCmd, opt).CombinedOutput()
	}
	DebugLog("SystemdDetectVirt - /usr/bin/systemd-detect-virt %s : '%+v %s'", opt, err, string(out))
	if err == nil {
		// virtualized environment detected
		virt = true
	}
	if len(out) == 0 && err != nil && opt != "-r" {
		return virt, vtype, ErrorLog("%v - Failed to call systemd-detect-virt %s - %s", err, opt, string(out))
	}
	vtype = string(out)
	return virt, strings.TrimSpace(vtype), err
}

// execSystemctlCmd will execute /usr/bin/systemctl with the requested command
func execSystemctlCmd(service, cmd string) error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl %s on %s", err, cmd, service)
	}
	if running {
		out, err := exec.Command(systemctlCmd, cmd, service).CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl %s on %s - %s", err, cmd, service, string(out))
		}
		DebugLog("execSystemctlCmd, called from '%v' - /usr/bin/systemctl %s '%s' : '%+v %s'", CalledFrom(), cmd, service, err, string(out))
	}
	return nil
}

// SystemctlRestart call systemctl restart on thing.
func SystemctlRestart(thing string) error {
	return execSystemctlCmd(thing, "restart")
}

// SystemctlReloadTryRestart call systemctl reload on thing.
func SystemctlReloadTryRestart(thing string) error {
	return execSystemctlCmd(thing, "reload-or-try-restart")
}

// SystemctlStart call systemctl start on thing.
func SystemctlStart(thing string) error {
	return execSystemctlCmd(thing, "start")
}

// SystemctlStop call systemctl stop on thing.
func SystemctlStop(thing string) error {
	return execSystemctlCmd(thing, "stop")
}

// SystemctlResetFailed calls systemctl reset-failed.
func SystemctlResetFailed() error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl reset-failed", err)
	}
	if running {
		out, err := exec.Command(systemctlCmd, "reset-failed").CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl reset-failed - %s", err, string(out))
		}
		DebugLog("SystemctlResetFailed - /usr/bin/systemctl reset-failed : '%+v %s'", err, string(out))
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

// checkSystemctlState checks for a special state
func checkSystemctlState(service, cmd string) (bool, error) {
	match := false
	out, err := exec.Command(systemctlCmd, cmd, service).CombinedOutput()
	DebugLog("checkSystemctlState, called from '%v' - /usr/bin/systemctl %s %s: '%+v %s'", CalledFrom(), cmd, service, err, string(out))
	if err == nil {
		match = true
	}
	if len(out) == 0 && err != nil {
		return match, ErrorLog("%v - Failed to call systemctl %s on %s", err, cmd, service)
	}
	return match, nil
}

// SystemctlIsEnabled return true only if systemctl suggests that the thing is
// enabled.
func SystemctlIsEnabled(thing string) (bool, error) {
	return checkSystemctlState(thing, "is-enabled")
}

// SystemctlIsRunning return true only if systemctl suggests that the thing is
// running.
func SystemctlIsRunning(thing string) (bool, error) {
	return checkSystemctlState(thing, "is-active")
}

// SystemctlIsStarting return true only if systemctl suggests that the system is
// starting.
func SystemctlIsStarting() bool {
	match := false
	out, err := exec.Command(systemctlCmd, "is-system-running").CombinedOutput()
	DebugLog("SystemctlIsStarting - /usr/bin/systemctl is-system-running : '%+v %s'", err, string(out))
	if strings.TrimSpace(string(out)) == "starting" {
		DebugLog("SystemctlIsStarting - system is in state 'starting'")
		match = true
	}
	return match
}

// SystemctlIsActive returns the output of 'systemctl is-active'
func SystemctlIsActive(thing string) (string, error) {
	out, err := exec.Command(systemctlCmd, "is-active", thing).CombinedOutput()
	DebugLog("SystemctlIsActive - /usr/bin/systemctl is-active : '%+v %s'", err, string(out))
	if len(out) == 0 && err != nil {
		return "", ErrorLog("%v - Failed to call systemctl is-active", err)
	}
	return strings.TrimSpace(string(out)), err
}

// GetSystemState returns the output of 'systemctl is-system-running'
func GetSystemState() (string, error) {
	retval := ""
	out, err := exec.Command(systemctlCmd, "is-system-running").CombinedOutput()
	DebugLog("GetSystemState - /usr/bin/systemctl is-system-running : '%+v %s'", err, string(out))
	if len(out) != 0 {
		retval = strings.TrimSpace(string(out))
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
		_ = ErrorLog("Failed to call '%s %v' to get the available services - %v", systemctlCmd, strings.Join(cmdArgs, " "), err)
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

// checkStates checks, if the expcted state matches the active state
func checkStates(actStates, expStates string) (string, string) {
	start := ""
	enable := ""
	for _, state := range strings.Split(expStates, ",") {
		// expected state
		sval := strings.ToLower(strings.TrimSpace(state))
		// check for valid states. Supported for now:
		// 'start', 'stop', 'enable' and 'disable'
		if sval != "start" && sval != "stop" && sval != "enable" && sval != "disable" {
			continue
		}
		// check, if the expected state is already availabel in the active states
		start, enable = chkActStates(actStates, sval, start, enable)
	}
	return start, enable
}

// chkActStates checks, if the expected state is already availabel in the active states
func chkActStates(actStates, sval, start, enable string) (string, string) {
	match := ""
	for _, aState := range strings.Split(actStates, ",") {
		aval := strings.ToLower(strings.TrimSpace(aState))
		if sval == aval {
			match = "true"
			break
		} else {
			match = "false"
		}
	}
	// evaluate start and enable result per expected state
	start, enable = evalStartEnable(sval, match, start, enable)
	return start, enable
}

// evaluate start and enable result per expected state
func evalStartEnable(sval, match, start, enable string) (string, string) {
	if sval == "start" || sval == "stop" {
		if start != "true" {
			start = match
		}
	} else {
		if enable != "true" {
			enable = match
		}
	}
	return start, enable
}

// CmpServiceStates compares the expected service states with the current
// active service states
func CmpServiceStates(actStates, expStates string) bool {
	ret := false
	if expStates == "" {
		return true
	}
	retStart, retEnable := checkStates(actStates, expStates)
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
	err := os.WriteFile(actTunedProfile, []byte(profileName), 0644)
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
	content, err := os.ReadFile(actTunedProfile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// TunedAdmOff calls tuned-adm to switch off the active profile.
func TunedAdmOff() error {
	active, err := SystemctlIsRunning("tuned.service")
	if err != nil {
		return err
	}
	if !active {
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
		InfoLog("Failed to call tuned-adm to get the active profile - %v %s", err, string(out))
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
	active, _ := SystemctlIsRunning(sapconf)
	enabled, _ := SystemctlIsEnabled(sapconf)
	actFile1 := CmdIsAvailable("/var/lib/sapconf/act_profile")
	actEmpty1 := FileIsEmpty("/var/lib/sapconf/act_profile")
	actFile2 := CmdIsAvailable("/run/sapconf_act_profile")
	actFile3 := CmdIsAvailable("/run/sapconf/active")
	DebugLog("IsSapconfActive - sapconf is active:%+v, enabled:%+v, /var/lib/sapconf/act_profile is available:%+v, /var/lib/sapconf/act_profile is empty:%+v, /run/sapconf_act_profile is available:%+v, /run/sapconf/active is available:%+v", active, enabled, actFile1, actEmpty1, actFile2, actFile3)
	if enabled || active || !actEmpty1 || actFile2 || actFile3 {
		return true
	}
	return false
}
