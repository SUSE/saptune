package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
)

func setupTestDirs(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()

	oldSysRunDir := SystemdSystemDropInDir
	oldSysDropIn := SystemdSystemDropInFile
	oldSysState := SystemdSystemStateFile
	oldSysComplaints := system.SystemdSystemComplaintsFile
	oldSysHigher := SystemdSystemConfDirsHigher
	oldSysLower := SystemdSystemConfDirsLower

	oldUserRunDir := SystemdUserDropInDir
	oldUserDropIn := SystemdUserDropInFile
	oldUserState := SystemdUserStateFile
	oldUserComplaints := system.SystemdUserComplaintsFile
	oldUserHigher := SystemdUserConfDirsHigher
	oldUserLower := SystemdUserConfDirsLower

	SystemdSystemDropInDir = filepath.Join(tempDir, "run/systemd/system.conf.d")
	SystemdSystemDropInFile = filepath.Join(SystemdSystemDropInDir, "80-saptune.conf")
	SystemdSystemStateFile = filepath.Join(tempDir, "run/saptune/systemd-system.conf.json")
	system.SystemdSystemComplaintsFile = filepath.Join(tempDir, "run/saptune/systemd-system-complaints.json")
	SystemdSystemConfDirsHigher = []string{filepath.Join(tempDir, "etc/systemd/system.conf.d")}
	SystemdSystemConfDirsLower = []string{filepath.Join(tempDir, "usr/lib/systemd/system.conf.d")}

	SystemdUserDropInDir = filepath.Join(tempDir, "run/systemd/user.conf.d")
	SystemdUserDropInFile = filepath.Join(SystemdUserDropInDir, "80-saptune.conf")
	SystemdUserStateFile = filepath.Join(tempDir, "run/saptune/systemd-user.conf.json")
	system.SystemdUserComplaintsFile = filepath.Join(tempDir, "run/saptune/systemd-user-complaints.json")
	SystemdUserConfDirsHigher = []string{filepath.Join(tempDir, "etc/systemd/user.conf.d")}
	SystemdUserConfDirsLower = []string{filepath.Join(tempDir, "usr/lib/systemd/user.conf.d")}

	_ = os.MkdirAll(SystemdSystemDropInDir, 0755)
	_ = os.MkdirAll(filepath.Dir(SystemdSystemStateFile), 0755)
	_ = os.MkdirAll(filepath.Dir(system.SystemdSystemComplaintsFile), 0755)
	_ = os.MkdirAll(SystemdSystemConfDirsHigher[0], 0755)
	_ = os.MkdirAll(SystemdSystemConfDirsLower[0], 0755)

	_ = os.MkdirAll(SystemdUserDropInDir, 0755)
	_ = os.MkdirAll(filepath.Dir(SystemdUserStateFile), 0755)
	_ = os.MkdirAll(filepath.Dir(system.SystemdUserComplaintsFile), 0755)
	_ = os.MkdirAll(SystemdUserConfDirsHigher[0], 0755)
	_ = os.MkdirAll(SystemdUserConfDirsLower[0], 0755)

	cleanup := func() {
		SystemdSystemDropInDir = oldSysRunDir
		SystemdSystemDropInFile = oldSysDropIn
		SystemdSystemStateFile = oldSysState
		system.SystemdSystemComplaintsFile = oldSysComplaints
		SystemdSystemConfDirsHigher = oldSysHigher
		SystemdSystemConfDirsLower = oldSysLower

		SystemdUserDropInDir = oldUserRunDir
		SystemdUserDropInFile = oldUserDropIn
		SystemdUserStateFile = oldUserState
		system.SystemdUserComplaintsFile = oldUserComplaints
		SystemdUserConfDirsHigher = oldUserHigher
		SystemdUserConfDirsLower = oldUserLower
	}

	return tempDir, cleanup
}

func TestApplySystemdConf(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	note1Entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStartSec",
			Operator: txtparser.OperatorEqual,
			Value:    "300s",
		},
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTasksMax",
			Operator: txtparser.OperatorResetAssign,
			Value:    "80%",
		},
	}

	// 1. Apply Note 123456
	chg, err := applySystemdConf(INISectionSystemdSystem, "123456", note1Entries, false)
	if err != nil {
		t.Errorf("applySystemdConf failed: %v", err)
	}
	if !chg {
		t.Errorf("Expected changed = true, got false")
	}

	// Verify content of 80-saptune.conf has Note 123456
	contentBytes, err := os.ReadFile(SystemdSystemDropInFile)
	if err != nil {
		t.Errorf("Could not read drop-in file: %v", err)
	}
	content := string(contentBytes)
	if !strings.Contains(content, "# This file is managed by saptune. Do not touch!") {
		t.Errorf("Missing managed header in drop-in")
	}
	if !strings.Contains(content, "[Manager]") {
		t.Errorf("Missing [Manager] in drop-in")
	}
	if !strings.Contains(content, "# SAP Note 123456") {
		t.Errorf("Missing note header in drop-in")
	}
	if !strings.Contains(content, "DefaultTimeoutStartSec=300s") {
		t.Errorf("Missing DefaultTimeoutStartSec=300s in drop-in")
	}
	if !strings.Contains(content, "DefaultTasksMax=\nDefaultTasksMax=80%") {
		t.Errorf("Missing reset and assignment for DefaultTasksMax in drop-in")
	}

	// 2. Apply Note 789012 (follow-up note: added to drop-in file)
	note2Entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStopSec",
			Operator: txtparser.OperatorEqual,
			Value:    "120s",
		},
	}
	chg, err = applySystemdConf(INISectionSystemdSystem, "789012", note2Entries, false)
	if err != nil {
		t.Errorf("applySystemdConf for second note failed: %v", err)
	}
	if !chg {
		t.Errorf("Expected changed = true for second note")
	}

	// Drop-in file should now contain BOTH Note 123456 and Note 789012
	contentBytes, _ = os.ReadFile(SystemdSystemDropInFile)
	content = string(contentBytes)
	if !strings.Contains(content, "# SAP Note 123456") {
		t.Errorf("Drop-in should contain Note 123456, got:\n%s", content)
	}
	if !strings.Contains(content, "DefaultTimeoutStartSec=300s") {
		t.Errorf("Drop-in should contain DefaultTimeoutStartSec=300s from Note 123456, got:\n%s", content)
	}
	if !strings.Contains(content, "# SAP Note 789012") {
		t.Errorf("Drop-in should contain Note 789012, got:\n%s", content)
	}
	if !strings.Contains(content, "DefaultTimeoutStopSec=120s") {
		t.Errorf("Missing DefaultTimeoutStopSec=120s in drop-in")
	}

	// 3. Revert Note 123456 (first note reverted: file rewritten, Note 789012 remains)
	chg, err = applySystemdConf(INISectionSystemdSystem, "123456", note1Entries, true)
	if err != nil {
		t.Errorf("Revert Note 123456 failed: %v", err)
	}
	if !chg {
		t.Errorf("Expected changed = true on revert")
	}

	contentBytes, _ = os.ReadFile(SystemdSystemDropInFile)
	content = string(contentBytes)
	if strings.Contains(content, "# SAP Note 123456") {
		t.Errorf("Drop-in should not contain reverted Note 123456, got:\n%s", content)
	}
	if !strings.Contains(content, "# SAP Note 789012") {
		t.Errorf("Drop-in should still contain Note 789012 after Note 123456 reverted, got:\n%s", content)
	}
	if !strings.Contains(content, "DefaultTimeoutStopSec=120s") {
		t.Errorf("Drop-in should still contain DefaultTimeoutStopSec=120s from Note 789012")
	}

	// 4. Revert Note 789012 (file should be removed completely)
	chg, err = applySystemdConf(INISectionSystemdSystem, "789012", note2Entries, true)
	if err != nil {
		t.Errorf("Revert Note 789012 failed: %v", err)
	}
	if !chg {
		t.Errorf("Expected changed = true on revert of last note")
	}
	if _, err := os.Stat(SystemdSystemDropInFile); !os.IsNotExist(err) {
		t.Errorf("Drop-in file should have been removed when last note reverted")
	}
}

func TestApplySystemdConfOverlappingParameters(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	note1Entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStartSec",
			Operator: txtparser.OperatorEqual,
			Value:    "100s",
		},
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTasksMax",
			Operator: txtparser.OperatorResetAssign,
			Value:    "80%",
		},
	}

	note2Entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStartSec",
			Operator: txtparser.OperatorEqual,
			Value:    "200s",
		},
	}

	// 1. Apply Note 1
	_, err := applySystemdConf(INISectionSystemdSystem, "Note1", note1Entries, false)
	if err != nil {
		t.Errorf("Apply Note1 failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(SystemdSystemDropInFile)
	content := string(contentBytes)
	if !strings.Contains(content, "DefaultTimeoutStartSec=100s") {
		t.Errorf("Expected DefaultTimeoutStartSec=100s in drop-in, got:\n%s", content)
	}

	// 2. Apply Note 2 (defines DefaultTimeoutStartSec=200s, overriding Note 1)
	_, err = applySystemdConf(INISectionSystemdSystem, "Note2", note2Entries, false)
	if err != nil {
		t.Errorf("Apply Note2 failed: %v", err)
	}

	contentBytes, _ = os.ReadFile(SystemdSystemDropInFile)
	content = string(contentBytes)
	// Latest note wins: value is 200s
	if !strings.Contains(content, "DefaultTimeoutStartSec=200s") {
		t.Errorf("Expected DefaultTimeoutStartSec=200s in drop-in, got:\n%s", content)
	}
	// Note 1 still contributes DefaultTasksMax
	if !strings.Contains(content, "DefaultTasksMax=80%") {
		t.Errorf("Expected DefaultTasksMax=80%% to be kept under Note 1, got:\n%s", content)
	}

	// 3. Revert Note 2: DefaultTimeoutStartSec should NOT be removed; it must revert to Note 1's value (100s)
	_, err = applySystemdConf(INISectionSystemdSystem, "Note2", note2Entries, true)
	if err != nil {
		t.Errorf("Revert Note2 failed: %v", err)
	}

	contentBytes, _ = os.ReadFile(SystemdSystemDropInFile)
	content = string(contentBytes)
	if !strings.Contains(content, "DefaultTimeoutStartSec=100s") {
		t.Errorf("Expected DefaultTimeoutStartSec to be kept with Note1 value 100s after reverting Note2, got:\n%s", content)
	}
	if strings.Contains(content, "DefaultTimeoutStartSec=200s") {
		t.Errorf("DefaultTimeoutStartSec=200s should have been removed after reverting Note2, got:\n%s", content)
	}

	// 4. Revert Note 1: file should be removed
	_, err = applySystemdConf(INISectionSystemdSystem, "Note1", note1Entries, true)
	if err != nil {
		t.Errorf("Revert Note1 failed: %v", err)
	}
	if _, err := os.Stat(SystemdSystemDropInFile); !os.IsNotExist(err) {
		t.Errorf("Drop-in file should have been removed when all notes reverted")
	}
}

func TestApplySystemdConfRevertFirstNote(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	note1Entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStartSec",
			Operator: txtparser.OperatorEqual,
			Value:    "100s",
		},
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTasksMax",
			Operator: txtparser.OperatorResetAssign,
			Value:    "80%",
		},
	}

	note2Entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStartSec",
			Operator: txtparser.OperatorEqual,
			Value:    "200s",
		},
	}

	// 1. Apply Note 1
	_, err := applySystemdConf(INISectionSystemdSystem, "Note1", note1Entries, false)
	if err != nil {
		t.Errorf("Apply Note1 failed: %v", err)
	}

	// 2. Apply Note 2
	_, err = applySystemdConf(INISectionSystemdSystem, "Note2", note2Entries, false)
	if err != nil {
		t.Errorf("Apply Note2 failed: %v", err)
	}

	// 3. Revert the FIRST Note (Note 1):
	// DefaultTimeoutStartSec was defined in both Note 1 and Note 2.
	// Since Note 2 is still applied, DefaultTimeoutStartSec must NOT be removed from the drop-in!
	// It must be kept with Note 2's value (200s).
	// DefaultTasksMax was only in Note 1, so it must be removed.
	_, err = applySystemdConf(INISectionSystemdSystem, "Note1", note1Entries, true)
	if err != nil {
		t.Errorf("Revert Note1 failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(SystemdSystemDropInFile)
	content := string(contentBytes)
	if !strings.Contains(content, "DefaultTimeoutStartSec=200s") {
		t.Errorf("Expected DefaultTimeoutStartSec=200s to be kept under Note2 after reverting Note1, got:\n%s", content)
	}
	if strings.Contains(content, "DefaultTasksMax") {
		t.Errorf("DefaultTasksMax should have been removed when Note1 was reverted, got:\n%s", content)
	}
	if strings.Contains(content, "# SAP Note Note1") {
		t.Errorf("Note1 block should have been removed, got:\n%s", content)
	}
	if !strings.Contains(content, "# SAP Note Note2") {
		t.Errorf("Note2 block should still be present, got:\n%s", content)
	}

	// 4. Revert Note 2: file should be removed
	_, err = applySystemdConf(INISectionSystemdSystem, "Note2", note2Entries, true)
	if err != nil {
		t.Errorf("Revert Note2 failed: %v", err)
	}
	if _, err := os.Stat(SystemdSystemDropInFile); !os.IsNotExist(err) {
		t.Errorf("Drop-in file should have been removed when all notes reverted")
	}
}

func TestApplySystemdConfEffectiveTuningTagging(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	// Create a note with a tagged section that does NOT match the running system (os=99*)
	noteDir := t.TempDir()
	oldExtra := ExtraTuningSheets
	ExtraTuningSheets = noteDir
	defer func() { ExtraTuningSheets = oldExtra }()

	noteContent := `[version]
VERSION=1
DATE=01.10.2026
DESCRIPTION=Tagged test note

[systemd-system.conf:os=99*]
DefaultTimeoutStartSec = 999s
`
	notePath := filepath.Join(noteDir, "TaggedNote99.conf")
	if err := os.WriteFile(notePath, []byte(noteContent), 0644); err != nil {
		t.Errorf("Failed to write test note: %v", err)
	}

	found, entries := getNoteEffectiveSectionEntries(INISectionSystemdSystem, "TaggedNote99")
	if !found {
		t.Errorf("Expected note to be found on disk")
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 effective entries for non-matching OS tag os=99*, got %d: %v", len(entries), entries)
	}
}

func TestGetSystemdConfValAndConflicts(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	// Case 1: Drop-in file does not exist for an unapplied note (should not report error)
	val, inform, err := GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTimeoutStartSec", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "" || inform != "missing_file" {
		t.Errorf("Expected val='' and inform='missing_file', got val='%s', inform='%s'", val, inform)
	}

	// Case 2: Drop-in file exists and key is compliant
	entries := []txtparser.INIEntry{
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStartSec",
			Operator: txtparser.OperatorEqual,
			Value:    "300s",
		},
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTimeoutStopSec",
			Operator: txtparser.OperatorEqual,
			Value:    "300s",
		},
		{
			Section:  INISectionSystemdSystem,
			Key:      "system.conf:DefaultTasksMax",
			Operator: txtparser.OperatorResetAssign,
			Value:    "80%",
		},
	}
	_, _ = applySystemdConf(INISectionSystemdSystem, "123456", entries, false)

	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTimeoutStartSec", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "300s" || inform != "" {
		t.Errorf("Expected val='300s' and inform='', got val='%s', inform='%s'", val, inform)
	}

	// Case 2b: Drop-in was generated at apply for Note 123456, but is now missing on verify
	ResetCaches()
	_ = os.Remove(SystemdSystemDropInFile)
	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTimeoutStartSec", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "" || !strings.Contains(inform, "missing_param:") {
		t.Errorf("Expected val='' and inform containing 'missing_param:', got val='%s', inform='%s'", val, inform)
	}
	// Calling for a second parameter in the same missing drop-in should also return missing_param: without duplicate error
	val2, inform2, err2 := GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTasksMax", "123456")
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}
	if val2 != "" || !strings.Contains(inform2, "missing_param:") {
		t.Errorf("Expected val2='' and inform2 containing 'missing_param:', got val2='%s', inform2='%s'", val2, inform2)
	}
	// Re-apply to restore drop-in for remaining test cases
	_, _ = applySystemdConf(INISectionSystemdSystem, "123456", entries, false)

	// Case 2c: systemd complained about a parameter during apply
	system.RecordSystemdComplaints(INISectionSystemdSystem, [][]string{{"# SAP Note 123456", "123456"}}, []string{"FooBar"})
	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:FooBar", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "" || !strings.Contains(inform, "complaint") {
		t.Errorf("Expected val='' and inform containing 'complaint', got val='%s', inform='%s'", val, inform)
	}

	// Case 2d: missing parameter in drop-in
	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:LogLevel", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "" || !strings.Contains(inform, "missing_param:") {
		t.Errorf("Expected val='' and inform containing 'missing_param:', got val='%s', inform='%s'", val, inform)
	}

	// Case 3: Conflict with higher priority drop-in (/etc/systemd/system.conf.d/22-admin.conf)
	adminConf := filepath.Join(SystemdSystemConfDirsHigher[0], "22-admin.conf")
	_ = os.WriteFile(adminConf, []byte("[Manager]\nDefaultTimeoutStopSec=60s\n"), 0644)

	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTimeoutStopSec", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "?" {
		t.Errorf("Expected val='?' on higher priority conflict, got '%s'", val)
	}
	if !strings.Contains(inform, "high:") || !strings.Contains(inform, "22-admin.conf") {
		t.Errorf("Expected inform to contain high priority conflict info, got '%s'", inform)
	}

	// Case 4: Conflict with lower priority drop-in (/usr/lib/systemd/system.conf.d/10-some-package.conf)
	pkgConf := filepath.Join(SystemdSystemConfDirsLower[0], "10-some-package.conf")
	_ = os.WriteFile(pkgConf, []byte("[Manager]\nDefaultTasksMax=1000\n"), 0644)

	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTasksMax", "123456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "80%" {
		t.Errorf("Expected val='80%%' on lower priority override, got '%s'", val)
	}
	if !strings.Contains(inform, "low:") || !strings.Contains(inform, "10-some-package.conf") {
		t.Errorf("Expected inform to contain lower priority override info, got '%s'", inform)
	}

	// Case 5: When both higher and lower priority files define the parameter, both are collected and reported
	lowerConf := filepath.Join(SystemdSystemConfDirsLower[0], "22-admin.conf")
	_ = os.WriteFile(lowerConf, []byte("[Manager]\nDefaultTimeoutStopSec=999s\n"), 0644)

	higherMatches, lowerMatches := checkSystemdConfDoubles(INISectionSystemdSystem, "DefaultTimeoutStopSec")
	if len(higherMatches) != 1 || !strings.HasSuffix(higherMatches[0], "22-admin.conf") {
		t.Errorf("Expected exactly 1 higher match for 22-admin.conf, got %v", higherMatches)
	}
	if len(lowerMatches) != 1 || !strings.HasSuffix(lowerMatches[0], "22-admin.conf") {
		t.Errorf("Expected lower priority match to be collected as well, got %v", lowerMatches)
	}

	val, inform, err = GetSystemdConfVal(INISectionSystemdSystem, "system.conf:DefaultTimeoutStopSec", "123456")
	if !strings.Contains(inform, "high:") || !strings.Contains(inform, "low:") {
		t.Errorf("Expected inform to contain both 'high:' and 'low:', got '%s'", inform)
	}
}

func TestApplySystemdConfSoftError(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	noteFile := filepath.Join(t.TempDir(), "test_systemd_note")
	content := `
[version]
VERSION=1
DATE=30.09.2026
DESCRIPTION=Test note

[systemd-system.conf]
DefaultTimeoutStartSec = 123
`
	_ = os.WriteFile(noteFile, []byte(content), 0644)
	ini := INISettings{ConfFilePath: noteFile, ID: "test1"}
	initialised, err := ini.Initialise()
	if err != nil {
		t.Errorf("Initialise failed: %v", err)
	}
	optimised, err := initialised.Optimise()
	if err != nil {
		t.Errorf("Optimise failed: %v", err)
	}

	optINI := optimised.(INISettings)
	err = optINI.Apply()
	if err != nil {
		t.Errorf("Expected nil error when daemon-reload succeeds/skipped, got %v", err)
	}
}

func TestExtractSystemdComplainedKeys(t *testing.T) {
	tempDir := t.TempDir()
	dropInFile := filepath.Join(tempDir, "80-saptune.conf")

	content := `# This file is managed by saptune. Do not touch!
[Manager]

# SAP Note test1
DefaultTimeoutStartSec=
DefaultTimeoutStartSec=123
FooBar=never
`
	if err := os.WriteFile(dropInFile, []byte(content), 0644); err != nil {
		t.Errorf("Failed to write test drop-in file: %v", err)
	}

	logs := `/run/systemd/system.conf.d/80-saptune.conf:5: Failed to parse sec value, ignoring:
/run/systemd/system.conf.d/80-saptune.conf:7: Unknown key name 'FooBar' in section [Manager], ignoring.`

	keys, _ := system.ExtractSystemdComplainedKeys(dropInFile, logs)
	if len(keys) != 2 {
		t.Errorf("Expected 2 complained keys, got %d: %v", len(keys), keys)
	}
	if keys[0] != "DefaultTimeoutStartSec" || keys[1] != "FooBar" {
		t.Errorf("Expected [DefaultTimeoutStartSec FooBar], got %v", keys)
	}
}

func TestSystemdConfListParametersCompliance(t *testing.T) {
	_, cleanup := setupTestDirs(t)
	defer cleanup()

	// Test 1: DefaultEnvironment="A=1" "B=2" (with operator =)
	noteContent1 := `[version]
VERSION=1
DATE=02.10.2026
DESCRIPTION=Test list params with =

[systemd-system.conf]
DefaultEnvironment="A=1" "B=2"
`
	notePath1 := filepath.Join(t.TempDir(), "TestListNote1")
	_ = os.WriteFile(notePath1, []byte(noteContent1), 0644)

	ini1 := INISettings{ConfFilePath: notePath1, ID: "TestListNote1"}
	insp1, err := ini1.Initialise()
	if err != nil {
		t.Errorf("Initialise failed: %v", err)
	}
	opt1, err := insp1.Optimise()
	if err != nil {
		t.Errorf("Optimise failed: %v", err)
	}
	opt1 = opt1.(INISettings).SetValuesToApply([]string{"system.conf:DefaultEnvironment"})
	err = opt1.Apply()
	if err != nil {
		t.Errorf("Apply failed: %v", err)
	}

	// Verify Note 1
	insp1After, err := ini1.Initialise()
	if err != nil {
		t.Errorf("Initialise after apply failed: %v", err)
	}
	allMatch1, comparisons1, _ := CompareNoteFields(insp1After, opt1, "TestListNote1", "forVerify")
	if !allMatch1 {
		comp := comparisons1["SysctlParams[system.conf:DefaultEnvironment]"]
		t.Errorf("Expected DefaultEnvironment with = to be compliant, but got MatchExpectation=false (act='%s', exp='%s')", comp.ActualValueJS, comp.ExpectedValueJS)
	}

	// Test 2: DefaultEnvironment=="C=1" "D=2" (with operator ==)
	noteContent2 := `[version]
VERSION=1
DATE=02.10.2026
DESCRIPTION=Test list params with ==

[systemd-system.conf]
DefaultEnvironment=="C=1" "D=2"
`
	notePath2 := filepath.Join(t.TempDir(), "TestListNote2")
	_ = os.WriteFile(notePath2, []byte(noteContent2), 0644)

	ini2 := INISettings{ConfFilePath: notePath2, ID: "TestListNote2"}
	insp2, err := ini2.Initialise()
	if err != nil {
		t.Errorf("Initialise failed: %v", err)
	}
	opt2, err := insp2.Optimise()
	if err != nil {
		t.Errorf("Optimise failed: %v", err)
	}
	opt2 = opt2.(INISettings).SetValuesToApply([]string{"system.conf:DefaultEnvironment"})
	err = opt2.Apply()
	if err != nil {
		t.Errorf("Apply failed: %v", err)
	}

	// Verify Note 2
	insp2After, err := ini2.Initialise()
	if err != nil {
		t.Errorf("Initialise after apply failed: %v", err)
	}
	allMatch2, comparisons2, _ := CompareNoteFields(insp2After, opt2, "TestListNote2", "forVerify")
	if !allMatch2 {
		comp := comparisons2["SysctlParams[system.conf:DefaultEnvironment]"]
		t.Errorf("Expected DefaultEnvironment with == to be compliant, but got MatchExpectation=false (act='%s', exp='%s')", comp.ActualValueJS, comp.ExpectedValueJS)
	}

	// Test 3: DefaultEnvironment="B=2" (single quoted value, quotes must be preserved in drop-in and verify)
	noteContent3 := `[version]
VERSION=1
DATE=02.10.2026
DESCRIPTION=Test single quoted param

[systemd-system.conf]
DefaultEnvironment="B=2"
`
	notePath3 := filepath.Join(t.TempDir(), "TestListNote3")
	_ = os.WriteFile(notePath3, []byte(noteContent3), 0644)

	ini3 := INISettings{ConfFilePath: notePath3, ID: "TestListNote3"}
	insp3, err := ini3.Initialise()
	if err != nil {
		t.Errorf("Initialise failed: %v", err)
	}
	opt3, err := insp3.Optimise()
	if err != nil {
		t.Errorf("Optimise failed: %v", err)
	}
	opt3 = opt3.(INISettings).SetValuesToApply([]string{"system.conf:DefaultEnvironment"})
	err = opt3.Apply()
	if err != nil {
		t.Errorf("Apply failed: %v", err)
	}

	dropInBytes, err := os.ReadFile(SystemdSystemDropInFile)
	if err != nil {
		t.Errorf("Failed to read drop-in file: %v", err)
	}
	if !strings.Contains(string(dropInBytes), "DefaultEnvironment=\"B=2\"") {
		t.Errorf("Expected quotation marks to be preserved in drop-in file, got:\n%s", string(dropInBytes))
	}

	// Verify Note 3
	insp3After, err := ini3.Initialise()
	if err != nil {
		t.Errorf("Initialise after apply failed: %v", err)
	}
	allMatch3, comparisons3, _ := CompareNoteFields(insp3After, opt3, "TestListNote3", "forVerify")
	if !allMatch3 {
		comp := comparisons3["SysctlParams[system.conf:DefaultEnvironment]"]
		t.Errorf("Expected DefaultEnvironment=\"B=2\" to be compliant, but got MatchExpectation=false (act='%s', exp='%s')", comp.ActualValueJS, comp.ExpectedValueJS)
	}
	comp3 := comparisons3["SysctlParams[system.conf:DefaultEnvironment]"]
	if comp3.ActualValueJS != "\"B=2\"" || comp3.ExpectedValueJS != "\"B=2\"" {
		t.Errorf("Expected quotes in ActualValueJS and ExpectedValueJS to be '\"B=2\"', got act='%s', exp='%s'", comp3.ActualValueJS, comp3.ExpectedValueJS)
	}
}
