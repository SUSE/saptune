package main

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
)

var OSNotesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes")
var ExtraFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra") + "/"
var OverTstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/etc/saptune/override") + "/"
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

func TestSetWidthOfColums(t *testing.T) {
	compare := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "IO_SCHEDULER_sr0", ActualValueJS: "cfq", ExpectedValueJS: "cfq"}
	w1 := 2
	w2 := 3
	w3 := 4
	w4 := 5
	v1, v2, v3, v4 := setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != w1 {
		t.Fatal(v1, w1)
	}
	if v2 != 16 {
		t.Fatal(v2, w2)
	}
	if v3 != w3 || v4 != w4 {
		t.Fatal(v3, w3, v4, w4)
	}
	compare = note.FieldComparison{ReflectFieldName: "OverrideParams", ReflectMapKey: "IO_SCHEDULER_sr0", ActualValueJS: "cfq", ExpectedValueJS: "cfq"}
	v1, v2, v3, v4 = setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != 3 {
		t.Fatal(v1, w1)
	}
	if v2 != w2 || v3 != w3 || v4 != w4 {
		t.Fatal(v2, w2, v3, w3, v4, w4)
	}
	compare = note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "governor", ActualValueJS: "all-none", ExpectedValueJS: "all-performance"}
	v1, v2, v3, v4 = setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != w1 {
		t.Fatal(v1, w1)
	}
	if v2 != 8 {
		t.Fatal(v2, w2)
	}
	if v3 != 15 {
		t.Fatal(v3, w3)
	}
	if v4 != 8 {
		t.Fatal(v4, w4)
	}
	compare = note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "", ActualValueJS: "all-none", ExpectedValueJS: "all-performance"}
	v1, v2, v3, v4 = setWidthOfColums(compare, w1, w2, w3, w4)
	if v1 != w1 || v2 != w2 || v3 != w3 || v4 != w4 {
		t.Fatal(v1, w1, v2, w2, v3, w3, v4, w4)
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

func TestNoteActions(t *testing.T) {
	// test setup
	setUp(t)

	// Test NoteActionList
	t.Run("NoteActionList", func(t *testing.T) {
		var listMatchText = `
All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note):
	extraNote	Configuration drop in for extra tests
			Version 0 from 04.06.2019 
	oldFile		Name_syntax
	simpleNote	Configuration drop in for simple tests
			Version 1 from 09.07.2019 
Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must instruct saptune to configure "tuned" daemon by running:
    saptune daemon start
`

		buffer := bytes.Buffer{}
		NoteActionList(&buffer, tApp, tuningOpts)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)
	})

	// Test NoteActionSimulate
	t.Run("NoteActionSimulate", func(t *testing.T) {
		var simulateMatchText = `If you run ` + "`saptune note apply simpleNote`" + `, the following changes will be applied to your system:

simpleNote -  

   Parameter                    | Value set   | Value expected  | Override  | Comment
--------------------------------+-------------+-----------------+-----------+--------------
   net.ipv4.ip_local_port_range | 32768 60999 | 31768 61999     |           |   


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
`
		simBuf := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionSimulate(&simBuf, nID, tApp)
		txt := simBuf.String()
		checkOut(t, txt, simulateMatchText)
	})

	// Test NoteActionApply
	t.Run("NoteActionApply", func(t *testing.T) {
		var applyMatchText = `The note has been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must instruct saptune to configure "tuned" daemon by running:
    saptune daemon start
`
		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionApply(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, applyMatchText)
	})

	// Test VerifyAllParameters
	t.Run("VerifyAllParameters", func(t *testing.T) {
		var verifyMatchText = `   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m

current order of applied notes is: simpleNote

The running system is currently well-tuned according to all of the enabled notes.
`
		buffer := bytes.Buffer{}
		VerifyAllParameters(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test NoteActionVerify
	t.Run("NoteActionVerify", func(t *testing.T) {
		var verifyMatchText = `
simpleNote -  

   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes

   (no change)


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m

current order of applied notes is: simpleNote

The system fully conforms to the specified note.
`
		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionVerify(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test NoteActionRevert
	t.Run("NoteActionRevert", func(t *testing.T) {
		var revertMatchText = `Parameters tuned by the note have been successfully reverted.
`
		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionRevert(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, revertMatchText)
	})

	// Test NoteActionShow
	t.Run("NoteActionShow", func(t *testing.T) {
		var showMatchText = `
Content of Note simpleNote:
[version]
# SAP-NOTE=simpleNote CATEGORY=simple VERSION=1 DATE=09.07.2019 NAME="Configuration drop in for simple tests" 

[sysctl]
net.ipv4.ip_local_port_range = 31768 61999

[reminder]
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.

`
		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionShow(&buffer, nID, "", ExtraFilesInGOPATH, tApp)
		txt := buffer.String()
		checkOut(t, txt, showMatchText)
	})

	tearDown(t)
}

func TestNoteActionRenameShowDelete(t *testing.T) {
	var showMatchText = `
Content of Note extraSimple:
[version]
# SAP-NOTE=simpleNote CATEGORY=simple VERSION=1 DATE=09.07.2019 NAME="Configuration drop in for simple tests" 

[sysctl]
net.ipv4.ip_local_port_range = 31768 61999

[reminder]
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.

`
	var renameMatchText = `
Content of Note renameSimple:
[version]
# SAP-NOTE=simpleNote CATEGORY=simple VERSION=1 DATE=09.07.2019 NAME="Configuration drop in for simple tests" 

[sysctl]
net.ipv4.ip_local_port_range = 31768 61999

[reminder]
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.

`
	buffer := bytes.Buffer{}
	nID := "extraSimple"
	fileName := fmt.Sprintf("%s%s.conf", ExtraFilesInGOPATH, nID)
	ovFileName := fmt.Sprintf("%s%s", OverTstFilesInGOPATH, nID)
	newID := "renameSimple"
	newFileName := fmt.Sprintf("%s%s.conf", ExtraFilesInGOPATH, newID)
	newovFileName := fmt.Sprintf("%s%s", OverTstFilesInGOPATH, newID)

	// copy an extra note for later rename
	fsrc := fmt.Sprintf("%ssimpleNote.conf", ExtraFilesInGOPATH)
	if err := system.CopyFile(fsrc, fileName); err != nil {
		t.Fatalf("copy of %s to %s failed: '%+v'", fsrc, fileName, err)
	}
	if err := system.CopyFile(fileName, ovFileName); err != nil {
		t.Fatalf("copy of %s to %s failed: '%+v'", fileName, ovFileName, err)
	}

	// check note files and show content of test note
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exist\n", fileName)
	}
	if _, err := os.Stat(ovFileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exist\n", ovFileName)
	}
	if _, err := os.Stat(newFileName); !os.IsNotExist(err) {
		t.Errorf("file '%s' already exists\n", newFileName)
	}
	// refresh note list (AllNotes)
	newTuningOpts := note.GetTuningOptions("", ExtraFilesInGOPATH)
	nApp := app.InitialiseApp(TstFilesInGOPATH, "", newTuningOpts, AllTestSolutions)

	NoteActionShow(&buffer, nID, "", ExtraFilesInGOPATH, nApp)
	txt := buffer.String()
	checkOut(t, txt, showMatchText)

	// test rename of note
	// stop rename of test note
	noRenameBuf := bytes.Buffer{}
	input := "no\n"
	//add additional test without override file later
	//confirmRenameMatchText := fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) to the new name '%s'? [y/n]: ", nID, newID)
	confirmRenameMatchText := fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) and the corresponding override file to the new name '%s'? [y/n]: ", nID, newID)

	NoteActionRename(strings.NewReader(input), &noRenameBuf, nID, newID, "", ExtraFilesInGOPATH, OverTstFilesInGOPATH, nApp)
	txt = noRenameBuf.String()
	checkOut(t, txt, confirmRenameMatchText)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exist\n", fileName)
	}
	if _, err := os.Stat(newFileName); !os.IsNotExist(err) {
		t.Errorf("file '%s' already exists\n", newFileName)
	}

	// rename test note
	renameBuf := bytes.Buffer{}
	input = "yes\n"

	NoteActionRename(strings.NewReader(input), &renameBuf, nID, newID, "", ExtraFilesInGOPATH, OverTstFilesInGOPATH, nApp)
	txt = renameBuf.String()
	checkOut(t, txt, confirmRenameMatchText)
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Errorf("file '%s' still exists\n", fileName)
	}
	if _, err := os.Stat(newFileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exist\n", newFileName)
	}

	// show content of renamed note
	// refresh note list (AllNotes) for 'Show'
	renTuningOpts := note.GetTuningOptions("", ExtraFilesInGOPATH)
	rApp := app.InitialiseApp(TstFilesInGOPATH, "", renTuningOpts, AllTestSolutions)

	showRenameBuf := bytes.Buffer{}
	NoteActionShow(&showRenameBuf, newID, "", ExtraFilesInGOPATH, rApp)
	txt = showRenameBuf.String()
	checkOut(t, txt, renameMatchText)

	// test delete of note
	// stop delete of test note
	noDeleteBuf := bytes.Buffer{}
	input = "no\n"
	//add additional test without override file later
	//deleteMatchText := fmt.Sprintf("Note to delete is a customer/vendor specific Note.\nDo you really want to delete this Note (%s)? [y/n]: ", newID)
	deleteMatchText := fmt.Sprintf("Note to delete is a customer/vendor specific Note.\nDo you really want to delete this Note (%s) and the corresponding override file? [y/n]: ", newID)

	NoteActionDelete(strings.NewReader(input), &noDeleteBuf, newID, "", ExtraFilesInGOPATH, OverTstFilesInGOPATH, rApp)
	txt = noDeleteBuf.String()
	checkOut(t, txt, deleteMatchText)
	if _, err := os.Stat(newFileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exists\n", newFileName)
	}

	// delete test note
	deleteBuf := bytes.Buffer{}
	input = "yes\n"

	NoteActionDelete(strings.NewReader(input), &deleteBuf, newID, "", ExtraFilesInGOPATH, OverTstFilesInGOPATH, rApp)
	txt = deleteBuf.String()
	checkOut(t, txt, deleteMatchText)
	if _, err := os.Stat(newFileName); !os.IsNotExist(err) {
		// as 'note delete' has failed, use system to clean up
		if err := os.Remove(newFileName); err != nil {
			t.Fatalf("remove of %s failed", newFileName)
		}
		if _, err := os.Stat(newovFileName); !os.IsNotExist(err) {
			// as 'note delete' has failed, use system to clean up
			if err := os.Remove(newovFileName); err != nil {
				t.Fatalf("remove of %s failed", newovFileName)
			}
			t.Errorf("file '%s' still exists\n", newovFileName)
		}
		t.Errorf("file '%s' still exists\n", newFileName)
	}
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

func TestPrintNoteFields(t *testing.T) {
	var printMatchText1 = `
941735 -  

   SAPNote, Version | Parameter           | Expected             | Override  | Actual               | Compliant
--------------------+---------------------+----------------------+-----------+----------------------+-----------
   941735,          | ShmFileSystemSizeMB | 1714                 |           | 488                  | no 
   941735,          | kernel.shmmax       | 18446744073709551615 |           | 18446744073709551615 | yes


`
	var printMatchText2 = `
941735 -  

   Parameter           | Value set            | Value expected       | Override  | Comment
-----------------------+----------------------+----------------------+-----------+--------------
   ShmFileSystemSizeMB | 488                  | 1714                 |           |   
   kernel.shmmax       | 18446744073709551615 | 18446744073709551615 |           |   


`
	var printMatchText3 = `   SAPNote, Version | Parameter           | Expected             | Override  | Actual               | Compliant
--------------------+---------------------+----------------------+-----------+----------------------+-----------
   941735,          | ShmFileSystemSizeMB | 1714                 |           | 488                  | no 
   941735,          | kernel.shmmax       | 18446744073709551615 |           | 18446744073709551615 | yes


`
	var printMatchText4 = `   Parameter           | Value set            | Value expected       | Override  | Comment
-----------------------+----------------------+----------------------+-----------+--------------
   ShmFileSystemSizeMB | 488                  | 1714                 |           |   
   kernel.shmmax       | 18446744073709551615 | 18446744073709551615 |           |   


`
	checkCorrectMessage := func(t *testing.T, got, want string) {
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

	fcomp1 := note.FieldComparison{ReflectFieldName: "ConfFilePath", ReflectMapKey: "", ActualValue: "/usr/share/saptune/notes/941735", ExpectedValue: "/usr/share/saptune/notes/941735", ActualValueJS: "/usr/share/saptune/notes/941735", ExpectedValueJS: "/usr/share/saptune/notes/941735", MatchExpectation: true}
	fcomp2 := note.FieldComparison{ReflectFieldName: "ID", ReflectMapKey: "", ActualValue: "941735", ExpectedValue: "941735", ActualValueJS: "941735", ExpectedValueJS: "941735", MatchExpectation: true}
	fcomp3 := note.FieldComparison{ReflectFieldName: "DescriptiveName", ReflectMapKey: "", ActualValue: "", ExpectedValue: "", ActualValueJS: "", ExpectedValueJS: "", MatchExpectation: true}
	fcomp4 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "ShmFileSystemSizeMB", ActualValue: "488", ExpectedValue: "1714", ActualValueJS: "488", ExpectedValueJS: "1714", MatchExpectation: false}
	fcomp5 := note.FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "kernel.shmmax", ActualValue: "18446744073709551615", ExpectedValue: "18446744073709551615", ActualValueJS: "18446744073709551615", ExpectedValueJS: "18446744073709551615", MatchExpectation: true}
	map941735 := map[string]note.FieldComparison{"ConfFilePath": fcomp1, "ID": fcomp2, "DescriptiveName": fcomp3, "SysctlParams[ShmFileSystemSizeMB]": fcomp4, "SysctlParams[kernel.shmmax]": fcomp5}
	noteComp := map[string]map[string]note.FieldComparison{"941735": map941735}

	t.Run("verify with header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "HEAD", noteComp, true)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText1)
	})
	t.Run("simulate with header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "HEAD", noteComp, false)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText2)
	})
	t.Run("verify without header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "NONE", noteComp, true)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText3)
	})
	t.Run("simulate without header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "NONE", noteComp, false)
		txt := buffer.String()
		checkCorrectMessage(t, txt, printMatchText4)
	})
}

func TestCheckUpdateLeftOvers(t *testing.T) {
	checkUpdateLeftOvers()
}

/*
func TestRevertAction(t *testing.T) {
	RevertAction("all")
	RevertAction("")
}

func TestDaemonAction(t *testing.T) {
	DaemonAction("start")
}
*/

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

func TestSolutionActions(t *testing.T) {
	// test setup

	// Test SolutionActionList
	t.Run("SolutionActionList", func(t *testing.T) {
		var listMatchText = `
All solutions (* denotes enabled solution, O denotes override file exists for solution, D denotes deprecated solutions):
	BWA                - 941735 2534844 SAP_BWA
	HANA               - 941735 1771258 1980196 1984787 2205917 2382421 2534844
	NETW               - 941735 1771258 1980196 1984787 2534844

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must instruct saptune to configure "tuned" daemon by running:
    saptune daemon start
`

		buffer := bytes.Buffer{}
		SolutionActionList(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)
	})

	// Test SolutionActionSimulate
	t.Run("SolutionActionSimulate", func(t *testing.T) {
		var simulateMatchText = `If you run ` + "`saptune solution apply sol1`" + `, the following changes will be applied to your system:
   Parameter                    | Value set   | Value expected  | Override  | Comment
--------------------------------+-------------+-----------------+-----------+--------------
   net.ipv4.ip_local_port_range | 32768 60999 | 31768 61999     |           |   


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
`
		simBuf := bytes.Buffer{}
		sName := "sol1"
		SolutionActionSimulate(&simBuf, sName, tApp)
		txt := simBuf.String()
		checkOut(t, txt, simulateMatchText)
	})

	// Test SolutionActionApply
	// need to run before 'Test SolutionActionVerify'
	t.Run("SolutionActionApply", func(t *testing.T) {
		var applyMatchText = `All tuning options for the SAP solution have been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must instruct saptune to configure "tuned" daemon by running:
    saptune daemon start
`
		buffer := bytes.Buffer{}
		sName := "sol1"
		SolutionActionApply(&buffer, sName, tApp)
		txt := buffer.String()
		checkOut(t, txt, applyMatchText)
	})

	// Test SolutionActionVerify
	// need to run after 'Test SolutionActionApply'
	t.Run("SolutionActionVerify", func(t *testing.T) {
		var verifyMatchText = `   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
The system fully conforms to the tuning guidelines of the specified SAP solution.
`
		buffer := bytes.Buffer{}
		sName := "sol1"
		SolutionActionVerify(&buffer, sName, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test SolutionActionRevert
	t.Run("SolutionActionRevert", func(t *testing.T) {
		var revertMatchText = `Parameters tuned by the notes referred by the SAP solution have been successfully reverted.
`
		buffer := bytes.Buffer{}
		sName := "sol1"
		SolutionActionRevert(&buffer, sName, tApp)
		txt := buffer.String()
		checkOut(t, txt, revertMatchText)
	})

	tearDown(t)
}
