package actions

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
)

var ExtraFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra") + "/"
var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/")
var AllTestSolutions = map[string]solution.Solution{
	"sol1":  solution.Solution{"simpleNote"},
	"sol2":  solution.Solution{"extraNote"},
	"sol12": solution.Solution{"simpleNote", "extraNote"},
}

var tuningOpts = note.GetTuningOptions("", ExtraFilesInGOPATH)
var tApp = app.InitialiseApp(TstFilesInGOPATH, "", tuningOpts, AllTestSolutions)

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
	exitCode := 0
	if os.Getenv("DO_EXIT") == "1" {
		PrintHelpAndExit(9)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestPrintHelpAndExit")
	cmd.Env = append(os.Environ(), "DO_EXIT=1")
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	if ok {
		ws := e.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		if exitCode != 9 {
			t.Fatalf("process ran with err %v, want exit status 9", err)
		}
		if !e.Success() {
			return
		}
	}
	t.Fatalf("process ran with err %v, want exit status 9", err)
}
