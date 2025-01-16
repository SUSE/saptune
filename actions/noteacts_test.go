package actions

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"strings"
	"testing"
)

func TestNoteActions(t *testing.T) {
	// test setup
	setUp(t)

	tstRetErrorExit = -1
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut
	errExBuffer := bytes.Buffer{}
	tstwriter = &errExBuffer

	// Test NoteActionList
	t.Run("NoteActionList", func(t *testing.T) {
		listMatchText := noteListMatchText
		buffer := bytes.Buffer{}
		NoteActionList(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)

		buffer.Reset()
		var listMatchText2 = `
All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note, C denotes custom note, D denotes deprecated notes):
 C	900929		Linux: STORAGE_PARAMETERS_WRONG_SET and 'mmap() failed'
			Version 7 from 31.07.2017
			https://me.sap.com/notes/900929
 [32m+	NEWSOL2NOTE	
[0m [32m- O	extraNote	Configuration drop in for extra tests
			Version 0 from 04.06.2019
[0m	oldFile		Name_syntax
 [32m*	simpleNote	Configuration drop in for simple tests
			Version 1 from 09.07.2019
[0m	wrongFileNamesyntax	

current order of enabled notes is: simpleNote NEWSOL2NOTE


Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`
		oldExtraTuningSheets := ExtraTuningSheets
		defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
		ExtraTuningSheets = ExtraTstFilesInGOPATH
		oldOverrideTuningSheets := OverrideTuningSheets
		defer func() { OverrideTuningSheets = oldOverrideTuningSheets }()
		OverrideTuningSheets = OverTstFilesInGOPATH
		tApp.TuneForSolutions = []string{"sol12"}
		tApp.TuneForNotes = []string{"NEWSOL2NOTE"}
		tApp.NoteApplyOrder = []string{"simpleNote", "NEWSOL2NOTE"}

		buffer = bytes.Buffer{}
		NoteActionList(&buffer, tApp)
		txt = buffer.String()
		checkOut(t, txt, listMatchText2)

		// test solutionStillEnabled as we currently have the needed
		// test data available
		solutionStillEnabled(tApp)
		if strings.Join(tApp.TuneForSolutions, " ") != "sol12" {
			t.Errorf("got: '%+v', expected: 'sol12'\n", strings.Join(tApp.TuneForSolutions, " "))
		}
		tApp.NoteApplyOrder = []string{}
		tApp.NoteApplyOrder = []string{"NEWSOL2NOTE"}
		solutionStillEnabled(tApp)
		if strings.Join(tApp.TuneForSolutions, " ") == "sol12" {
			t.Errorf("got: '%+v', expected: ''\n", strings.Join(tApp.TuneForSolutions, " "))
		}

		tApp.TuneForSolutions = []string{}
		tApp.TuneForNotes = []string{}
		tApp.NoteApplyOrder = []string{}

	})

	// Test NoteActionSimulate
	t.Run("NoteActionSimulate", func(t *testing.T) {
		var simulateMatchText = `If you run ` + "`saptune note apply simpleNote`" + `, the following changes will be applied to your system:

simpleNote - Configuration drop in for simple tests
			Version 1 from 09.07.2019 

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

		simBuf.Reset()
		errExBuffer.Reset()
		tstRetErrorExit = -1
		errMatchText := `ERROR: Failed to test the current system against the specified note: the Note ID "" is not recognised by saptune.
Run "saptune note list" for a complete list of supported notes.
and then please double check your input and /etc/sysconfig/saptune
`
		errExitMatchText := PrintHelpAndExitMatchText

		NoteActionSimulate(&simBuf, "", tApp)
		txt = simBuf.String()
		checkOut(t, txt, errExitMatchText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errtxt := errExBuffer.String()
		checkOut(t, errtxt, errMatchText)
	})

	// Test NoteActionApply
	t.Run("NoteActionApply", func(t *testing.T) {
		var applyMatchText = `The note has been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`

		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionApply(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, applyMatchText)

		buffer.Reset()
		errExBuffer.Reset()
		tstRetErrorExit = -1
		errMatchText := `ERROR: Failed to tune for note : the Note ID "" is not recognised by saptune.
Run "saptune note list" for a complete list of supported notes.
and then please double check your input and /etc/sysconfig/saptune
`
		errExitMatchText := PrintHelpAndExitMatchText + applyMatchText

		NoteActionApply(&buffer, "", tApp)
		txt = buffer.String()
		checkOut(t, txt, errExitMatchText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errtxt := errExBuffer.String()
		checkOut(t, errtxt, errMatchText)
	})

	// Test VerifyAllParameters
	t.Run("VerifyAllParameters", func(t *testing.T) {
		var verifyMatchText = `
   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m

current order of enabled notes is: simpleNote

[32m[1mThe running system is currently well-tuned according to all of the enabled notes.[22m[0m
`
		buffer := bytes.Buffer{}
		VerifyAllParameters(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test NoteActionVerify
	t.Run("NoteActionVerify", func(t *testing.T) {
		var verifyMatchText = `
simpleNote - Configuration drop in for simple tests
			Version 1 from 09.07.2019 

   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m

current order of enabled notes is: simpleNote

[32m[1mThe system fully conforms to the specified note.[22m[0m
`
		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionVerify(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test NoteActionVerifyApplied
	t.Run("NoteActionVerifyApplied", func(t *testing.T) {
		var verifyMatchText = `
   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m

current order of enabled notes is: simpleNote

[32m[1mThe running system is currently well-tuned according to all of the enabled notes.[22m[0m
`
		buffer := bytes.Buffer{}
		nID := "applied"
		NoteActionVerify(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test NoteActionEnabled
	t.Run("NoteActionEnabled", func(t *testing.T) {
		enabledMatchText := "simpleNote"

		buffer := bytes.Buffer{}
		NoteActionEnabled(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, enabledMatchText)
	})

	// Test NoteActionApplied
	t.Run("NoteActionApplied", func(t *testing.T) {
		appliedMatchText := "simpleNote"

		buffer := bytes.Buffer{}
		NoteActionApplied(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, appliedMatchText)
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
		oldNoteTuningSheets := NoteTuningSheets
		defer func() { NoteTuningSheets = oldNoteTuningSheets }()
		NoteTuningSheets = ""
		oldExtraTuningSheets := ExtraTuningSheets
		defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
		ExtraTuningSheets = ExtraFilesInGOPATH

		buffer := bytes.Buffer{}
		nID := "simpleNote"
		NoteActionShow(&buffer, nID, tApp)
		txt := buffer.String()
		checkOut(t, txt, showMatchText)

		buffer.Reset()
		errExBuffer.Reset()
		tstRetErrorExit = -1
		errMatchText := `ERROR: the Note ID "" is not recognised by saptune.
Run "saptune note list" for a complete list of supported notes.
and then please double check your input and /etc/sysconfig/saptune
ERROR: Note  not found in  or /home/ci_tst/gopath/src/github.com/SUSE/saptune/testdata/extra/.
ERROR: Failed to read file '' - open : no such file or directory
`
		errExitMatchText := PrintHelpAndExitMatchText + `
Content of Note :

`

		NoteActionShow(&buffer, "", tApp)
		txt = buffer.String()
		checkOut(t, txt, errExitMatchText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errtxt := errExBuffer.String()
		checkOut(t, errtxt, errMatchText)
	})

	tearDown(t)
}

func TestNoteActionCreate(t *testing.T) {
	tstRetErrorExit = -1
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut
	buffer := bytes.Buffer{}
	tstwriter = &buffer

	oldEditor := os.Getenv("EDITOR")
	defer func() { os.Setenv("EDITOR", oldEditor) }()
	os.Setenv("EDITOR", "/usr/bin/echo")

	newTuningOpts := note.GetTuningOptions("", ExtraFilesInGOPATH)
	nApp := app.InitialiseApp(TstFilesInGOPATH, "", newTuningOpts, AllTestSolutions)
	createBuf := bytes.Buffer{}

	// test with missing template file
	nID := "hugo"
	createMatchText := "ERROR: Problems while editing note definition file '/etc/saptune/extra/hugo.conf' - open /usr/share/saptune/NoteTemplate.conf: no such file or directory\n"
	cMatchText := ""
	NoteActionCreate(&createBuf, nID, nApp)
	txt := createBuf.String()
	checkOut(t, txt, cMatchText)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, createMatchText)

	templateFile = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/NoteTemplate.conf")
	// test with available template file
	createBuf.Reset()
	buffer.Reset()
	createMatchText = ""
	cMatchText = ""
	tstRetErrorExit = -1
	oldExtraTuningSheets := ExtraTuningSheets
	defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
	ExtraTuningSheets = ExtraFilesInGOPATH
	fname := fmt.Sprintf("%s%s.conf", ExtraTuningSheets, nID)
	NoteActionCreate(&createBuf, nID, nApp)
	txt = createBuf.String()
	checkOut(t, txt, cMatchText)
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, createMatchText)
	if _, err := os.Stat(fname); err == nil {
		t.Errorf("found a created file '%s' even that no input was provided to the editor", fname)
	}
	os.Remove(fname)

	// test with empty noteID
	createBuf.Reset()
	buffer.Reset()
	createMatchText = PrintHelpAndExitMatchText
	cMatchText = ""
	tstRetErrorExit = -1
	NoteActionCreate(&createBuf, "", nApp)
	txt = createBuf.String()
	checkOut(t, txt, createMatchText)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, cMatchText)
	if _, err := os.Stat(fname); err == nil {
		t.Errorf("found a created file '%s' even that no input was provided to the editor", fname)
	}
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

	oldNoteTuningSheets := NoteTuningSheets
	defer func() { NoteTuningSheets = oldNoteTuningSheets }()
	NoteTuningSheets = ""
	oldExtraTuningSheets := ExtraTuningSheets
	defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
	ExtraTuningSheets = ExtraFilesInGOPATH
	oldOverrideTuningSheets := OverrideTuningSheets
	defer func() { OverrideTuningSheets = oldOverrideTuningSheets }()
	OverrideTuningSheets = OverTstFilesInGOPATH

	buffer := bytes.Buffer{}
	nID := "extraSimple"
	fileName := fmt.Sprintf("%s%s.conf", ExtraFilesInGOPATH, nID)
	newID := "renameSimple"
	newFileName := fmt.Sprintf("%s%s.conf", ExtraFilesInGOPATH, newID)

	// copy an extra note for later rename
	fsrc := fmt.Sprintf("%ssimpleNote.conf", ExtraFilesInGOPATH)
	if err := system.CopyFile(fsrc, fileName); err != nil {
		t.Fatalf("copy of %s to %s failed: '%+v'", fsrc, fileName, err)
	}

	// check note files and show content of test note
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exist\n", fileName)
	}
	if _, err := os.Stat(newFileName); !os.IsNotExist(err) {
		t.Errorf("file '%s' already exists\n", newFileName)
	}
	// refresh note list (AllNotes)
	newTuningOpts := note.GetTuningOptions("", ExtraFilesInGOPATH)
	nApp := app.InitialiseApp(TstFilesInGOPATH, "", newTuningOpts, AllTestSolutions)

	NoteActionShow(&buffer, nID, nApp)
	txt := buffer.String()
	checkOut(t, txt, showMatchText)

	// test rename of note
	// stop rename of test note
	noRenameBuf := bytes.Buffer{}
	input := "no\n"
	//add additional test without override file later
	confirmRenameMatchText := fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) to the new name '%s'? [y/n]: ", nID, newID)

	NoteActionRename(strings.NewReader(input), &noRenameBuf, nID, newID, nApp)
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

	NoteActionRename(strings.NewReader(input), &renameBuf, nID, newID, nApp)
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
	NoteActionShow(&showRenameBuf, newID, rApp)
	txt = showRenameBuf.String()
	checkOut(t, txt, renameMatchText)

	// test delete of note
	// stop delete of test note
	noDeleteBuf := bytes.Buffer{}
	input = "no\n"
	deleteMatchText := fmt.Sprintf("Note to delete is a customer/vendor specific Note.\nDo you really want to delete this Note (%s)? [y/n]: ", newID)

	NoteActionDelete(strings.NewReader(input), &noDeleteBuf, newID, rApp)
	txt = noDeleteBuf.String()
	checkOut(t, txt, deleteMatchText)
	if _, err := os.Stat(newFileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exists\n", newFileName)
	}

	// delete test note
	deleteBuf := bytes.Buffer{}
	input = "yes\n"

	NoteActionDelete(strings.NewReader(input), &deleteBuf, newID, rApp)
	txt = deleteBuf.String()
	checkOut(t, txt, deleteMatchText)
	if _, err := os.Stat(newFileName); !os.IsNotExist(err) {
		// as 'note delete' has failed, use system to clean up
		if err := os.Remove(newFileName); err != nil {
			t.Fatalf("remove of %s failed", newFileName)
		}
		t.Errorf("file '%s' still exists\n", newFileName)
	}
}

func TestNoteActionCustomise(t *testing.T) {
	tstRetErrorExit = -1
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut
	buffer := bytes.Buffer{}
	tstwriter = &buffer

	oldEditor := os.Getenv("EDITOR")
	defer func() { os.Setenv("EDITOR", oldEditor) }()
	os.Setenv("EDITOR", "/usr/bin/echo")

	oldExtraTuningSheets := ExtraTuningSheets
	defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
	ExtraTuningSheets = ExtraTstFilesInGOPATH
	oldOverrideTuningSheets := OverrideTuningSheets
	defer func() { OverrideTuningSheets = oldOverrideTuningSheets }()
	OverrideTuningSheets = OverTstFilesInGOPATH
	oldNoteTuningSheets := NoteTuningSheets
	defer func() { NoteTuningSheets = oldNoteTuningSheets }()
	NoteTuningSheets = TstFilesInGOPATH
	newTuningOpts := note.GetTuningOptions(TstFilesInGOPATH, ExtraTstFilesInGOPATH)
	cApp := app.InitialiseApp(TstFilesInGOPATH, "", newTuningOpts, AllTestSolutions)

	// test with empty noteID
	custBuffer := bytes.Buffer{}
	custMatchText := ""
	// as we are in 'test mode' collect all errExit messages and continue,
	// instead of exit function - so differ from real life
	// test with missing note id
	errMatchText := `ERROR: the Note ID "" is not recognised by saptune.
Run "saptune note list" for a complete list of supported notes.
and then please double check your input and /etc/sysconfig/saptune
ERROR: Problems while editing note definition file '/home/ci_tst/gopath/src/github.com/SUSE/saptune/testdata/etc/saptune/override/' - write /tmp/.sttemp: copy_file_range: is a directory
`
	NoteActionCustomise(&custBuffer, "", cApp)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt := buffer.String()
	checkOut(t, txt, errMatchText)

	// change EDITOR command
	fakeEditorCommand := path.Join(TstFilesInGOPATH, "tstedit")
	os.Setenv("EDITOR", fakeEditorCommand)
	editorTxt := `Hello from test editor
`

	// test with existing override file - testNote (not applied)
	buffer.Reset()
	custBuffer.Reset()
	errMatchText = ""
	tstRetErrorExit = -1
	errMatchText = ""
	overFileName := path.Join(OverTstFilesInGOPATH, "testNote")
	NoteActionCustomise(&custBuffer, "testNote", cApp)
	custTxt := custBuffer.String()
	checkOut(t, custTxt, custMatchText)
	cont, err := system.ReadConfigFile(overFileName, false)
	if err != nil {
		t.Error(err)
	}
	if string(cont) != editorTxt {
		t.Errorf("got: '%+v', expected: '%s'\n", string(cont), editorTxt)
	}
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)

	// test without override file - test2Note (applied Note)
	buffer.Reset()
	custBuffer.Reset()
	errMatchText = ""
	tstRetErrorExit = -1
	errMatchText = ""
	// fake note applied
	cApp.NoteApplyOrder = []string{"test2Note"}
	emptySrc := path.Join(TstFilesInGOPATH, "saptune_NOEXIT")
	stateFile := cApp.State.GetPathToNote("test2Note")
	os.MkdirAll(path.Join(TstFilesInGOPATH, "/data/run/saptune/saved_state"), 0755)
	if err := system.CopyFile(emptySrc, stateFile); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(path.Join(TstFilesInGOPATH, "/data"))

	overFileName = path.Join(OverTstFilesInGOPATH, "test2Note")
	NoteActionCustomise(&custBuffer, "test2Note", cApp)
	custTxt = custBuffer.String()
	checkOut(t, custTxt, custMatchText)
	if _, err := os.Stat(overFileName); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exists\n", overFileName)
	}
	cont, err = system.ReadConfigFile(overFileName, false)
	if err != nil {
		t.Error(err)
	}
	if string(cont) != editorTxt {
		t.Errorf("got: '%+v', expected: '%s'\n", string(cont), editorTxt)
	}
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)
	// remove testdata/etc/saptune/overrid/test2Note at the end
	cApp.NoteApplyOrder = []string{}
	os.Remove(overFileName)

	// test without override file and without changes - test2Note
	os.Setenv("EDITOR", "/usr/bin/echo")
	buffer.Reset()
	custBuffer.Reset()
	errMatchText = ""
	tstRetErrorExit = -1
	errMatchText = ""
	overFileName = path.Join(OverTstFilesInGOPATH, "test2Note")
	NoteActionCustomise(&custBuffer, "test2Note", cApp)
	custTxt = custBuffer.String()
	checkOut(t, custTxt, custMatchText)
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)

	if _, err := os.Stat(overFileName); !os.IsNotExist(err) {
		t.Errorf("file '%s' exists, even that we do not change it\n", overFileName)
		cont, err = system.ReadConfigFile(overFileName, false)
		if err != nil {
			t.Error(err)
		} else {
			t.Errorf("content is '%+v'\n", string(cont))
		}
		// remove testdata/etc/saptune/overrid/test2Note at the end
		os.Remove(overFileName)
	}
}

func TestNoteActionEdit(t *testing.T) {
	tstRetErrorExit = -1
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut
	buffer := bytes.Buffer{}
	tstwriter = &buffer

	oldEditor := os.Getenv("EDITOR")
	defer func() { os.Setenv("EDITOR", oldEditor) }()
	os.Setenv("EDITOR", "/usr/bin/echo")

	oldExtraTuningSheets := ExtraTuningSheets
	defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
	ExtraTuningSheets = ExtraTstFilesInGOPATH
	oldOverrideTuningSheets := OverrideTuningSheets
	defer func() { OverrideTuningSheets = oldOverrideTuningSheets }()
	OverrideTuningSheets = OverTstFilesInGOPATH
	oldNoteTuningSheets := NoteTuningSheets
	defer func() { NoteTuningSheets = oldNoteTuningSheets }()
	NoteTuningSheets = TstFilesInGOPATH
	newTuningOpts := note.GetTuningOptions(TstFilesInGOPATH, ExtraTstFilesInGOPATH)
	eApp := app.InitialiseApp(TstFilesInGOPATH, "", newTuningOpts, AllTestSolutions)

	// test with empty noteID
	editBuffer := bytes.Buffer{}
	editMatchText := ""
	// as we are in 'test mode' collect all errExit messages and continue,
	// instead of exit function - so differ from real life
	// test with missing note id
	errMatchText := `ERROR: the Note ID "" is not recognised by saptune.
Run "saptune note list" for a complete list of supported notes.
and then please double check your input and /etc/sysconfig/saptune
ERROR: The Note definition file you want to edit is a saptune internal (shipped) Note and can NOT be edited. Use 'saptune note customise' instead. Exiting ...
ERROR: Problems while editing Note definition file '/home/ci_tst/gopath/src/github.com/SUSE/saptune/testdata/' - write /tmp/.sttemp: copy_file_range: is a directory
`
	NoteActionEdit(&editBuffer, "", eApp)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt := buffer.String()
	checkOut(t, txt, errMatchText)

	// change EDITOR command
	fakeEditorCommand := path.Join(TstFilesInGOPATH, "tstedit")
	os.Setenv("EDITOR", fakeEditorCommand)
	editorTxt := `Hello from test editor
`

	// test with a system note - testNote
	buffer.Reset()
	editBuffer.Reset()
	editMatchText = ""
	//editMatchText = PrintHelpAndExitMatchText
	errMatchText = `ERROR: The Note definition file you want to edit is a saptune internal (shipped) Note and can NOT be edited. Use 'saptune note customise' instead. Exiting ...
`
	tstRetErrorExit = -1
	NoteActionEdit(&editBuffer, "testNote", eApp)
	editTxt := editBuffer.String()
	checkOut(t, editTxt, editMatchText)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)

	// test with existing override file - extraTestNote (not applied)
	buffer.Reset()
	editBuffer.Reset()
	editMatchText = ""
	errMatchText = ""
	tstRetErrorExit = -1
	extraFileName := path.Join(ExtraTstFilesInGOPATH, "extraTestNote.conf")
	NoteActionEdit(&editBuffer, "extraTestNote", eApp)
	editTxt = editBuffer.String()
	checkOut(t, editTxt, editMatchText)
	cont, err := system.ReadConfigFile(extraFileName, false)
	if err != nil {
		t.Error(err)
	}
	if string(cont) != editorTxt {
		t.Errorf("got: '%+v', expected: '%s'\n", string(cont), editorTxt)
	}
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)

	// test without override file - extraTest2Note (applied Note)
	buffer.Reset()
	editBuffer.Reset()
	errMatchText = ""
	tstRetErrorExit = -1
	errMatchText = ""
	// fake note applied
	eApp.NoteApplyOrder = []string{"extraTest2Note"}
	emptySrc := path.Join(TstFilesInGOPATH, "saptune_NOEXIT")
	stateFile := eApp.State.GetPathToNote("extraTest2Note")
	os.MkdirAll(path.Join(TstFilesInGOPATH, "/data/run/saptune/saved_state"), 0755)
	if err := system.CopyFile(emptySrc, stateFile); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(path.Join(TstFilesInGOPATH, "/data"))

	extraFileName = path.Join(ExtraTstFilesInGOPATH, "extraTest2Note.conf")
	NoteActionEdit(&editBuffer, "extraTest2Note", eApp)
	editTxt = editBuffer.String()
	checkOut(t, editTxt, editMatchText)
	cont, err = system.ReadConfigFile(extraFileName, false)
	if err != nil {
		t.Error(err)
	}
	if string(cont) != editorTxt {
		t.Errorf("got: '%+v', expected: '%s'\n", string(cont), editorTxt)
	}
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)
	eApp.NoteApplyOrder = []string{}

	// test without changes - extraTest2Note
	os.Setenv("EDITOR", "/usr/bin/echo")
	buffer.Reset()
	editBuffer.Reset()
	errMatchText = ""
	tstRetErrorExit = -1
	errMatchText = ""
	NoteActionEdit(&editBuffer, "extraTest2Note", eApp)
	editTxt = editBuffer.String()
	checkOut(t, editTxt, editMatchText)
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, errMatchText)

}
