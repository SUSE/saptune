package app

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"testing"
)

func TestOptimiseSolutionOnly(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	VerifyConfig(t, tuneApp, []string{}, []string{})
	// Optimise sol1, then revert it
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "9")
	if err := tuneApp.RevertSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "10")
	// Optimise sol2 and revert it
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "11")
	if err := tuneApp.RevertSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "12")

	// Optimise sol2, then sol1, then sol1 again, and then sol1 again, and finally revert both (all)
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "13")
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "14")
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	// change expected value from "optimised2" to "optimised1", as we do no
	// longer apply a note again, which was already applied before.
	VerifyFileContent(t, SampleParamFile, "optimised1", "15")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "16")

	// Optimise sol12, then sol1, and then revert sol12, and then revert sol1
	if _, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "17")
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12"})
	// change expected value from "optimised1" to "optimised2", as we do no
	// longer apply a note again, which was already applied before.
	VerifyFileContent(t, SampleParamFile, "optimised2", "18")
	if err := tuneApp.RevertSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "19")
	if err := tuneApp.RevertSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "", "20")

	// Optimise sol1, sol2, sol12, and then sol2 and sol1 again, eventually revert all
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "21")
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "22")
	if _, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "23")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	// Note "1001" wants to restore the file to empty, while note "1002" wants to restore it to "optimised1"
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1", "24")

	// Try optimising for non-existing solution
	if _, err := tuneApp.TuneSolution("this one does not exist"); err == nil {
		t.Fatal("did not error")
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
}

func TestOverlappingSolutions(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	VerifyConfig(t, tuneApp, []string{}, []string{})

	// Optimise sol2, sol1, sol12, and then revert sol12
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "25")
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "26")
	if _, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12", "sol2"})
	// change expected value from "optimised2" to "optimised1", as we do no
	// longer apply a note again, which was already applied before.
	VerifyFileContent(t, SampleParamFile, "optimised1", "27")
	if err := tuneApp.RevertSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	// Reverting sol12 should not affect anything
	// change expected value from "optimised2" to "optimised1", as we do no
	// longer apply a note again, which was already applied before.
	VerifyFileContent(t, SampleParamFile, "optimised1", "28")
}

func TestIsSolutionEnabled(t *testing.T) {
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	tuneApp.TuneForSolutions = []string{"sol1"}
	if tuneApp.IsSolutionEnabled("sol15") {
		t.Error("expected 'false' but got 'true'")
	}
	if !tuneApp.IsSolutionEnabled("sol1") {
		t.Error("expected 'true' but got 'false'")
	}
}

func TestAppliedSolution(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1001")
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1002")
	tuneApp.TuneForSolutions = []string{"sol1"}

	expSol := ""
	expState := ""
	applSol, state := tuneApp.AppliedSolution()
	if expSol != applSol {
		t.Errorf("got: %+v, expected: %+v\n", applSol, expSol)
	}
	if expState != state {
		t.Errorf("got: %+v, expected: %+v\n", state, expState)
	}

	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/saptune_NOEXIT")
	dest := tuneApp.State.GetPathToNote("1001")
	os.MkdirAll(path.Join(SampleNoteDataDir, "/data/run/saptune/saved_state"), 0755)
	err := system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dest)

	expSol = "sol1"
	expState = "fully"
	applSol, state = tuneApp.AppliedSolution()
	if expSol != applSol {
		t.Errorf("got: %+v, expected: %+v\n", applSol, expSol)
	}
	if expState != state {
		t.Errorf("got: %+v, expected: %+v\n", state, expState)
	}

	tuneApp.TuneForSolutions = []string{"sol12"}
	expSol = "sol12"
	expState = "partial"
	applSol, state = tuneApp.AppliedSolution()
	if expSol != applSol {
		t.Errorf("got: %+v, expected: %+v\n", applSol, expSol)
	}
	if expState != state {
		t.Errorf("got: %+v, expected: %+v\n", state, expState)
	}
}
