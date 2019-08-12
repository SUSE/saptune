package main

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"os"
	"os/exec"
	"path"
	"syscall"
	"testing"
)

var OSNotesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes")
var OSPackageInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/")
var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra")
var AllTestSolutions = map[string]solution.Solution{
	"sol1":  solution.Solution{"1001"},
	"sol2":  solution.Solution{"1002"},
	"sol12": solution.Solution{"1001", "1002"},
}

var tuningOpts = note.GetTuningOptions("", TstFilesInGOPATH)
var tApp = app.InitialiseApp(OSPackageInGOPATH, "", tuningOpts, AllTestSolutions)
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

func TestNoteActionList(t *testing.T) {
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
}

func TestNoteActionApply(t *testing.T) {
	var applyMatchText = `The note has been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must instruct saptune to configure "tuned" daemon by running:
    saptune daemon start
`
	buffer := bytes.Buffer{}
	nID := "simpleNote"
	NoteActionApply(&buffer, nID, tApp)
	txt := buffer.String()
	checkOut(t, txt, applyMatchText)
}

func TestNoteActionVerify(t *testing.T) {
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
}

func TestNoteActionRevert(t *testing.T) {
	var revertMatchText = `Parameters tuned by the note have been successfully reverted.
Please note: the reverted note may still show up in list of enabled notes, if an enabled solution refers to it.
`
	buffer := bytes.Buffer{}
	nID := "simpleNote"
	NoteActionRevert(&buffer, nID, tApp)
	txt := buffer.String()
	checkOut(t, txt, revertMatchText)
}

func TestPrintNoteFields(t *testing.T) {
	//tuningOptions := note.GetTuningOptions(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes"), "")
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
		//txt := PrintNoteFields("HEAD", noteComp, true)
		checkCorrectMessage(t, txt, printMatchText1)
	})
	t.Run("simulate with header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "HEAD", noteComp, false)
		txt := buffer.String()
		//txt := PrintNoteFields("HEAD", noteComp, false)
		checkCorrectMessage(t, txt, printMatchText2)
	})
	t.Run("verify without header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "NONE", noteComp, true)
		txt := buffer.String()
		//txt := PrintNoteFields("NONE", noteComp, true)
		checkCorrectMessage(t, txt, printMatchText3)
	})
	t.Run("simulate without header", func(t *testing.T) {
		buffer := bytes.Buffer{}
		PrintNoteFields(&buffer, "NONE", noteComp, false)
		txt := buffer.String()
		//txt := PrintNoteFields("NONE", noteComp, false)
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
