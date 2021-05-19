package note

import (
	"encoding/json"
	"github.com/SUSE/saptune/system"
	"io/ioutil"
	"os"
	"path"
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
	return path.Join(SaptuneParameterStateDir, param)
}

// IDInParameterList checks, if given noteID is already part of the
// parameter list of notes
func IDInParameterList(noteID string, list []ParameterNoteEntry) bool {
	for _, note := range list {
		if note.NoteID == noteID {
			return true
		}
	}
	return false
}

// ListParams lists all stored parameter states. Return parameter names
func ListParams() (ret []string, err error) {
	if err = os.MkdirAll(SaptuneParameterStateDir, 0755); err != nil {
		return
	}
	// List SaptuneParameterStateDir and collect parameter names from file names
	dirContent, err := ioutil.ReadDir(SaptuneParameterStateDir)
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
	pEntries := ParameterNotes{
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	pEntries = GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) == 0 {
		//system.InfoLog("Write parameter start value '%s' to file '%s'", value, GetPathToParameter(param))
		// file does not exist, create start entry
		pEntry := ParameterNoteEntry{
			NoteID: "start",
			Value:  value,
		}
		pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
		err := StoreParameter(param, pEntries, true)
		if err != nil {
			system.WarningLog("Failed to store start values for parameter file '%s' for parameter '%s'", GetPathToParameter(param), param)
		}
	}
}

// AddParameterNoteValues adds note parameter values to the state file.
func AddParameterNoteValues(param, value, noteID string) {
	pEntries := ParameterNotes{
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	pEntries = GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) != 0 && !IDInParameterList(noteID, pEntries.AllNotes) {
		//system.InfoLog("Write note '%s' parameter value '%s' to file '%s'", noteID, value, GetPathToParameter(param))
		// file exis
		pEntry := ParameterNoteEntry{
			NoteID: noteID,
			Value:  value,
		}
		pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
		err := StoreParameter(param, pEntries, true)
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
	content, err := ioutil.ReadFile(GetPathToParameter(param))
	if err != nil {
		return pEntries
	}
	if len(content) != 0 {
		err = json.Unmarshal(content, &pEntries)
	}
	return pEntries
}

// GetAllSavedParameters reads all saved parameters from the state directory
// fill structure app.AllParameters
func GetAllSavedParameters() map[string]ParameterNotes {
	params := make(map[string]ParameterNotes)
	allParams, err := ListParams()
	if err == nil {
		for _, param := range allParams {
			pEntries := GetSavedParameterNotes(param)
			if len(pEntries.AllNotes) == 0 {
				return params
			}
			params[param] = pEntries
		}
	}
	return params
}

// StoreParameter stores parameter values to state directory
// Write a json file with the name of the given parameter containing the
// applied noteIDs for this parameter and the associated parameter values
func StoreParameter(param string, obj ParameterNotes, overwriteExisting bool) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(SaptuneParameterStateDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(GetPathToParameter(param)); os.IsNotExist(err) || overwriteExisting {
		return ioutil.WriteFile(GetPathToParameter(param), content, 0644)
	}
	return nil
}

// PositionInParameterList gets the position in the slice AllNotes for a
// given noteID
// return the position of the note within the slice.
// do not sort the slice
func PositionInParameterList(noteID string, list []ParameterNoteEntry) int {
	for cnt, note := range list {
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
	pEntries := ParameterNotes{
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	// read values from the parameter state file
	pEntries = GetSavedParameterNotes(param)
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
		entry := PositionInParameterList(noteID, pEntries.AllNotes)
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
		err := StoreParameter(param, pEntries, true)
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
