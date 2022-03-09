package actions

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

var SolutionSheetsInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/sols") + "/"
var ExtraFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra") + "/"
var OverTstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/etc/saptune/override") + "/"
var DeprecFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/deprecated") + "/"
var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/")

var AllTestSolutions = map[string]solution.Solution{
	"sol1":  {"simpleNote"},
	"sol2":  {"extraNote"},
	"sol12": {"simpleNote", "extraNote"},
}

var tuningOpts = note.GetTuningOptions("", ExtraFilesInGOPATH)
var tApp = app.InitialiseApp(TstFilesInGOPATH, "", tuningOpts, AllTestSolutions)

// setup for ErroExit catches
var tstRetErrorExit = -1
var tstosExit = func(val int) {
	tstRetErrorExit = val
}
var tstwriter io.Writer
var tstErrorExitOut = func(str string, out ...interface{}) error {
	fmt.Fprintf(tstwriter, "ERROR: "+str, out...)
	return fmt.Errorf(str+"\n", out...)
}

var checkOut = func(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		fmt.Println("==============")
		fmt.Println(got)
		fmt.Println("==============")
		fmt.Println(want)
		fmt.Println("==============")
		t.Errorf("Output differs from expected one")
	}
}

var setUpSol = func(t *testing.T) {
	t.Helper()
	// prepare deprecated solution, custom solution and override
	noteFiles := TstFilesInGOPATH + "/"
	extraNoteFiles := TstFilesInGOPATH + "/extra/"
	solution.CustomSolutions = solution.GetOtherSolution(ExtraFilesInGOPATH, noteFiles, extraNoteFiles)
	solution.OverrideSolutions = solution.GetOtherSolution(OverTstFilesInGOPATH, noteFiles, "")
	solution.DeprecSolutions = solution.GetOtherSolution(DeprecFilesInGOPATH, "", "")
	//solution.AllSolutions = solution.GetSolutionDefintion(solution.SolutionSheets)
	solution.AllSolutions = solution.GetSolutionDefintion(SolutionSheetsInGOPATH, extraNoteFiles, noteFiles)
}

var tearDownSol = func(t *testing.T) {
	t.Helper()
	solution.CustomSolutions = solution.GetOtherSolution("", "", "")
	solution.OverrideSolutions = solution.GetOtherSolution("", "", "")
	solution.DeprecSolutions = solution.GetOtherSolution("", "", "")
	//solution.AllSolutions = solution.GetSolutionDefintion(solution.SolutionSheets)
	solution.AllSolutions = solution.GetSolutionDefintion(SolutionSheetsInGOPATH, ExtraTuningSheets, NoteTuningSheets)
}

var setUp = func(t *testing.T) {
	t.Helper()
	// setup code
	// clear note settings in test file
	tApp.TuneForSolutions = []string{}
	tApp.TuneForNotes = []string{}
	tApp.NoteApplyOrder = []string{}
	if err := tApp.SaveConfig(); err != nil {
		t.Errorf("could not save saptune config file")
	}
}

var tearDown = func(t *testing.T) {
	t.Helper()
	// tear-down code
	// restore test file content
	tApp.TuneForSolutions = []string{}
	tApp.TuneForNotes = []string{"1680803", "2205917", "2684254"}
	tApp.NoteApplyOrder = []string{}
	tApp.NoteApplyOrder = []string{"2205917", "2684254", "1680803"}
	if err := tApp.SaveConfig(); err != nil {
		t.Errorf("could not save saptune config file")
	}
}

func TestRevertAction(t *testing.T) {
	var revertMatchText = `Reverting all notes and solutions, this may take some time...
Parameters tuned by the notes and solutions have been successfully reverted.
`
	buffer := bytes.Buffer{}
	RevertAction(&buffer, "all", tApp)
	txt := buffer.String()
	checkOut(t, txt, revertMatchText)

	// test for PrintHelpAndExit
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	// this errExitMatchText differs from the 'real' text by the last 2 lines
	// because of test situation, the 'exit 1' in PrintHelpAndExit is not
	// executed (as designed for testing)
	errExitMatchText := fmt.Sprintf(`saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | revertall | enabled | applied ]
  saptune note [ apply | simulate | verify | customise | create | edit | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled | applied ]
  saptune solution [ apply | simulate | verify | customise | create | edit | revert | show | delete ] SolutionName
  saptune solution rename SolutionName newSolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff | analysis | release ]
   saptune staging [ analysis | diff ] [ NoteID... | SolutionID... | all ]
   saptune staging release [--force|--dry-run] [ NoteID... | SolutionID... | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Call external script '/usr/sbin/saptune_check'
  saptune check
Print current saptune status:
  saptune status
Print current saptune version:
  saptune version
Print this message:
  saptune help
Reverting all notes and solutions, this may take some time...
Parameters tuned by the notes and solutions have been successfully reverted.
`)

	buffer.Reset()
	// reset tApp variables, which were deleted by 'revert all'
	tearDown(t)
	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer
	RevertAction(&buffer, "NotAll", tApp)
	txt = buffer.String()
	checkOut(t, txt, errExitMatchText)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}
	// reset tApp variables, which were deleted by 'revert all'
	tearDown(t)
}

func TestGetFileName(t *testing.T) {
	tstRetErrorExit = -1
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer

	// test with existing extra note
	nID := "simpleNote"
	fname, extra := getFileName(nID, "", ExtraFilesInGOPATH)
	chkname := fmt.Sprintf("%s%s.conf", ExtraFilesInGOPATH, nID)
	if fname != chkname {
		t.Errorf("wrong file name: '%s' instead of '%s'\n", fname, chkname)
	}
	if !extra {
		t.Errorf("note is an extra note and not an internal one\n")
	}
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	// initialise next test
	errExitbuffer.Reset()
	tstRetErrorExit = -1
	// test with non-existing extra note
	nID = "hugo"
	getFnameMatchText := fmt.Sprintf("ERROR: Note %s not found in %s or %s.\n", nID, "", ExtraFilesInGOPATH)
	fname, extra = getFileName(nID, "", ExtraFilesInGOPATH)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	checkOut(t, errExOut, getFnameMatchText)
}

func TestReadYesNo(t *testing.T) {
	yesnoMatchText := fmt.Sprintf("Answer is [y/n]: ")
	buffer := bytes.Buffer{}
	input := "yes\n"
	if !readYesNo("Answer is", strings.NewReader(input), &buffer) {
		t.Errorf("answer is NOT yes, but '%s'\n", buffer.String())
	}
	txt := buffer.String()
	checkOut(t, txt, yesnoMatchText)

	buffer = bytes.Buffer{}
	input = "no\n"
	if readYesNo("Answer is", strings.NewReader(input), &buffer) {
		t.Errorf("answer is NOT no, but '%s'\n", buffer.String())
	}
	txt = buffer.String()
	checkOut(t, txt, yesnoMatchText)
}

func TestPrintHelpAndExit(t *testing.T) {
	tstRetErrorExit = -1
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	errExitMatchText := fmt.Sprintf(`saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | revertall | enabled | applied ]
  saptune note [ apply | simulate | verify | customise | create | edit | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled | applied ]
  saptune solution [ apply | simulate | verify | customise | create | edit | revert | show | delete ] SolutionName
  saptune solution rename SolutionName newSolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff | analysis | release ]
   saptune staging [ analysis | diff ] [ NoteID... | SolutionID... | all ]
   saptune staging release [--force|--dry-run] [ NoteID... | SolutionID... | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Call external script '/usr/sbin/saptune_check'
  saptune check
Print current saptune status:
  saptune status
Print current saptune version:
  saptune version
Print this message:
  saptune help
`)
	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer
	buffer := bytes.Buffer{}
	PrintHelpAndExit(&buffer, 9)
	txt := buffer.String()
	checkOut(t, txt, errExitMatchText)
	if tstRetErrorExit != 9 {
		t.Errorf("error exit should be '9' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}
}
