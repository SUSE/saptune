package app

import (
	"encoding/json"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// SaptuneStateDir defines saptunes saved state directory
const SaptuneStateDir = "/var/lib/saptune/saved_state"

// State stores and manages serialised note states.
type State struct {
	StateDirPrefix string
}

// GetPathToNote returns path to the serialised note state file.
func (state *State) GetPathToNote(noteID string) string {
	return path.Join(state.StateDirPrefix, SaptuneStateDir, noteID)
}

// Store creates a file under state directory with the object serialised
// into JSON. Overwrite existing file if there is any.
func (state *State) Store(noteID string, obj note.Note, overwriteExisting bool) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(state.StateDirPrefix, SaptuneStateDir), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(state.GetPathToNote(noteID)); os.IsNotExist(err) || overwriteExisting {
		return ioutil.WriteFile(state.GetPathToNote(noteID), content, 0644)
	}
	return nil
}

// List all stored note states. Return note numbers.
func (state *State) List() (ret []string, err error) {
	if err = os.MkdirAll(path.Join(state.StateDirPrefix, SaptuneStateDir), 0755); err != nil {
		return
	}
	// List SaptuneStateDir and collect number from file names
	dirContent, err := ioutil.ReadDir(path.Join(state.StateDirPrefix, SaptuneStateDir))
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return
	}
	ret = make([]string, 0, len(dirContent))
	for _, info := range dirContent {
		ret = append(ret, info.Name())
	}
	return
}

// Retrieve deserialises a SAP note into the destination pointer.
// The destination must be a pointer.
func (state *State) Retrieve(noteID string, dest interface{}) error {
	content, err := ioutil.ReadFile(state.GetPathToNote(noteID))
	if err != nil {
		return err
	}
	return json.Unmarshal(content, dest)
}

// Remove a serialised state file.
func (state *State) Remove(noteID string) error {
	_, err := os.Stat(state.GetPathToNote(noteID))
	if os.IsNotExist(err) {
		return nil
	} else if err == nil {
		return os.Remove(state.GetPathToNote(noteID))
	} else {
		return err
	}
}

// CheckForOldRevertData checks, if there is saved state information in an older,
// no longer supported saptune format available
// return true, if old saved state files are found
func (state *State) CheckForOldRevertData() (oldUpdFiles []string, check bool) {
	check = false
	oldUpdFiles = make([]string, 0, 0)
	if savedNotes, err := state.List(); err == nil {
		for _, entry := range savedNotes {
			fileName := strings.TrimSuffix(entry, "_n2c")
			if entry != fileName {
				// there was a saved state file available during
				// update from version 1 to version 2
				// check, if the saved state file is already
				// available (as a leftover from the migration)
				if _, err := os.Stat(state.GetPathToNote(fileName)); err == nil {
					// both saved state files exists
					// (the saved state file of an applied
					// note and the 'n2c' file from the
					// update path v1 to v2
					oldUpdFiles = append(oldUpdFiles, fileName)
					system.WarningLog("found old saved state file for Note '%s'.", fileName)
					check = true
				} else {
					oldUpdFiles = append(oldUpdFiles, entry)
				}
			}
		}
	}
	return
}
