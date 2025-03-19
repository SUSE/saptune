package note

import (
	"encoding/json"
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"strconv"
)

// ParameterNoteEntry stores the parameter values set by a Note
type ParameterNoteEntry struct {
	NoteID string
	Value  string
}

// ParameterNotes includes a list of applied notes, which manipulate the
// given system parameter
// the entries are stored in exactly the order the Notes were applied.
// So you get a (timeline) chain of applied parameter values
type ParameterNotes struct {
	AllNotes []ParameterNoteEntry
}

// GetPathToParameter returns path to the serialised parameter state file.
func GetPathToParameter(param string) string {
	return path.Join(system.SaptuneParameterStateDir, param)
}

// ListParams lists all stored parameter states. Return parameter names
func ListParams() (ret []string, err error) {
	if err = os.MkdirAll(system.SaptuneParameterStateDir, 0755); err != nil {
		return
	}
	// List SaptuneParameterStateDir and collect parameter names from file names
	dirContent, err := os.ReadDir(system.SaptuneParameterStateDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return
	}
	ret = make([]string, 0, len(dirContent))
	for _, pname := range dirContent {
		ret = append(ret, pname.Name())
	}
	return
}

// CreateParameterStartValues creates the parameter state file and inserts
// the start values.
func CreateParameterStartValues(param, value string) {
	pEntries := GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) == 0 {
		system.DebugLog("Write parameter start value '%s' to file '%s'", value, GetPathToParameter(param))
		// file does not exist, create start entry
		pEntry := ParameterNoteEntry{
			NoteID: "start",
			Value:  value,
		}
		pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
		err := pEntries.StoreParameter(param, true)
		if err != nil {
			system.WarningLog("Failed to store start values for parameter file '%s' for parameter '%s'", GetPathToParameter(param), param)
		}
	}
}

// AddParameterNoteValues adds note parameter values to the state file.
func AddParameterNoteValues(param, value, noteID, action string) {
	pEntries := GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) != 0 {
		// file exist
		system.DebugLog("Write note '%s' parameter value '%s' to file '%s'", noteID, value, GetPathToParameter(param))
		pEntry := ParameterNoteEntry{
			NoteID: noteID,
			Value:  value,
		}
		if !pEntries.IDInParameterList(noteID) {
			// noteID not yet available in file, add or insert
			if action == "add" {
				// append entry at the end of the file
				// regular apply workflow
				system.DebugLog("append note '%s' parameter value '%s'", noteID, value)
				pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
			} else {
				// workflow needed by 'note refresh'
				// action contains the index position to insert
				// the new entry
				system.DebugLog("insert note '%s' parameter value '%s' at position '%s'", noteID, value, action)
				idx, _ := strconv.Atoi(action)
				// ignore idx <= 0 (index cannot be less than 0
				// and idx == 0 is the 'start' entry)
				if idx >= len(pEntries.AllNotes) {
					// append to the end
					pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
				} else if idx > 0 {
					// allocate space for new element
					//pEntries.AllNotes = append(pEntries.AllNotes, 0)
					// shift elements
					//copy(pEntries.AllNotes[idx+1:], pEntries.AllNotes[idx:]
					pEntries.AllNotes = append(pEntries.AllNotes[:idx+1], pEntries.AllNotes[idx:]...)
					// insert at 'idx' position
					pEntries.AllNotes[idx] = pEntry
				}
			}
		} else {
			// noteID available in file, change value
			// workflow needed by 'note refresh'
			system.DebugLog("change note '%s' parameter value '%s'", noteID, value)
			idx := pEntries.PositionInParameterList(noteID)
			// ignore idx == 0 (note ID not in file, no file or
			// only 'start' in file)
			if idx > 0 {
				pEntries.AllNotes[idx] = pEntry
			}
		}
		err := pEntries.StoreParameter(param, true)
		if err != nil {
			system.WarningLog("Failed to store note '%s' values for parameter file '%s' for parameter '%s'", noteID, GetPathToParameter(param), param)
		}
	}
}

// GetSavedParameterNotes reads content of stored parameter states.
// Return the content as ParameterNotes
func GetSavedParameterNotes(param string) ParameterNotes {
	pEntries := ParameterNotes{
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	content, err := os.ReadFile(GetPathToParameter(param))
	if err != nil {
		return pEntries
	}
	if len(content) != 0 {
		_ = json.Unmarshal(content, &pEntries)
	}
	return pEntries
}

// GetAllSavedParameters reads all saved parameters from the state directory
// fill structure app.AllParameters
func GetAllSavedParameters() map[string]ParameterNotes {
	params := make(map[string]ParameterNotes)
	allParams, err := ListParams()
	if err != nil {
		return params
	}
	for _, param := range allParams {
		pEntries := GetSavedParameterNotes(param)
		if len(pEntries.AllNotes) == 0 {
			continue
		}
		params[param] = pEntries
	}
	return params
}

// StoreParameter stores parameter values to state directory
// Write a json file with the name of the given parameter containing the
// applied noteIDs for this parameter and the associated parameter values
func (pent ParameterNotes) StoreParameter(param string, overwriteExisting bool) error {
	content, err := json.Marshal(pent)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(system.SaptuneParameterStateDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(GetPathToParameter(param)); os.IsNotExist(err) || overwriteExisting {
		return os.WriteFile(GetPathToParameter(param), content, 0644)
	}
	return nil
}

// IDInParameterList checks, if given noteID is already part of the
// parameter list of notes
func (pent ParameterNotes) IDInParameterList(noteID string) bool {
	for _, note := range pent.AllNotes {
		if note.NoteID == noteID {
			return true
		}
	}
	return false
}

// PositionInParameterList gets the position in the slice AllNotes for a
// given noteID
// return the position of the note within the slice.
// do not sort the slice
func (pent ParameterNotes) PositionInParameterList(noteID string) int {
	for cnt, note := range pent.AllNotes {
		if note.NoteID == noteID {
			return cnt
		}
	}
	return 0
}

// RevertParameter reverts parameter values and removes noteID reference
// from the parameter file
// return value of parameter and related noteID
func RevertParameter(param, noteID string) (string, string) {
	pvalue := ""
	pnoteID := ""
	// read values from the parameter state file
	pEntries := GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) == 0 {
		return pvalue, pnoteID
	}
	lastNote := pEntries.AllNotes[len(pEntries.AllNotes)-1]
	pvalue = lastNote.Value
	pnoteID = lastNote.NoteID
	if lastNote.NoteID == noteID {
		// if the requested noteID is the last one in AllNotes
		// remove this entry and set the parameter value to the 'next to last'
		pEntries.AllNotes = pEntries.AllNotes[:len(pEntries.AllNotes)-1]
		next2lastNote := pEntries.AllNotes[len(pEntries.AllNotes)-1]
		pvalue = next2lastNote.Value
		pnoteID = next2lastNote.NoteID
	} else {
		// if the requested noteID is NOT the last one in AllNotes
		// remove this entry but do not set a new parameter value
		//
		// if the requested noteID has no entry in the parameter file
		// because an override file disabled the parameter setting
		// prevent removal of start value by checking 'entry > 0'
		entry := pEntries.PositionInParameterList(noteID)
		if entry > 0 {
			// the requested noteID is NOT the last one in AllNotes
			pEntries.AllNotes = append(pEntries.AllNotes[0:entry], pEntries.AllNotes[entry+1:]...)
		}
	}
	if len(pEntries.AllNotes) == 1 {
		// remove parameter state file, if only one entry ('start') is left.
		CleanUpParamFile(param)
	} else {
		//store changes pEntries
		err := pEntries.StoreParameter(param, true)
		if err != nil {
			system.WarningLog("Problems during storing new parameter values")
		}
	}
	return pvalue, pnoteID
}

// CleanUpParamFile removes the parameter state file
func CleanUpParamFile(param string) {
	remFileName := GetPathToParameter(param)
	if _, err := os.Stat(remFileName); err == nil {
		os.Remove(remFileName)
	}
}

// IsLastNoteOfParameter returns true, if there is no parameter state file
// or false, which means that another note changing the parameter is still
// applied
func IsLastNoteOfParameter(param string) bool {
	chkFileName := GetPathToParameter(param)
	if _, err := os.Stat(chkFileName); os.IsNotExist(err) {
		return true
	}
	return false
}
