package app

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/sap/note"
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/sap/solution"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var OSPackageInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/HouzuoGuo/saptune/ospackage/")
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
	content, _ := ioutil.ReadFile(SampleParamFile)
	pa.Data = string(content)
	return pa, nil
}
func (pa SampleParam) Optimise(way interface{}) (param.Parameter, error) {
	pa.Data = "optimised" + fmt.Sprint(way)
	return pa, nil
}
func (pa SampleParam) Apply() error {
	return ioutil.WriteFile(SampleParamFile, []byte(pa.Data), 0644)
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
	return n1.Param.Apply()
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
	return n2.Param.Apply()
}

var AllTestNotes = map[string]note.Note{"1001": SampleNote1{}, "1002": SampleNote2{}}
var AllTestSolutions = map[string]solution.Solution{
	"sol1":  solution.Solution{"1001"},
	"sol2":  solution.Solution{"1002"},
	"sol12": solution.Solution{"1001", "1002"},
}

// Make sure that the app struct and its configuration file both have the same enabled notes and enabled solutions.
func VerifyConfig(t *testing.T, app *App, hasNotes []string, hasSolutions []string) {
	if !reflect.DeepEqual(app.TuneForNotes, hasNotes) {
		panic(fmt.Sprintf("Notes diff %v %v", hasNotes, app.TuneForNotes))
	}
	if !reflect.DeepEqual(app.TuneForSolutions, hasSolutions) {
		panic(fmt.Sprintf("Solutions diff %v %v", hasSolutions, app.TuneForSolutions))
	}
	appReloaded := InitialiseApp(app.SysconfigPrefix, app.State.StateDirPrefix, AllTestNotes, AllTestSolutions)
	if !reflect.DeepEqual(app.TuneForNotes, appReloaded.TuneForNotes) {
		panic(fmt.Sprintf("Notes diff %v %v", appReloaded.TuneForNotes, app.TuneForNotes))
	}
	if !reflect.DeepEqual(app.TuneForSolutions, appReloaded.TuneForSolutions) {
		panic(fmt.Sprintf("Solutions diff %v %v", appReloaded.TuneForNotes, app.TuneForSolutions))
	}
}

func WriteFileOrPanic(filePath, content string) {
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		panic(err)
	}
}

// Verify that the file content is exactly as specified.
func VerifyFileContent(t *testing.T, filePath, content string) {
	if fileContent, err := ioutil.ReadFile(filePath); err != nil {
		t.Fatal(err)
	} else if string(fileContent) != content {
		panic(fmt.Sprintf("file content mismatch\nexpected:%s\nactual:%s", content, string(fileContent)))
	}
}

func TestReadConfig(t *testing.T) {
	// Read the default config should not yield anything
	tuneApp := InitialiseApp(OSPackageInGOPATH, "", AllTestNotes, AllTestSolutions)
	if len(tuneApp.TuneForSolutions) != 0 || len(tuneApp.TuneForNotes) != 0 {
		fmt.Printf("'%v'", tuneApp.TuneForSolutions[0])
		fmt.Println(len(tuneApp.TuneForNotes))
		t.Fatal(tuneApp)
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
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.RevertNote("1001", true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")
	// Optimise note2 and revert it
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertNote("1002", true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")

	// Optimise note2, then note1, then note1 again, and then note2 again, and finally revert both (all)
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	if err := tuneApp.TuneNote("1001"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001", "1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.TuneNote("1001"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001", "1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1001", "1002"}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")
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
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.RevertSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")
	// Optimise sol2 and revert it
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")

	// Optimise sol2, then sol1, then sol1 again, and then sol1 again, and finally revert both (all)
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")

	// Optimise sol12, then sol1, and then revert sol12, and then revert sol1
	if _, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.RevertSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.RevertSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "")

	// Optimise sol1, sol2, sol12, and then sol2 and sol1 again, eventually revert all
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if _, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	// Note "1001" wants to restore the file to empty, while note "1002" wants to restore it to "optimised1"
	VerifyConfig(t, tuneApp, []string{}, []string{})
	VerifyFileContent(t, SampleParamFile, "optimised1")

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

	// Optimise sol1, sol2, sol12, and then revert sol12
	if _, err := tuneApp.TuneSolution("sol2"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if _, err := tuneApp.TuneSolution("sol1"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if _, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12", "sol2"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertSolution("sol12"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol2"})
	// Reverting sol12 should not affect anything
	VerifyFileContent(t, SampleParamFile, "optimised2")
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
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1002"}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if err := tuneApp.RevertNote("1002", true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.TuneNote("1002"); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{"1002"}, []string{"sol1"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	if removedNotes, err := tuneApp.TuneSolution("sol12"); err != nil {
		t.Fatal(err)
	} else if len(removedNotes) != 1 && removedNotes[0] != "1002" {
		t.Fatal(removedNotes)
	}
	// note2 should be removed from list
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised2")
	// Revert all
	if err := tuneApp.RevertAll(false); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{"sol1", "sol12"})
	VerifyFileContent(t, SampleParamFile, "optimised1")
	if err := tuneApp.RevertAll(true); err != nil {
		t.Fatal(err)
	}
	VerifyConfig(t, tuneApp, []string{}, []string{})
	// Note1 memorises "", note2 memorises "optimised1"
	VerifyFileContent(t, SampleParamFile, "optimised1")
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
	if conforming, comparisons, err := tuneApp.VerifyNote("1002"); err != nil || len(comparisons) == 0 || !conforming {
		t.Fatal(conforming, comparisons, err)
	}
	// neither sol1 nor "1001" is conformed
	if conforming, comparisons, err := tuneApp.VerifyNote("1001"); err != nil || len(comparisons) == 0 || conforming {
		t.Fatal(conforming, comparisons, err)
	}
	if notes, comparisons, err := tuneApp.VerifySolution("sol12"); err != nil || len(notes) != 1 || len(comparisons) != 2 || notes[0] != "1001" {
		t.Fatal(notes, comparisons, err)
	}
	if notes, comparisons, err := tuneApp.VerifyAll(); err != nil || len(notes) != 1 || len(comparisons) != 2 || notes[0] != "1001" {
		t.Fatal(notes, comparisons, err)
	}
}
