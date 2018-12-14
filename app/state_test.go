package app

import (
	"github.com/SUSE/saptune/sap/note"
	"os"
	"path"
	"testing"
)

// sample note implementation 1
type Note1 struct {
	Str string
}

func (n Note1) Name() string {
	return "this is note 1"
}
func (n Note1) AlsoCovers() map[int]string {
	return map[int]string{}
}
func (n Note1) Initialise() (note.Note, error) {
	return nil, nil // not used
}
func (n Note1) Optimise() (note.Note, error) {
	n.Str = "optimised"
	return n, nil
}
func (n Note1) Apply() error {
	return nil // not used
}

// sample note implementation 2
type Note2 struct {
	Int int
}

func (n Note2) Name() string {
	return "this is note 2"
}
func (n Note2) AlsoCovers() map[int]string {
	return map[int]string{}
}
func (n Note2) Initialise() (note.Note, error) {
	return nil, nil // not used
}
func (n Note2) Optimise() (note.Note, error) {
	n.Int = 2
	return n, nil
}
func (n Note2) Apply() error {
	return nil // not used
}

func TestState(t *testing.T) {
	tmpDir := path.Join(os.TempDir(), "saptune-test")
	defer os.RemoveAll(tmpDir)
	state := State{StateDirPrefix: tmpDir}

	note1 := Note1{Str: "initial value"}
	note2 := Note2{Int: 1}

	n1file := state.GetPathToNote("1")
	if n1file != "/tmp/saptune-test/var/lib/saptune/saved_state/1" {
		t.Fatal(n1file)
	}
	n2file := state.GetPathToNote("2")
	if n2file != "/tmp/saptune-test/var/lib/saptune/saved_state/2" {
		t.Fatal(n2file)
	}

	// Store and list
	if num, err := state.List(); len(num) != 0 || err != nil {
		t.Fatal(num, err)
	}
	if err := state.Store("1", note1, true); err != nil {
		t.Fatal(err)
	}
	if err := state.Store("2", note2, true); err != nil {
		t.Fatal(err)
	}
	if err := state.Store("2", note2, false); err != nil {
		t.Fatal(err)
	}
	if num, err := state.List(); err != nil || len(num) != 2 || num[0] != "1" || num[1] != "2" {
		t.Fatal(num, err)
	}

	// Retrieve and compare
	readNote1 := Note1{}
	readNote2 := Note2{}
	if err := state.Retrieve("1", &readNote1); err != nil || readNote1 != note1 {
		t.Fatal(err, readNote1)
	}
	if err := state.Retrieve("2", &readNote2); err != nil || readNote2 != note2 {
		t.Fatal(err, readNote2)
	}
	// Remove
	if err := state.Remove("1"); err != nil {
		t.Fatal(err)
	}
	if err := state.Remove("1"); err != nil { // remove again should not raise error
		t.Fatal(err)
	}
	if num, err := state.List(); err != nil || len(num) != 1 || num[0] != "2" {
		t.Fatal(num, err)
	}
	if err := state.Remove("2"); err != nil {
		t.Fatal(err)
	}
	if num, err := state.List(); len(num) != 0 || err != nil {
		t.Fatal(num, err)
	}
}
