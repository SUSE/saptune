package note

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
)

// Configurable drop-in and search directories (vars to allow test redirection)
var (
	SystemdSystemDropInDir  = "/run/systemd/system.conf.d"
	SystemdSystemDropInFile = "/run/systemd/system.conf.d/80-saptune.conf"
	SystemdSystemStateFile  = "/run/saptune/sections/systemd-system.conf.json"
	SystemdUserDropInDir    = "/run/systemd/user.conf.d"
	SystemdUserDropInFile   = "/run/systemd/user.conf.d/80-saptune.conf"
	SystemdUserStateFile    = "/run/saptune/sections/systemd-user.conf.json"

	SystemdSystemConfDirsHigher = []string{
		"/etc/systemd/system.conf.d",
	}
	SystemdSystemConfDirsLower = []string{
		"/usr/local/lib/systemd/system.conf.d",
		"/usr/lib/systemd/system.conf.d",
	}

	SystemdUserConfDirsHigher = []string{
		"/etc/systemd/user.conf.d",
	}
	SystemdUserConfDirsLower = []string{
		"/usr/local/lib/systemd/user.conf.d",
		"/usr/lib/systemd/user.conf.d",
	}

	NoteTuningSheets     = "/var/lib/saptune/working/notes"
	ExtraTuningSheets    = "/etc/saptune/extra"
	OverrideTuningSheets = "/etc/saptune/override"
)

const saptuneDropInBase = "80-saptune.conf"

// GetSystemdConfVal reads the actual value from the drop-in file and checks
// for conflicts in other drop-ins.
func GetSystemdConfVal(section, key, noteID string) (string, string, error) {
	system.DebugLog("GetSystemdConfVal - section is '%s', key is '%s', noteID is '%s'", section, key, noteID)
	_, dropInFile, stateFile := getSystemdDropInPaths(section)
	cleanKey := cleanSystemdKey(key)

	// Check if the drop-in file exists
	if _, err := os.Stat(dropInFile); os.IsNotExist(err) {
		// The drop-in is dynamically created on apply if needed and
		// removed on revert.
		// It is only expected (and reported as an error) if an applied
		// Note has the new sections and indeed saptune generated this
		// drop-in at apply and it is now missing on verify.
		if wasDropInGeneratedForNote(stateFile, dropInFile, noteID) {
			if !reportedMissingDropIns[dropInFile] {
				reportedMissingDropIns[dropInFile] = true
				system.ErrorLog("Expected systemd drop-in file '%s' does not exist", dropInFile)
			}
			// When the drop-in was removed after apply, the
			// parameters that were present in the missing drop-in
			// are marked with footnote [21]
			higherMatches, lowerMatches := checkSystemdConfDoubles(section, cleanKey)
			informParts := []string{"missing_param:" + dropInFile}
			if len(higherMatches) > 0 {
				informParts = append(informParts, "high:"+strings.Join(higherMatches, ";"))
			}
			if len(lowerMatches) > 0 {
				informParts = append(informParts, "low:"+strings.Join(lowerMatches, ";"))
			}
			return "", strings.Join(informParts, "§"), nil
		}
		return "", "missing_file", nil
	}

	// 1. Did systemd complain about this key during apply?
	if isSystemdKeyComplained(section, noteID, cleanKey) {
		higherMatches, lowerMatches := checkSystemdConfDoubles(section, cleanKey)
		informParts := []string{"complaint"}
		if len(higherMatches) > 0 {
			informParts = append(informParts, "high:"+strings.Join(higherMatches, ";"))
		}
		if len(lowerMatches) > 0 {
			informParts = append(informParts, "low:"+strings.Join(lowerMatches, ";"))
		}
		return "", strings.Join(informParts, "§"), nil
	}

	// 2. Read key from drop-in file
	dropInVal, found := readDropInKey(dropInFile, cleanKey)
	if !found {
		higherMatches, lowerMatches := checkSystemdConfDoubles(section, cleanKey)
		informParts := []string{"missing_param:" + dropInFile}
		if len(higherMatches) > 0 {
			informParts = append(informParts, "high:"+strings.Join(higherMatches, ";"))
		}
		if len(lowerMatches) > 0 {
			informParts = append(informParts, "low:"+strings.Join(lowerMatches, ";"))
		}
		return "", strings.Join(informParts, "§"), nil
	}

	// 3. Key exists and was not complained about.
	// Check conflicts in other drop-ins.
	higherMatches, lowerMatches := checkSystemdConfDoubles(section, cleanKey)
	informParts := []string{}
	actVal := dropInVal
	if len(higherMatches) > 0 {
		actVal = "?"
		informParts = append(informParts, "high:"+strings.Join(higherMatches, ";"))
	}
	if len(lowerMatches) > 0 {
		informParts = append(informParts, "low:"+strings.Join(lowerMatches, ";"))
	}

	inform := strings.Join(informParts, "§")
	system.DebugLog("GetSystemdConfVal - returns actVal as '%s', inform as '%s', error as 'nil'", actVal, inform)
	return actVal, inform, nil
}

// OptSystemdConfVal returns the expected value for a systemd configuration
// tunable
func OptSystemdConfVal(val string) string {
	system.DebugLog("OptSystemdConfVal - val is '%s'", val)
	system.DebugLog("OptSystemdConfVal - returns val as '%s'", val)
	return val
}

// SetSystemdConfVal applies the settings to the system
func SetSystemdConfVal(hasSystem, hasUser, revert bool, noteID string, systemEntries []txtparser.INIEntry, userEntries []txtparser.INIEntry) []error {
	system.DebugLog("SetSystemdConfVal - hasSystem is '%v', hasUser is '%v', revert is '%v', noteID is '%s', systemEntries is '%+v', userEntries is '%+v'", hasSystem, hasUser, revert, noteID, systemEntries, userEntries)
	var sErr, uErr error
	changedSystem := false
	changedUser := false

	if hasSystem {
		changedSystem, sErr = applySystemdConf(INISectionSystemdSystem, noteID, systemEntries, revert)
	}
	if hasUser {
		changedUser, uErr = applySystemdConf(INISectionSystemdUser, noteID, userEntries, revert)
	}

	if changedSystem {
		// write /run/.saptune.systemConfReload as marker to call
		// DaemonReload at the end of saptune
		writeMarker(system.SystemConfReloadDetection, "system.conf reload", noteID, revert)
	}
	if changedUser {
		// write /run/.saptune.userConfReload as marker to call
		// DaemonReload at the end of saptune
		writeMarker(system.UserConfReloadDetection, "user.conf reload", noteID, revert)
	}
	system.DebugLog("SetSystemdConfVal - returns sErr as '%v', uErr as'%v'", sErr, uErr)
	return []error{sErr, uErr}
}

// writeMarker creates a marker file in /run
func writeMarker(file, txt, noteID string, revert bool) {
	system.DebugLog("writeMarker - file is '%s', txt is '%s', noteID is '%s', revert is '%v'", file, txt, noteID, revert)
	entry := "" // content for apply
	if revert {
		entry = "revert:" + noteID + "\n"
	}
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		system.ErrorLog("Problems to open file '%s' to write marker content '%s'.", file, entry)
		return
	}
	defer fd.Close()
	if _, err = fd.WriteString(entry); err != nil {
		system.ErrorLog("problems during append of entry '%s' to file '%s'.", entry, file)
	}
	system.DebugLog("writeMarker - returns")
}

// cleanSystemdKey strips the (system.conf) or (user.conf) suffix from a key
// name
func cleanSystemdKey(key string) string {
	system.DebugLog("cleanSystemdKey - key is '%s'", key)
	key = strings.TrimPrefix(key, "system.conf:")
	key = strings.TrimPrefix(key, "user.conf:")
	system.DebugLog("cleanSystemdKey - returns key as '%s'", strings.TrimSpace(key))
	return strings.TrimSpace(key)
}

// getSystemdDropInPaths returns the drop-in dir, file path, and state file
// path for the given section
func getSystemdDropInPaths(section string) (string, string, string) {
	system.DebugLog("getSystemdDropInPaths - section is '%s'", section)
	if section == INISectionSystemdUser {
		return SystemdUserDropInDir, SystemdUserDropInFile, SystemdUserStateFile
	}
	return SystemdSystemDropInDir, SystemdSystemDropInFile, SystemdSystemStateFile
}

// wasDropInGeneratedForNote checks if saptune generated a drop-in for this
// note at apply
func wasDropInGeneratedForNote(stateFile, dropInFile, noteID string) bool {
	system.DebugLog("wasDropInGeneratedForNote - stateFile is '%s', dropInFile is '%s', noteID is '%s'", stateFile, dropInFile, noteID)
	if noteID == "" {
		return false
	}
	appliedNotes := loadAppliedNotes(stateFile, dropInFile)
	for _, n := range appliedNotes {
		if n.NoteID == noteID {
			system.DebugLog("wasDropInGeneratedForNote - returns true")
			return true
		}
	}
	system.DebugLog("wasDropInGeneratedForNote - returns false")
	return false
}

// isSystemdKeyComplained checks if systemd complained about cleanKey during
// apply
func isSystemdKeyComplained(section, noteID, cleanKey string) bool {
	system.DebugLog("isSystemdKeyComplained - section is '%s', noteID is '%s', cleanKey is '%s'", section, noteID, cleanKey)
	complaintsFile := system.GetSystemdComplaintsFile(section)
	complaints := system.LoadComplaints(complaintsFile)
	if keys, ok := complaints[noteID]; ok {
		for _, k := range keys {
			if k == cleanKey {
				system.DebugLog("isSystemdKeyComplained - returns 'true'")
				return true
			}
		}
	}
	// Also check if any note recorded a complaint for this key
	for _, keys := range complaints {
		for _, k := range keys {
			if k == cleanKey {
				system.DebugLog("isSystemdKeyComplained - returns 'true'")
				return true
			}
		}
	}
	system.DebugLog("isSystemdKeyComplained - returns 'false'")
	return false
}

// readDropInKey extracts the latest assigned value of a key from a drop-in file
func readDropInKey(filePath, targetKey string) (string, bool) {
	system.DebugLog("readDropInKey - filePath is '%s', targetKey is '%s'", filePath, targetKey)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", false
	}
	lines := strings.Split(string(content), "\n")
	found := false
	val := ""
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if len(line) == 0 || line[0] == '#' || line[0] == ';' || line[0] == '[' {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			continue
		}
		k := strings.TrimSpace(line[:eqIdx])
		if k == targetKey {
			found = true
			val = strings.TrimLeft(line[eqIdx+1:], " \t")
		}
	}
	system.DebugLog("readDropInKey - returns val as '%s', found as '%v'", val, found)
	return val, found
}

// appendUnique append item to slice only if it's not already there
func appendUnique(slice []string, item string) []string {
	system.DebugLog("appendUnique - slice is '%+v', item is '%s'", slice, item)
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	system.DebugLog("appendUnique - returns")
	return append(slice, item)
}

// checkSystemdConfDoubles scans drop-in configuration directories for
// occurrences of cleanKey,
// categorizing matches into higher or lower priority than 80-saptune.conf.
func checkSystemdConfDoubles(section, cleanKey string) ([]string, []string) {
	system.DebugLog("checkSystemdConfDoubles - section is '%s', cleanKey is '%s'", section, cleanKey)
	var higherMatches []string
	var lowerMatches []string

	runtimeDir, dropInFile, _ := getSystemdDropInPaths(section)

	var higherDirs []string
	if section == INISectionSystemdUser {
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			userHomeDir := filepath.Join(home, ".config", "systemd", "user.conf.d")
			higherDirs = append(higherDirs, userHomeDir)
		}
		higherDirs = append(higherDirs, SystemdUserConfDirsHigher...)
	} else {
		higherDirs = append(higherDirs, SystemdSystemConfDirsHigher...)
	}

	var lowerDirs []string
	if section == INISectionSystemdUser {
		lowerDirs = append(lowerDirs, SystemdUserConfDirsLower...)
	} else {
		lowerDirs = append(lowerDirs, SystemdSystemConfDirsLower...)
	}

	// 1. Scan Higher Priority Directories (/etc, ~/.config)
	for _, dir := range higherDirs {
		files := getConfFiles(dir)
		for _, f := range files {
			if fileContainsSystemdKey(f, cleanKey) {
				higherMatches = appendUnique(higherMatches, f)
			}
		}
	}

	// 2. Scan Runtime Directory (/run/systemd/{system,user}.conf.d)
	runtimeFiles := getConfFiles(runtimeDir)
	for _, f := range runtimeFiles {
		if f == dropInFile {
			continue // Saptune's own drop-in
		}
		base := filepath.Base(f)
		if fileContainsSystemdKey(f, cleanKey) {
			if base > saptuneDropInBase {
				higherMatches = appendUnique(higherMatches, f)
			} else if base < saptuneDropInBase {
				lowerMatches = appendUnique(lowerMatches, f)
			}
		}
	}

	// 3. Scan Lower Priority Directories (/usr/local/lib, /usr/lib)
	for _, dir := range lowerDirs {
		files := getConfFiles(dir)
		for _, f := range files {
			if fileContainsSystemdKey(f, cleanKey) {
				lowerMatches = appendUnique(lowerMatches, f)
			}
		}
	}

	system.DebugLog("checkSystemdConfDoubles - returns higherMatches as '%+v', lowerMatches as '%+v'", higherMatches, lowerMatches)
	return higherMatches, lowerMatches
}

// getConfFiles returns config files from a given directory
func getConfFiles(dir string) []string {
	system.DebugLog("getConfFiles - dir is '%s'", dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".conf") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	system.DebugLog("getConfFiles - returns files as '%+v'", files)
	return files
}

// fileContainsSystemdKey returns true, if the config file contains the given
// systemd key
func fileContainsSystemdKey(filePath, targetKey string) bool {
	system.DebugLog("fileContainsSystemdKey - filePath is '%s', targetKey is '%s'", filePath, targetKey)
	content, err := os.ReadFile(filePath)
	if err != nil {
		system.DebugLog("fileContainsSystemdKey - returns 'false'")
		return false
	}
	lines := strings.Split(string(content), "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if len(line) == 0 || line[0] == '#' || line[0] == ';' || line[0] == '[' {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			continue
		}
		k := strings.TrimSpace(line[:eqIdx])
		if k == targetKey {
			system.DebugLog("fileContainsSystemdKey - returns 'true'")
			return true
		}
	}
	system.DebugLog("fileContainsSystemdKey - returns 'false'")
	return false
}

// SystemdNoteBlock represents a block of configuration lines in 80-saptune.conf
// for a specific Note
type SystemdNoteBlock struct {
	NoteID string   `json:"note_id"`
	Lines  []string `json:"lines"`
}

// AppliedNoteEntries stores raw note entries for multi-note parameter
// management
type AppliedNoteEntries struct {
	NoteID  string               `json:"note_id"`
	Entries []txtparser.INIEntry `json:"entries"`
}

// loadAppliedNotes loads note entries from the state file
func loadAppliedNotes(stateFile, dropInFile string) []AppliedNoteEntries {
	system.DebugLog("loadAppliedNotes - stateFile is '%s', dropInFile is '%s'", stateFile, dropInFile)
	data, err := os.ReadFile(stateFile)
	if err == nil {
		var notes []AppliedNoteEntries
		if err := json.Unmarshal(data, &notes); err == nil && len(notes) > 0 {
			system.DebugLog("loadAppliedNotes - returns notes as '%+v'", notes)
			return notes
		}
	}
	// Fallback: parse from dropInFile if stateFile does not exist yet
	contentBytes, err := os.ReadFile(dropInFile)
	if err == nil {
		blocks := parseDropInBlocks(string(contentBytes))
		var notes []AppliedNoteEntries
		for _, b := range blocks {
			var entries []txtparser.INIEntry
			for _, l := range b.Lines {
				eqIdx := strings.Index(l, "=")
				if eqIdx == -1 {
					continue
				}
				k := strings.TrimSpace(l[:eqIdx])
				v := strings.TrimLeft(l[eqIdx+1:], " \t")
				if k != "" && v != "" {
					entries = append(entries, txtparser.INIEntry{
						Key:      k,
						Operator: txtparser.OperatorEqual,
						Value:    v,
					})
				}
			}
			if len(entries) > 0 {
				notes = append(notes, AppliedNoteEntries{
					NoteID:  b.NoteID,
					Entries: entries,
				})
			}
		}
		system.DebugLog("loadAppliedNotes - returns notes as '%+v'", notes)
		return notes
	}
	system.DebugLog("loadAppliedNotes - returns 'nil'")
	return nil
}

// saveAppliedNotes saves note entries to the state file
func saveAppliedNotes(stateFile string, notes []AppliedNoteEntries) error {
	system.DebugLog("saveAppliedNotes - stateFile is '%s', notes is '%+v'", stateFile, notes)
	if len(notes) == 0 {
		_ = os.Remove(stateFile)
		return nil
	}
	_ = os.MkdirAll(filepath.Dir(stateFile), 0755)
	data, err := json.Marshal(notes)
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0644)
}

// parseDropInBlocks parses an existing 80-saptune.conf into blocks per Note
func parseDropInBlocks(content string) []SystemdNoteBlock {
	system.DebugLog("parseDropInBlocks - content is '%s'", content)
	var blocks []SystemdNoteBlock
	var currentBlock *SystemdNoteBlock

	lines := strings.Split(content, "\n")
	for _, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "# SAP Note ") {
			if currentBlock != nil {
				blocks = append(blocks, *currentBlock)
			}
			noteID := strings.TrimSpace(strings.TrimPrefix(trimmed, "# SAP Note "))
			currentBlock = &SystemdNoteBlock{
				NoteID: noteID,
				Lines:  []string{},
			}
			continue
		}
		if currentBlock != nil {
			if trimmed != "" {
				currentBlock.Lines = append(currentBlock.Lines, trimmed)
			}
		}
	}
	if currentBlock != nil {
		blocks = append(blocks, *currentBlock)
	}
	system.DebugLog("parseDropInBlocks - returns blocks as '%+v'", blocks)
	return blocks
}

// renderDropInContent renders 80-saptune.conf containing all applied note
// blocks
func renderDropInContent(blocks []SystemdNoteBlock) string {
	system.DebugLog("renderDropInContent - blocks is '%+v'", blocks)
	var sb strings.Builder
	sb.WriteString("# This file is managed by saptune. Do not touch!\n[Manager]\n")
	for _, b := range blocks {
		if len(b.Lines) == 0 {
			continue
		}
		sb.WriteString("\n# SAP Note " + b.NoteID + "\n")
		for _, l := range b.Lines {
			sb.WriteString(l + "\n")
		}
	}
	system.DebugLog("renderDropInContent - returns sb as '%s'", sb.String())
	return sb.String()
}

// getNoteEffectiveSectionEntries parses the note file and override, returning
// whether the note was found on disk and its effective entries for a section
func getNoteEffectiveSectionEntries(section, noteID string) (bool, []txtparser.INIEntry) {
	system.DebugLog("getNoteEffectiveSectionEntries - section is '%s', noteID is '%s'", section, noteID)
	if noteID == "" {
		system.DebugLog("getNoteEffectiveSectionEntries - returns 'false', 'nil', because of empty NoteID")
		return false, nil
	}

	parsed, err := getParsedNote(noteID)
	if err != nil || parsed == nil {
		system.DebugLog("getNoteEffectiveSectionEntries - returns 'false', 'nil', because Note file not found or failed to parse")
		return false, nil // Note file not found or failed to parse
	}

	// Check for overrides
	override, ow := txtparser.GetOverrides("ovw", noteID)

	var entries []txtparser.INIEntry
	for _, entry := range parsed.AllValues {
		if entry.Section == section {
			if override && ow != nil && len(ow.KeyValue[section]) > 0 {
				if owEntry, ok := ow.KeyValue[section][entry.Key]; ok {
					if owEntry.Value == "" || owEntry.Value == "untouched" {
						continue
					}
					entry.Value = owEntry.Value
					if owEntry.Operator != "" {
						entry.Operator = owEntry.Operator
					}
				}
			}
			if entry.Value != "" && entry.Value != "untouched" {
				entries = append(entries, entry)
			}
		}
	}
	system.DebugLog("getNoteEffectiveSectionEntries - returns 'true', entries as '%+v'", entries)
	return true, entries
}

// getParsedNote is a helper that returns a cached or newly parsed INI file for
// a given note ID.
func getParsedNote(noteID string) (*txtparser.INIFile, error) {
	system.DebugLog("getParsedNote - noteID is '%s'", noteID)
	if cached, exists := parsedNoteCache[noteID]; exists {
		system.DebugLog("getParsedNote - returns cached as '%+v', 'nil'", cached)
		return cached, nil
	}

	var noteFile string
	candidates := []string{
		filepath.Join(ExtraTuningSheets, noteID+".conf"),
		filepath.Join(NoteTuningSheets, noteID),
		filepath.Join("/usr/share/saptune/notes", noteID),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			noteFile = c
			break
		}
	}
	if noteFile == "" {
		system.DebugLog("getParsedNote - returns 'nil', error 'note file for ID %s not found'", noteID)
		return nil, system.ErrorLog("note file for ID %s not found", noteID)
	}

	parsed, err := txtparser.ParseINIFile(noteFile, false)
	if err != nil {
		system.DebugLog("getParsedNote - returns 'nil', err as '%v'", err)
		return nil, err
	}

	parsedNoteCache[noteID] = parsed
	system.DebugLog("getParsedNote - returns parsed as '%+v', 'nil'", parsed)
	return parsed, nil
}

// applySystemdConf creates, updates, or reverts drop-in entries for a given
// section and note.
// If a parameter is defined in multiple Notes, the latest Note wins and its
// value is written in the drop-in.
// When a Note is reverted, parameters that are also defined in other applied
// Notes are kept with the value of the now latest Note.
// Returns (changed bool, err error).
func applySystemdConf(section, noteID string, entries []txtparser.INIEntry, revert bool) (bool, error) {
	system.DebugLog("applySystemdConf - section is '%s', noteID is '%s', entries is '%+v', revert is '%v'", section, noteID, entries, revert)
	dropInDir, dropInFile, stateFile := getSystemdDropInPaths(section)

	appliedNotes := loadAppliedNotes(stateFile, dropInFile)
	_, dropInStatErr := os.Stat(dropInFile)
	dropInExists := (dropInStatErr == nil)

	if revert {
		if !dropInExists && len(appliedNotes) == 0 {
			system.NoticeLog("Drop-in file '%s' does not exist during revert", dropInFile)
			system.DebugLog("applySystemdConf - returns 'false', 'nil'")
			return false, nil
		}

		// Remove noteID from appliedNotes
		var newApplied []AppliedNoteEntries
		found := false
		for _, n := range appliedNotes {
			if n.NoteID == noteID {
				found = true
				continue
			}
			newApplied = append(newApplied, n)
		}
		if !found {
			// Note had no entries in this section
			system.DebugLog("applySystemdConf - returns 'false', 'nil', because Note had no entries in this section '%s'", section)
			return false, nil
		}
		appliedNotes = newApplied
	} else {
		// Apply mode: filter out untouched / empty entries
		var validEntries []txtparser.INIEntry
		for _, entry := range entries {
			if entry.Value == "" || entry.Value == "untouched" {
				continue
			}
			validEntries = append(validEntries, entry)
		}
		if len(validEntries) == 0 {
			system.DebugLog("applySystemdConf - returns 'false', 'nil'")
			return false, nil
		}

		// Update or append noteID in appliedNotes
		// (preserving sequence, latest at the end)
		updated := false
		for i, n := range appliedNotes {
			if n.NoteID == noteID {
				appliedNotes[i].Entries = validEntries
				curr := appliedNotes[i]
				appliedNotes = append(appliedNotes[:i], appliedNotes[i+1:]...)
				appliedNotes = append(appliedNotes, curr)
				updated = true
				break
			}
		}
		if !updated {
			appliedNotes = append(appliedNotes, AppliedNoteEntries{
				NoteID:  noteID,
				Entries: validEntries,
			})
		}
	}

	// Save updated state
	if err := saveAppliedNotes(stateFile, appliedNotes); err != nil {
		system.ErrorLog("Failed to save systemd state to %s: %w", stateFile, err)
		system.DebugLog("applySystemdConf - returns 'false', err as '%v'", err)
		return false, err
	}

	// If no applied notes remain for this section, remove the drop-in file
	if len(appliedNotes) == 0 {
		if err := os.Remove(dropInFile); err != nil && !os.IsNotExist(err) {
			system.ErrorLog("Failed to remove drop-in file '%s': %v", dropInFile, err)
			system.DebugLog("applySystemdConf - returns 'false', err as '%v'", err)
			return false, err
		}
		system.DebugLog("applySystemdConf - returns 'true', 'nil'")
		return true, nil
	}

	// Determine for each parameter the latest applied Note that defines it
	// (evaluating effective tuning)
	latestNoteForParam := make(map[string]string)
	latestEntryForParam := make(map[string]txtparser.INIEntry)
	var paramOrder []string

	for _, n := range appliedNotes {
		var effectiveEntries []txtparser.INIEntry
		noteFound, diskEntries := getNoteEffectiveSectionEntries(section, n.NoteID)
		if noteFound {
			effectiveEntries = diskEntries
		} else {
			// Fallback: note file not found on disk (e.g. in unit tests)
			effectiveEntries = n.Entries
		}

		for _, entry := range effectiveEntries {
			cleanKey := cleanSystemdKey(entry.Key)
			if _, exists := latestNoteForParam[cleanKey]; !exists {
				paramOrder = append(paramOrder, cleanKey)
			}
			latestNoteForParam[cleanKey] = n.NoteID
			latestEntryForParam[cleanKey] = entry
		}
	}

	// Group parameters under their winning owner note
	linesPerNote := make(map[string][]string)
	for _, n := range appliedNotes {
		linesPerNote[n.NoteID] = []string{}
	}

	for _, cleanKey := range paramOrder {
		owner := latestNoteForParam[cleanKey]
		entry := latestEntryForParam[cleanKey]
		val := entry.Value
		if entry.Operator == txtparser.OperatorResetAssign {
			linesPerNote[owner] = append(linesPerNote[owner], cleanKey+"=")
			linesPerNote[owner] = append(linesPerNote[owner], cleanKey+"="+val)
		} else {
			linesPerNote[owner] = append(linesPerNote[owner], cleanKey+"="+val)
		}
	}

	var blocks []SystemdNoteBlock
	for _, n := range appliedNotes {
		lines := linesPerNote[n.NoteID]
		if len(lines) > 0 {
			blocks = append(blocks, SystemdNoteBlock{
				NoteID: n.NoteID,
				Lines:  lines,
			})
		}
	}

	if len(blocks) == 0 {
		if err := os.Remove(dropInFile); err != nil && !os.IsNotExist(err) {
			system.ErrorLog("Failed to remove drop-in file '%s': %v", dropInFile, err)
			system.DebugLog("applySystemdConf - returns 'false', err as '%v'", err)
			return false, err
		}
		system.DebugLog("applySystemdConf - returns 'true', 'nil'")
		return true, nil
	}

	if err := os.MkdirAll(dropInDir, 0755); err != nil {
		system.ErrorLog("Failed to create drop-in directory '%s': %v", dropInDir, err)
		system.DebugLog("applySystemdConf - returns 'false', err as '%v'", err)
		return false, err
	}

	rendered := renderDropInContent(blocks)

	if dropInExists {
		system.NoticeLog("Drop-in file '%s' already exists, overwriting", dropInFile)
	}
	if err := os.WriteFile(dropInFile, []byte(rendered), 0644); err != nil {
		system.ErrorLog("Failed to write drop-in file '%s': %v", dropInFile, err)
		system.DebugLog("applySystemdConf - returns 'false', err as '%v'", err)
		return false, err
	}

	system.DebugLog("applySystemdConf - returns 'true', 'nil'")
	return true, nil
}
