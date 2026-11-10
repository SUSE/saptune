package system

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Configurable drop-in and search directories (vars to allow test redirection)
var (
	SystemdSystemDropInFile     = "/run/systemd/system.conf.d/80-saptune.conf"
	SystemdSystemComplaintsFile = "/run/saptune/sections/systemd-system-complaints.json"
	SystemdUserDropInFile       = "/run/systemd/user.conf.d/80-saptune.conf"
	SystemdUserComplaintsFile   = "/run/saptune/sections/systemd-user-complaints.json"
)

var (
	regDropInLineNum = regexp.MustCompile(`80-saptune\.conf:(\d+)`)
	regKeyInLog      = regexp.MustCompile(`(?:Unknown lvalue|Unknown key name|Unknown assignment)\s+['"]([a-zA-Z0-9_.-]+)['"]`)
	regSAPNote       = regexp.MustCompile(`# SAP Note ([a-zA-Z0-9_.-]+)`)
	regRevertNotes   = regexp.MustCompile(`revert:([a-zA-Z0-9_.-]+)`)
)

// INISectionSystemdSystem is the section name of systemd system.conf
var INISectionSystemdSystem = "systemd-system.conf"

// INISectionSystemdUser is the section name of systemd user.conf
var INISectionSystemdUser = "systemd-user.conf"

// SystemConfReloadDetection is a marker file that systemd DaemonReload needs
// to be called for 'system' at the end of saptune
var SystemConfReloadDetection = "/run/.saptune.systemConfReload"

// UserConfReloadDetection is a marker file that systemd DaemonReload needs
// to be called for 'user' at the end of saptune
var UserConfReloadDetection = "/run/.saptune.userConfReload"

// FinalSystemdDaemonReload calls systemd DaemonReload at the end of saptune
// if the marker files /run/.saptune.userConfReload and/or
// /run/.saptune.systemConfReload are available
func FinalSystemdDaemonReload() {
	DebugLog("FinalSystemdDaemonReload")
	// reload for system.conf
	action, serr := systemdDaemonReload(false)
	// reload for user.conf
	action, uerr := systemdDaemonReload(true)
	if serr != nil || uerr != nil {
		ErrorLog("systemd has detected problems and the %s might be incomplete", action)
	}
	DebugLog("FinalSystemdDaemonReload - returns")
}

// sstemdDaemonReload calls systemd DaemonReload if the marker files
// /run/.saptune.userConfReload and/or /run/.saptune.systemConfReload
// are available
func systemdDaemonReload(user bool) (string, error) {
	DebugLog("systemdDaemonReload - user is '%v'", user)
	var systemdReloadErr error
	file := SystemConfReloadDetection
	dropIn := SystemdSystemDropInFile
	section := INISectionSystemdSystem
	if user {
		file = UserConfReloadDetection
		dropIn = SystemdUserDropInFile
		section = INISectionSystemdUser
	}

	action := "tuning"

	content, err := os.ReadFile(file)
	if err == nil {
		revert := false
		noteIDs := [][]string{}
		lines := strings.Split(string(content), "\n")
		if strings.Contains(strings.Join(lines, " "), "revert:") {
			revert = true
			noteIDs = regRevertNotes.FindAllStringSubmatch(strings.Join(lines, " "), -1)
		}
		logs, err := SystemctlDaemonReloadWithLogCheck("80-saptune\\.conf", user)
		if err != nil {
			systemdReloadErr = err
		}
		if revert {
			removeSystemdComplaints(section, noteIDs)
			action = "revert"
		} else {
			complainedKeys, sapNotes := ExtractSystemdComplainedKeys(dropIn, logs)
			RecordSystemdComplaints(section, sapNotes, complainedKeys)
		}
		err = os.Remove(file)
		if err != nil {
			ErrorLogNoStdErr("removal of reload marker file '%s' failed - '%v'", file, err)
		}
	}
	DebugLog("systemdDaemonReload - returns action as '%s', systemdReloadErr as '%v'", action, systemdReloadErr)
	return action, systemdReloadErr
}

// ExtractSystemdComplainedKeys extracts parameter keys mentioned in systemd
// error/warning logs
func ExtractSystemdComplainedKeys(dropInFile, logs string) ([]string, [][]string) {
	DebugLog("ExtractSystemdComplainedKeys - dropInFile is '%s', logs is '%s'", dropInFile, logs)
	if logs == "" {
		return nil, nil
	}

	keySet := make(map[string]bool)
	sapNotes := [][]string{}

	// Pattern A: match line numbers in 80-saptune.conf
	dropInContent, err := os.ReadFile(dropInFile)
	if err == nil {
		lines := strings.Split(string(dropInContent), "\n")
		matches := regDropInLineNum.FindAllStringSubmatch(logs, -1)
		sapNotes = regSAPNote.FindAllStringSubmatch(strings.Join(lines, " "), -1)
		for _, m := range matches {
			if len(m) > 1 {
				lineNum, err := strconv.Atoi(m[1])
				if err == nil && lineNum > 0 && lineNum <= len(lines) {
					rawLine := strings.TrimSpace(lines[lineNum-1])
					if !strings.HasPrefix(rawLine, "#") && !strings.HasPrefix(rawLine, ";") && strings.Contains(rawLine, "=") {
						eqIdx := strings.Index(rawLine, "=")
						k := strings.TrimSpace(rawLine[:eqIdx])
						if k != "" {
							keySet[k] = true
						}
					}
				}
			}
		}
	}

	// Pattern B: match quoted key names in error messages
	matchesB := regKeyInLog.FindAllStringSubmatch(logs, -1)
	for _, m := range matchesB {
		if len(m) > 1 && m[1] != "" && m[1] != "Manager" {
			keySet[m[1]] = true
		}
	}

	var result []string
	for k := range keySet {
		result = append(result, k)
	}
	sort.Strings(result)
	DebugLog("ExtractSystemdComplainedKeys - returns result as '%+v', sapNotes as '%+v'", result, sapNotes)
	return result, sapNotes
}

// RecordSystemdComplaints saves complained keys for a note into /run/saptune
func RecordSystemdComplaints(section string, sapNotes [][]string, complainedKeys []string) {
	DebugLog("RecordSystemdComplaints - section is '%s', sapNotes is '%+v', complainedKeys is '%+v'", section, sapNotes, complainedKeys)
	complaintsFile := GetSystemdComplaintsFile(section)
	complaints := LoadComplaints(complaintsFile)
	for _, id := range sapNotes {
		if len(id) > 1 {
			if len(complainedKeys) > 0 {
				complaints[id[1]] = complainedKeys
			} else {
				delete(complaints, id[1])
			}
		}
	}

	if len(complaints) == 0 {
		_ = os.Remove(complaintsFile)
		DebugLog("RecordSystemdComplaints - returns after removal of '%s'", complaintsFile)
		return
	}
	_ = os.MkdirAll(filepath.Dir(complaintsFile), 0755)
	data, err := json.Marshal(complaints)
	if err == nil {
		_ = os.WriteFile(complaintsFile, data, 0644)
	}
	DebugLog("RecordSystemdComplaints - returns after writing '%s'", complaintsFile)
}

// removeSystemdComplaints removes complained keys for a note from /run/saptune
func removeSystemdComplaints(section string, noteIDs [][]string) {
	DebugLog("removeSystemdComplaints - section is '%s', noteIDs is '%+v'", section, noteIDs)
	complaintsFile := GetSystemdComplaintsFile(section)
	complaints := LoadComplaints(complaintsFile)
	for _, id := range noteIDs {
		if len(id) > 1 {
			delete(complaints, id[1])
		}
	}
	if len(complaints) == 0 {
		_ = os.Remove(complaintsFile)
		DebugLog("removeSystemdComplaints - returns after removal of '%s'", complaintsFile)
		return
	}
	data, err := json.Marshal(complaints)
	if err == nil {
		_ = os.WriteFile(complaintsFile, data, 0644)
	}
	DebugLog("removeSystemdComplaints - returns after writing '%s'", complaintsFile)
}

// GetSystemdComplaintsFile returns the complaints file path for the given
// section
func GetSystemdComplaintsFile(section string) string {
	DebugLog("GetSystemdComplaintsFile - section is '%s'", section)
	if section == INISectionSystemdUser {
		return SystemdUserComplaintsFile
	}
	return SystemdSystemComplaintsFile
}

// LoadComplaints reads the content if the complaints file
func LoadComplaints(file string) map[string][]string {
	DebugLog("LoadComplaints - file is '%s'", file)
	res := make(map[string][]string)
	data, err := os.ReadFile(file)
	if err == nil {
		_ = json.Unmarshal(data, &res)
	}
	DebugLog("LoadComplaints - returns res as '%+v'", res)
	return res
}
