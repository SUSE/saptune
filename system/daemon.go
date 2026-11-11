package system

import (
	"fmt"
	"os/exec"
	"strings"
)

var systemddvCmd = "/usr/bin/systemd-detect-virt"
var systemctlCmd = "/usr/bin/systemctl"
var journalctlCmd = "/usr/bin/journalctl"

// SystemctlEnable call systemctl enable on thing.
func SystemctlEnable(thing string) error {
	out, err := exec.Command(systemctlCmd, "enable", thing).CombinedOutput()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl enable on %s - %s", err, thing, strings.TrimSpace(string(out)))
	}
	DebugLog("SystemctlEnable - /usr/bin/systemctl enable '%s' : '%+v %s'", thing, err, strings.TrimSpace(string(out)))
	return nil
}

// SystemctlStatus call systemctl status on thing.
func SystemctlStatus(thing string) error {
	out, err := exec.Command(systemctlCmd, "status", thing).CombinedOutput()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl status on %s - %s", err, thing, strings.TrimSpace(string(out)))
	}
	NoticeLog("SystemctlStatus - '%+v'\n", strings.TrimSpace(string(out)))
	return nil
}

// SystemctlDisable call systemctl disable on thing.
func SystemctlDisable(thing string) error {
	out, err := exec.Command(systemctlCmd, "disable", thing).CombinedOutput()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl disable on %s - %s", err, thing, strings.TrimSpace(string(out)))
	}
	DebugLog("SystemctlDisable - /usr/bin/systemctl disable '%s' : '%+v %s'", thing, err, strings.TrimSpace(string(out)))
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
	DebugLog("SystemdDetectVirt - /usr/bin/systemd-detect-virt %s : '%+v %s'", opt, err, strings.TrimSpace(string(out)))
	if err == nil {
		// virtualized environment detected
		virt = true
	}
	if len(out) == 0 && err != nil && opt != "-r" {
		return virt, vtype, ErrorLog("%v - Failed to call systemd-detect-virt %s - %s", err, opt, strings.TrimSpace(string(out)))
	}
	vtype = strings.TrimSpace(string(out))
	if vtype == "none" {
		vtype = "bare-metal"
		err = nil
	}
	return virt, vtype, err
}

// execSystemctlCmd will execute /usr/bin/systemctl with the requested command
func execSystemctlCmd(service, cmd string) error {
	running, err := IsSystemRunning()
	if err != nil {
		return ErrorLog("%v - Failed to call systemctl %s on %s", err, cmd, service)
	}
	if running {
		DebugLog("execSystemctlCmd, called from '%v' - executing command '%s %s %s'", CalledFrom(), systemctlCmd, cmd, service)
		out, err := exec.Command(systemctlCmd, cmd, service).CombinedOutput()
		if err != nil {
			return ErrorLog("%v - Failed to call systemctl %s on %s - %s", err, cmd, service, strings.TrimSpace(string(out)))
		}
		DebugLog("execSystemctlCmd, called from '%v' - command '%s %s %s' returned '%+v : %s'", CalledFrom(), systemctlCmd, cmd, service, err, strings.TrimSpace(string(out)))
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
			return ErrorLog("%v - Failed to call systemctl reset-failed - %s", err, strings.TrimSpace(string(out)))
		}
		DebugLog("SystemctlResetFailed - /usr/bin/systemctl reset-failed : '%+v %s'", err, strings.TrimSpace(string(out)))
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
	DebugLog("checkSystemctlState, called from '%v' - /usr/bin/systemctl %s %s: '%+v %s'", CalledFrom(), cmd, service, err, strings.TrimSpace(string(out)))
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

// SystemctlIsActive returns the output of 'systemctl is-active'
func SystemctlIsActive(thing string) (string, error) {
	out, err := exec.Command(systemctlCmd, "is-active", thing).CombinedOutput()
	DebugLog("SystemctlIsActive - /usr/bin/systemctl is-active : '%+v %s'", err, strings.TrimSpace(string(out)))
	if len(out) == 0 && err != nil {
		return "", ErrorLog("%v - Failed to call systemctl is-active", err)
	}
	return strings.TrimSpace(string(out)), err
}

// SystemctlDaemonReload calls 'systemctl daemon-reload' or
// 'systemctl --user daemon-reload'
func SystemctlDaemonReload(user bool) error {
	var args []string
	if user {
		args = []string{"--user", "daemon-reload"}
	} else {
		args = []string{"daemon-reload"}
	}
	out, err := exec.Command(systemctlCmd, args...).CombinedOutput()
	DebugLog("SystemctlDaemonReload - %s %v: '%+v %s'", systemctlCmd, args, err, strings.TrimSpace(string(out)))
	if err != nil {
		ErrorLog("%v - Failed to call %s %v: %s", err, systemctlCmd, args, strings.TrimSpace(string(out)))
	}
	return err
}

// SystemctlDaemonReloadWithLogCheck reloads systemd configuration
// (system or user) and captures logs
func SystemctlDaemonReloadWithLogCheck(pattern string, user bool) (string, error) {
	DebugLog("SystemctlDaemonReloadWithLogCheck - pattern is '%s', user is '%v'", pattern, user)
	if !CmdIsAvailable(systemctlCmd) {
		DebugLog("systemctl command '%s' not available, skipping daemon-reload", systemctlCmd)
		return "", nil
	}

	// store systemd journal cursor
	cursor := getJournalCursor(user)
	// systemd daemon reload
	reloadErr := SystemctlDaemonReload(user)
	// capture systemd logs after cursor
	logs := ""
	if cursor != "" {
		logs = captureLogs(cursor, pattern, user)
	}

	if logs != "" {
		ErrorLog("systemd reported problems for drop-in:\n%s", logs)
		if reloadErr == nil {
			reloadErr = fmt.Errorf("systemd has detected problems and the tuning might be incomplete")
		}
	}

	DebugLog("SystemctlDaemonReloadWithLogCheck - return logs as '%s', reloadErr as '%v'", logs, reloadErr)
	return logs, reloadErr
}

// getJournalCursor returns the current cursor of systemd journal
func getJournalCursor(user bool) string {
	DebugLog("getJournalCursor - user is '%v'", user)
	cursor := ""
	if !CmdIsAvailable(journalctlCmd) {
		DebugLog("command '%s' not available, no cursor capture possible", journalctlCmd)
		return cursor
	}

	var cursorArgs []string
	if user {
		cursorArgs = []string{"--user", "-n", "0", "--show-cursor"}
	} else {
		cursorArgs = []string{"-n", "0", "--show-cursor"}
	}

	cursorCmd := exec.Command(journalctlCmd, cursorArgs...)
	cursorOut, err := cursorCmd.Output()
	if err == nil {
		outStr := strings.TrimSpace(string(cursorOut))
		lines := strings.Split(outStr, "\n")
		lastLine := lines[len(lines)-1]
		if strings.HasPrefix(lastLine, "-- cursor: ") {
			cursor = strings.TrimPrefix(lastLine, "-- cursor: ")
		}
	}
	DebugLog("getJournalCursor - return cursor as '%s'", cursor)
	return cursor
}

// captureLogs gets the logs from systemd journal (after cursor)
func captureLogs(cursor, pattern string, user bool) string {
	DebugLog("captureLogs - cursor is '%s', pattern is '%s', user is '%v'", cursor, pattern, user)
	logs := ""
	logArgs := []string{}
	if !CmdIsAvailable(journalctlCmd) {
		return logs
	}
	if user {
		logArgs = []string{"--user"}
	}
	args := append(logArgs, "--after-cursor", cursor, "-t", "systemd")
	if pattern != "" {
		args = append(args, "-g", pattern)
	}
	logCmd := exec.Command(journalctlCmd, args...)
	logOut, logErr := logCmd.Output()
	if logErr == nil {
		logs = strings.TrimSpace(string(logOut))
	}
	DebugLog("captureLogs - returns logs as '%s'", logs)
	return logs
}

// detectSystemState detects the systemctl system state
func detectSystemState() ([]byte, error) {
	out, err := exec.Command(systemctlCmd, "is-system-running").CombinedOutput()
	DebugLog("detectSystemRunning - /usr/bin/systemctl is-system-running : '%+v %s'", err, strings.TrimSpace(string(out)))
	return out, err
}

// GetSystemState returns the output of 'systemctl is-system-running'
func GetSystemState() (string, error) {
	retval := ""
	out, err := detectSystemState()
	if len(out) != 0 {
		retval = strings.TrimSpace(string(out))
	}
	return retval, err
}

// SysIsStarting return true only if systemctl suggests that the system is
// starting.
func SysIsStarting() bool {
	match := false
	out, _ := detectSystemState()
	if strings.TrimSpace(string(out)) == "starting" {
		DebugLog("SysIsStarting - system is in state 'starting'")
		match = true
	}
	return match
}

// SysIsRunning return true only if systemctl reports 'running' or
// 'degraded'
func SysIsRunning() bool {
	match := false
	out, _ := detectSystemState()
	if strings.TrimSpace(string(out)) == "running" || strings.TrimSpace(string(out)) == "degraded" {
		DebugLog("SysIsRunning - system is starting/running, match true")
		match = true
	}
	return match
}

// IsSystemRunning returns true, if 'is-system-running' reports 'running'
// 'degraded' or 'starting'. In all other cases it returns false, which means:
// do not call 'start' or 'restart' to prevent 'Transaction is destructive'
// messages
func IsSystemRunning() (bool, error) {
	match := false
	out, err := detectSystemState()
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
