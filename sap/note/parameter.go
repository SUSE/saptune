/*
SAP notes tune one or more parameters at a time.
*/
package note

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type ParameterNoteEntry struct {
	NoteId   string
	Value    string
}

type ParameterNotes struct {
	AllNotes []ParameterNoteEntry	// list of applied notes, which manipulate the given system parameter
}

// Directory where to store the parameter state files
// separated from the note state file directory
const SaptuneParameterStateDir = "/var/lib/saptune/parameter"

// Return path to the serialised parameter state file.
func GetPathToParameter(param string) string {
        return path.Join(SaptuneParameterStateDir, param)
}

// Check, if given noteID is already part of the parameter list of notes
func NoteInParameterList(noteID string, list []ParameterNoteEntry) bool {
	for _, note := range list {
		if note.NoteId == noteID {
			return true
		}
	}
	return false
}

// List all stored paramter states. Return paramter names
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

// Create parameter state file and insert the start values.
func CreateParameterStartValues(param, value string) {
	pEntries := ParameterNotes {
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	pEntries = GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) == 0 {
		//log.Printf("Write parameter start value '%s' to file '%s'", value, GetPathToParameter(param))
		// file does not exist, create start entry
		pEntry := ParameterNoteEntry{
			NoteId:   "start",
			Value:    value,
		}
		pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
		err := StoreParameter(param, pEntries, true)
		if err != nil {
			log.Printf("Failed to store start values for parameter file '%s' for parameter '%s'", GetPathToParameter(param), param)
		}
	}
}

// Add note parameter values to the state file.
func AddParameterNoteValues(param, value, noteID string) {
	pEntries := ParameterNotes {
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	pEntries = GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) != 0 && !NoteInParameterList(noteID, pEntries.AllNotes) {
		//log.Printf("Write note '%s' parameter value '%s' to file '%s'", noteID, value, GetPathToParameter(param))
		// file exis
		pEntry := ParameterNoteEntry{
			NoteId:   noteID,
			Value:    value,
		}
		pEntries.AllNotes = append(pEntries.AllNotes, pEntry)
		err := StoreParameter(param, pEntries, true)
		if err != nil {
			log.Printf("Failed to store note '%s' values for parameter file '%s' for parameter '%s'", noteID, GetPathToParameter(param), param)
		}
	}
}

// Read content of stored paramter states. Return the content as ParameterNotes
func GetSavedParameterNotes(param string) (ParameterNotes) {
	pEntries := ParameterNotes {
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	content, err := ioutil.ReadFile(GetPathToParameter(param))
	if err != nil {
		return pEntries
	}
	err = json.Unmarshal(content, &pEntries)
	return pEntries
}

// read all saved parameters from the state directory
// fill structure app.AllParameters
func GetAllSavedParameters() (map[string]ParameterNotes) {
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

// Store parameter values to state directory
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

// for a given noteID get the position in the slice AllNotes
// return the position of the note within the slice.
// do not sort the slice
func PositionInParameterList(noteID string, list []ParameterNoteEntry) int {
	for cnt, note := range list {
		if note.NoteId == noteID {
			return cnt
		}
	}
	return 0
}

// Revert parameter value and remove noteID reference from parameter file
// return value of parameter, if needed to change, empty string else
func RevertParameter(param string, noteID string) (string) {
	pvalue := ""
	pEntries := ParameterNotes {
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	// read values from the parameter state file
	pEntries = GetSavedParameterNotes(param)
	if len(pEntries.AllNotes) == 0 {
		return pvalue
	}
	lastNote := pEntries.AllNotes[len(pEntries.AllNotes)-1]
	pvalue = lastNote.Value
	if lastNote.NoteId == noteID {
		// if the requested noteID is the last one in AllNotes
		// remove this entry and set the parameter value to the 'next to last'
		pEntries.AllNotes = pEntries.AllNotes[:len(pEntries.AllNotes)-1]
		next2lastNote := pEntries.AllNotes[len(pEntries.AllNotes)-1]
		pvalue = next2lastNote.Value
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
		remFileName := GetPathToParameter(param)
		if _, err := os.Stat(remFileName); err == nil {
			os.Remove(remFileName)
		}
	} else {
		//store changes pEntries
		err := StoreParameter(param, pEntries, true)
		if err != nil {
			fmt.Println("Problems during storing new parameter values")
		}
	}
	return pvalue
}

/*
For just reading the last element of a slice:

sl[len(sl)-1]

For removing it:

sl = sl[:len(sl)-1]
*/
