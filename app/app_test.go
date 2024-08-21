package app

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

var OSPackageInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/")
var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/")
var SampleNoteDataDir = "/tmp/saptunetest"
var SampleParamFile = path.Join(SampleNoteDataDir, "saptune-sample-param")

// Sample note and parameter implementations
type SampleParam struct {
	Data string
}

func (pa SampleParam) Name() string {
	return "Sample parameter"
}
func (pa SampleParam) Inspect() (param.Parameter, error) {
	content, _ := os.ReadFile(SampleParamFile)
	pa.Data = string(content)
	return pa, nil
}
func (pa SampleParam) Optimise(way interface{}) (param.Parameter, error) {
	pa.Data = "optimised" + fmt.Sprint(way)
	return pa, nil
}
func (pa SampleParam) Apply(way interface{}) error {
	return os.WriteFile(SampleParamFile, []byte(pa.Data), 0644)
}

type SampleNote1 struct {
	Param SampleParam
}

func (n1 SampleNote1) Name() string {
	return "sample note 1"
}
func (n1 SampleNote1) Initialise() (note.Note, error) {
	newParam, err := n1.Param.Inspect()
	n1.Param = newParam.(SampleParam)
	return n1, err
}
func (n1 SampleNote1) Optimise() (note.Note, error) {
	newParam, err := n1.Param.Optimise("1")
	n1.Param = newParam.(SampleParam)
	return n1, err
}
func (n1 SampleNote1) Apply() error {
	return n1.Param.Apply("1")
}

type SampleNote2 struct {
	Param SampleParam
}

func (n2 SampleNote2) Name() string {
	return "sample note 2"
}
func (n2 SampleNote2) Initialise() (note.Note, error) {
	newParam, err := n2.Param.Inspect()
	n2.Param = newParam.(SampleParam)
	return n2, err
}
func (n2 SampleNote2) Optimise() (note.Note, error) {
	newParam, err := n2.Param.Optimise("2")
	n2.Param = newParam.(SampleParam)
	return n2, err
}
func (n2 SampleNote2) Apply() error {
	return n2.Param.Apply("2")
}

var AllTestNotes = map[string]note.Note{"1001": SampleNote1{}, "1002": SampleNote2{}}
var AllTestSolutions = map[string]solution.Solution{
	"sol1":  {"1001"},
	"sol2":  {"1002"},
	"sol12": {"1001", "1002"},
}

// Make sure that the app struct and its configuration file both have the same enabled notes and enabled solutions.
func VerifyConfig(t *testing.T, app *App, hasNotes []string, hasSolutions []string) {
	if !reflect.DeepEqual(app.TuneForNotes, hasNotes) {
		t.Errorf("Notes diff %v %v", hasNotes, app.TuneForNotes)
	}
	if !reflect.DeepEqual(app.TuneForSolutions, hasSolutions) {
		t.Errorf("Solutions diff %v %v", hasSolutions, app.TuneForSolutions)
	}
	appReloaded := InitialiseApp(app.SysconfigPrefix, app.State.StateDirPrefix, AllTestNotes, AllTestSolutions)
	if !reflect.DeepEqual(app.TuneForNotes, appReloaded.TuneForNotes) {
		t.Errorf("Notes diff %v %v", appReloaded.TuneForNotes, app.TuneForNotes)
	}
	if !reflect.DeepEqual(app.TuneForSolutions, appReloaded.TuneForSolutions) {
		t.Errorf("Solutions diff %v %v", appReloaded.TuneForNotes, app.TuneForSolutions)
	}
}

func WriteFileOrPanic(filePath, content string) {
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		panic(err)
	}
}

// Verify that the file content is exactly as specified.
func VerifyFileContent(t *testing.T, filePath, content, no string) {
	if fileContent, err := os.ReadFile(filePath); err != nil {
		t.Fatal(err)
	} else if string(fileContent) != content {
		t.Errorf("%s - file content mismatch\nexpected:%s\nactual:%s", no, content, string(fileContent))
	}
}

func TestReadConfig(t *testing.T) {
	// Read the default config should not yield anything
	tuneApp := InitialiseApp(OSPackageInGOPATH, "", AllTestNotes, AllTestSolutions)
	if len(tuneApp.TuneForSolutions) != 0 || len(tuneApp.TuneForNotes) != 0 {
		fmt.Println(len(tuneApp.TuneForSolutions))
		fmt.Println(len(tuneApp.TuneForNotes))
		t.Fatal(tuneApp)
	}
	// Read from non existing file
	tuneApp = InitialiseApp("/tmp/saptune", "", AllTestNotes, AllTestSolutions)
	if len(tuneApp.TuneForSolutions) != 0 || len(tuneApp.TuneForNotes) != 0 {
		fmt.Println(len(tuneApp.TuneForSolutions))
		fmt.Println(len(tuneApp.TuneForNotes))
		t.Fatal(tuneApp)
	}

	time.Sleep(5 * time.Second)
	// Read from testdata config 'testdata/etc/sysconfig/saptune'
	_ = system.CopyFile(path.Join(TstFilesInGOPATH, "etc/sysconfig/saptune_tstorg"), path.Join(TstFilesInGOPATH, "etc/sysconfig/saptune"))
	tApp := InitialiseApp(TstFilesInGOPATH, "", AllTestNotes, AllTestSolutions)
	matchTxt := `
current order of enabled notes is: 2205917 2684254 1680803

`
	buffer := bytes.Buffer{}
	tApp.PrintNoteApplyOrder(&buffer)
	txt := buffer.String()
	if txt != matchTxt {
		fmt.Println("==============")
		fmt.Println(txt)
		fmt.Println("==============")
		fmt.Println(matchTxt)
		fmt.Println("==============")
		t.Errorf("Output differs from expected one")
	}
}

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

func TestCombiningSolutionAndNotes(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	VerifyConfig(t, tuneApp, []string{}, []string{})
	// Optimise sol1, note2, revert note2, add note2, and then add sol12
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "29")
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1002"}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "30")
	if err := tuneApp.RevertNote("1002", true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "31")
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1002"}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "32")
	if removedNotes, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	} else if len(removedNotes) != 1 && removedNotes[0] != "1002" {
		t.Fatal(removedNotes)
	}
	// note2 should be removed from list
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised2", "33")
	// Revert all
	if err := tuneApp.RevertAll(false); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised1", "34")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	// Note1 memorises "", note2 memorises "optimised1"
	VerifyFileContent(t, SampleParamFile, "optimised1", "35")
}

func TestVerifyNoteAndSolutions(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	VerifyConfig(t, tuneApp, []string{}, []string{})

	// Tune for sol1 and "1002", so that system will conform to "1002" but not sol1.
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	if notes, comparisons, err := tuneApp.VerifySolution("sol1"); err != nil || len(notes) != 1 || len(comparisons) != 1 || notes[0] != "1001" {
		t.Fatal(notes, comparisons, err)
	}
	if conforming, comparisons, _, err := tuneApp.VerifyNote("1002"); err != nil || len(comparisons) == 0 || !conforming {
		t.Fatal(conforming, comparisons, err)
	}
	// neither sol1 nor "1001" is conformed
	if conforming, comparisons, _, err := tuneApp.VerifyNote("1001"); err != nil || len(comparisons) == 0 || conforming {
		t.Fatal(conforming, comparisons, err)
	}
	if notes, comparisons, err := tuneApp.VerifySolution("sol12"); err != nil || len(notes) != 1 || len(comparisons) != 2 || notes[0] != "1001" {
		t.Fatal(notes, comparisons, err)
	}
	if notes, comparisons, err := tuneApp.VerifyAll(); err != nil || len(notes) != 1 || len(comparisons) != 2 || notes[0] != "1001" {
		t.Fatal(notes, comparisons, err)
	}
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

func TestTuneAll(t *testing.T) {
	os.RemoveAll(SampleNoteDataDir)
	defer os.RemoveAll(SampleNoteDataDir)
	tuneApp := InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1001")
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "1002")
	if err := tuneApp.TuneAll(); err != nil {
		t.Errorf("Error during TuneAll - '%v'\n", err)
	}
	tuneApp = InitialiseApp(path.Join(SampleNoteDataDir, "conf"), path.Join(SampleNoteDataDir, "data"), AllTestNotes, AllTestSolutions)
	tuneApp.NoteApplyOrder = append(tuneApp.NoteApplyOrder, "8932147")
	if err := tuneApp.TuneAll(); err != nil {
		t.Errorf("Error during TuneAll - '%v'\n", err)
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

func TestInitialiseApp(t *testing.T) {
	tstApp := InitialiseApp("/sys/", "", AllTestNotes, AllTestSolutions)
	if len(tstApp.TuneForSolutions) != 0 && len(tstApp.TuneForNotes) != 0 && len(tstApp.NoteApplyOrder) != 0 {
		fmt.Println(len(tstApp.TuneForSolutions), tstApp.TuneForSolutions)
		fmt.Println(len(tstApp.TuneForNotes), tstApp.TuneForSolutions)
		fmt.Println(len(tstApp.NoteApplyOrder), tstApp.TuneForSolutions)
		t.Error(tstApp)
	}
}
