package app

import (
	"encoding/json"
	"github.com/HouzuoGuo/saptune/sap/note"
	"io/ioutil"
	"os"
	"path"
)

const SaptuneStateDir = "/var/lib/saptune/saved_state"

// Store and manage serialised note states.
type State struct {
	StateDirPrefix string
}

// Return path to the serialised note state file.
func (state *State) GetPathToNote(noteID string) string {
	return path.Join(state.StateDirPrefix, SaptuneStateDir, noteID)
}

// Create a file under state directory with the object serialised into JSON. Overwrite existing file if there is any.
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

// Deserialise an SAP note into the destination pointer. The destination must be a pointer.
func (state *State) Retrieve(noteID string, dest interface{}) (err error) {
	content, err := ioutil.ReadFile(state.GetPathToNote(noteID))
	if err != nil {
		return
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
