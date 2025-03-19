package app

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"strings"
	"testing"
)

func TestGetSortedSolutionNotes(t *testing.T) {
	tuneApp := InitialiseApp(OSPackageInGOPATH, "", AllTestNotes, AllTestSolutions)
	tuneApp.TuneForSolutions = []string{"sol1"}
	if sols := tuneApp.GetSortedSolutionEnabledNotes(); len(sols) != 1 {
		t.Fatal(sols)
	}
	tuneApp.TuneForSolutions = []string{"sol2"}
	if sols := tuneApp.GetSortedSolutionEnabledNotes(); len(sols) != 1 {
		t.Fatal(sols)
	}
	tuneApp.TuneForSolutions = []string{"sol1", "sol2"}
	if sols := tuneApp.GetSortedSolutionEnabledNotes(); len(sols) != 2 {
		t.Fatal(sols)
	}
	tuneApp.TuneForSolutions = []string{"sol1", "sol2", "sol12"}
	if sols := tuneApp.GetSortedSolutionEnabledNotes(); len(sols) != 2 {
		t.Fatal(sols, len(sols))
	}
}

func TestOptimiseNoteOnly(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	VerifyConfig(t, tuneApp, []string{}, []string{})
	// Optimise note1, then revert it
	if err := tuneApp.TuneNote("1001"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1", "1")
	if err := tuneApp.RevertNote("1001", true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "2")
	// Optimise note2 and revert it
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised2", "3")
	if err := tuneApp.RevertNote("1002", true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "4")

	// Optimise note2, then note1, then note1 again, and then note2 again, and finally revert both (all)
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	if err := tuneApp.TuneNote("1001"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001", "1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1", "5")
	if err := tuneApp.TuneNote("1001"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001", "1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1", "6")
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001", "1002"}, []string{})
	// change expected value from "optimised1" back to "optimised2"
	// we do no longer apply a note again, which was already applied before
	// but the check was moved to main.go (NoteAction) to suppress
	// misleading messages for the customer
	// so function 'TuneNote' will work as before.
	VerifyFileContent(t, SampleParamFile, "optimised2", "7")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "8")
	// Try optimising for non-existing notes
	if err := tuneApp.TuneNote("8932147"); err == nil {
		t.Fatal("did not error")
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
}

func TestGetNoteByID(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)

	// check for existing Note
	if _, err := tuneApp.GetNoteByID("1001"); err != nil {
		t.Errorf("Note ID '1001' not found, but should be available. AllNote is '%+v', err is '%+v'\n", tuneApp.AllNotes, err)
	}
	// check for non-existing Note
	if _, err := tuneApp.GetNoteByID("8932147"); err == nil {
		t.Errorf("Note ID '8932147' should NOT be available, but is reported as available. AllNote is '%+v'\n", tuneApp.AllNotes)
	}
}

func TestNoteSanityCheck(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)

	sectPath := "/run/saptune/sections"
	sectFile := "/run/saptune/sections/8932147.sections"
	// copy empty file
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/saptune_NOEXIT")
	dest := tuneApp.State.GetPathToNote("8932147")
	os.MkdirAll(path.Join(SampleNoteDataDir, "/data/run/saptune/saved_state"), 0755)
	err := system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dest)
	os.MkdirAll(sectPath, 0755)
	err = system.CopyFile(src, sectFile)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(sectFile)

	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Error(err)
	}
	if err := tuneApp.TuneNote("1001"); err != nil {
		t.Error(err)
	}
	if err := tuneApp.NoteSanityCheck(); err != nil {
		t.Errorf("Error during NoteSanityCheck - '%v'\n", err)
	}

	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "8932147")
	if err := tuneApp.NoteSanityCheck(); err != nil {
		t.Errorf("Error during NoteSanityCheck - '%v'\n", err)
	}

	os.Remove(dest)
	os.Remove(sectFile)
	// existing, but empty section file
	err = system.CopyFile(src, sectFile)
	if err != nil {
		t.Error(err)
	}
	// copy NON empty file
	src = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name")
	err = system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "8932147")
	if err := tuneApp.NoteSanityCheck(); err != nil {
		t.Errorf("Error during NoteSanityCheck - '%v'\n", err)
	}

	os.Remove(dest)
	os.Remove(sectFile)
	// no section file
	src = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name")
	err = system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "8932147")
	if err := tuneApp.NoteSanityCheck(); err != nil {
		t.Errorf("Error during NoteSanityCheck - '%v'\n", err)
	}

	os.Remove(dest)
	os.Remove(sectFile)
	// section file, but no state file
	src = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/saptune_NOEXIT")
	err = system.CopyFile(src, sectFile)
	if err != nil {
		t.Error(err)
	}
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "8932147")
	if err := tuneApp.NoteSanityCheck(); err != nil {
		t.Errorf("Error during NoteSanityCheck - '%v'\n", err)
	}

	tuneApp = InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
}

func TestAppliedNotes(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1001")
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1002")

	expNotes := ""
	applNotes := tuneApp.AppliedNotes()
	if expNotes != applNotes {
		t.Errorf("got: %+v, expected: %+v\n", applNotes, expNotes)
	}

	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/saptune_NOEXIT")
	dest := tuneApp.State.GetPathToNote("1001")
	os.MkdirAll(path.Join(SampleNoteDataDir, "/data/run/saptune/saved_state"), 0755)
	err := system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dest)

	expNotes = "1001"
	applNotes = tuneApp.AppliedNotes()
	if expNotes != applNotes {
		t.Errorf("got: %+v, expected: %+v\n", applNotes, expNotes)
	}
}

func TestIsNoteApplied(t *testing.T) {
	// test fix for bsc#1167618
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1001")
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1002")

	// copy empty file
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/saptune_NOEXIT")
	dest := tuneApp.State.GetPathToNote("8932147")
	os.MkdirAll(path.Join(SampleNoteDataDir, "/data/run/saptune/saved_state"), 0755)
	err := system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dest)

	expmmatch := ""
	mmatch, applied := tuneApp.IsNoteApplied("8932147")
	if applied {
		t.Errorf("expected 'false' but got '%+v'\n", applied)
	}
	if expmmatch != mmatch {
		t.Errorf("got: %+v, expected: %+v\n", mmatch, expmmatch)
	}

	os.Remove(dest)
	// copy NON empty file
	src = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name")
	err = system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}

	expmmatch = "mismatch"
	mmatch, applied = tuneApp.IsNoteApplied("8932147")
	if !applied {
		t.Errorf("expected 'false' but got '%+v'\n", applied)
	}
	if expmmatch != mmatch {
		t.Errorf("got: %+v, expected: %+v\n", mmatch, expmmatch)
	}
}

func TestGetSortedAllNotes(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	expNotes := "1001 1002"
	allNotes := strings.Join(tuneApp.GetSortedAllNotes(), " ")
	if expNotes != allNotes {
		t.Errorf("got: %+v, expected: %+v\n", allNotes, expNotes)
	}
}
