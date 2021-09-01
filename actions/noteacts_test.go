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

	// Test NoteActionList
	t.Run("NoteActionList", func(t *testing.T) {
		var listMatchText = `
All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note, C denotes custom note):
	extraNote	Configuration drop in for extra tests
			Version 0 from 04.06.2019 
	oldFile		Name_syntax
	simpleNote	Configuration drop in for simple tests
			Version 1 from 09.07.2019 

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`

		buffer := bytes.Buffer{}
		NoteActionList(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)
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
simpleNote - Configuration drop in for simple tests
			Version 1 from 09.07.2019  

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

current order of enabled notes is: simpleNote

The system fully conforms to the specified note.
`
		buffer := bytes.Buffer{}
		nID := "simpleNote"
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
	os.Setenv("EDITOR", "/usr/bin/echo")

	newTuningOpts := note.GetTuningOptions("", ExtraFilesInGOPATH)
	nApp := app.InitialiseApp(TstFilesInGOPATH, "", newTuningOpts, AllTestSolutions)
	// test with missing template file
	nID := "hugo"
	createMatchText := fmt.Sprintf("ERROR: Problems while editing note definition file '/etc/saptune/extra/hugo.conf' - open /usr/share/saptune/NoteTemplate.conf: no such file or directory\n")
	NoteActionCreate(nID, nApp)
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt := buffer.String()
	checkOut(t, txt, createMatchText)

	templateFile = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/NoteTemplate.conf")
	// test with available template file
	buffer.Reset()
	createMatchText = fmt.Sprintf("")
	tstRetErrorExit = -1
	oldExtraTuningSheets := ExtraTuningSheets
	defer func() { ExtraTuningSheets = oldExtraTuningSheets }()
	ExtraTuningSheets = ExtraFilesInGOPATH
	fname := fmt.Sprintf("%s%s.conf", ExtraTuningSheets, nID)
	NoteActionCreate(nID, nApp)
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt = buffer.String()
	checkOut(t, txt, createMatchText)
	if _, err := os.Stat(fname); err == nil {
		t.Errorf("found a created file '%s' even that no input was provided to the editor", fname)
	}
	os.Remove(fname)
	os.Setenv("EDITOR", oldEditor)
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
