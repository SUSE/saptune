package actions

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
)

var templateFile = "/usr/share/saptune/NoteTemplate.conf"
var editor = os.Getenv("EDITOR")

// NoteAction  Note actions like apply, revert, verify asm.
func NoteAction(actionName, noteID, newNoteID string, tuneApp *app.App) {
	switch actionName {
	case "apply":
		NoteActionApply(os.Stdout, noteID, tuneApp)
	case "list":
		NoteActionList(os.Stdout, tuneApp, tuningOptions)
	case "verify":
		NoteActionVerify(os.Stdout, noteID, tuneApp)
	case "simulate":
		NoteActionSimulate(os.Stdout, noteID, tuneApp)
	case "customise":
		NoteActionCustomise(noteID, tuneApp)
	case "create":
		NoteActionCreate(noteID, tuneApp)
	case "show":
		NoteActionShow(os.Stdout, noteID, NoteTuningSheets, ExtraTuningSheets, tuneApp)
	case "delete":
		NoteActionDelete(os.Stdin, os.Stdout, noteID, NoteTuningSheets, ExtraTuningSheets, OverrideTuningSheets, tuneApp)
	case "rename":
		NoteActionRename(os.Stdin, os.Stdout, noteID, newNoteID, NoteTuningSheets, ExtraTuningSheets, OverrideTuningSheets, tuneApp)
	case "revert":
		NoteActionRevert(os.Stdout, noteID, tuneApp)
	case "enabled":
		NoteActionEnabled(os.Stdout, tuneApp)
	default:
		PrintHelpAndExit(os.Stdout, 1)
	}
}

// NoteActionApply applies Note parameter settings to the system
func NoteActionApply(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}

	// Do not apply the note, if it was applied before
	// Otherwise, the state file (serialised parameters) will be
	// overwritten, and it will no longer be possible to revert the
	// note to the state before it was tuned.
	if str, ok := tuneApp.IsNoteApplied(noteID); ok {
		if str == "" {
			system.InfoLog("note '%s' already applied. Nothing to do", noteID)
		}
		system.ErrorExit("", 0)
	}
	if err := tuneApp.TuneNote(noteID); err != nil {
		system.ErrorExit("Failed to tune for note %s: %v", noteID, err)
	}
	fmt.Fprintf(writer, "The note has been applied successfully.\n")
	rememberMessage(writer)
}

// NoteActionList lists all available Note definitions
func NoteActionList(writer io.Writer, tuneApp *app.App, tOptions note.TuningOptions) {
	fmt.Fprintf(writer, "\nAll notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note):\n")
	solutionNoteIDs := tuneApp.GetSortedSolutionEnabledNotes()
	for _, noteID := range tOptions.GetSortedIDs() {
		noteObj := tOptions[noteID]
		format := "\t%s\t\t%s\n"
		if len(noteID) >= 8 {
			format = "\t%s\t%s\n"
		}
		if _, err := os.Stat(fmt.Sprintf("%s%s", OverrideTuningSheets, noteID)); err == nil {
			format = " O" + format
		}
		if i := sort.SearchStrings(solutionNoteIDs, noteID); i < len(solutionNoteIDs) && solutionNoteIDs[i] == noteID {
			j := tuneApp.PositionInNoteApplyOrder(noteID)
			if j < 0 { // noteID was reverted manually
				format = " " + setGreenText + "-" + format + resetTextColor
			} else {
				format = " " + setGreenText + "*" + format + resetTextColor
			}
		} else if i := sort.SearchStrings(tuneApp.TuneForNotes, noteID); i < len(tuneApp.TuneForNotes) && tuneApp.TuneForNotes[i] == noteID {
			format = " " + setGreenText + "+" + format + resetTextColor
		}
		fmt.Fprintf(writer, format, noteID, noteObj.Name())
	}
	tuneApp.PrintNoteApplyOrder(writer)
	rememberMessage(writer)
}

// NoteActionVerify compares all parameter settings from a Note definition
// against the system settings
func NoteActionVerify(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		VerifyAllParameters(writer, tuneApp)
	} else {
		// Check system parameters against the specified note, no matter the note has been tuned for or not.
		conforming, comparisons, _, err := tuneApp.VerifyNote(noteID)
		if err != nil {
			system.ErrorExit("Failed to test the current system against the specified note: %v", err)
		}
		noteComp := make(map[string]map[string]note.FieldComparison)
		noteComp[noteID] = comparisons
		PrintNoteFields(writer, "HEAD", noteComp, true)
		tuneApp.PrintNoteApplyOrder(writer)
		if !conforming {
			system.ErrorExit("The parameters listed above have deviated from the specified note.\n")
		} else {
			fmt.Fprintf(writer, "The system fully conforms to the specified note.\n")
		}
	}
}

// NoteActionSimulate shows all changes that will be applied to the system if
// the Note will be applied.
func NoteActionSimulate(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	// Run verify and print out all fields of the note
	if _, comparisons, _, err := tuneApp.VerifyNote(noteID); err != nil {
		system.ErrorExit("Failed to test the current system against the specified note: %v", err)
	} else {
		fmt.Fprintf(writer, "If you run `saptune note apply %s`, the following changes will be applied to your system:\n", noteID)
		noteComp := make(map[string]map[string]note.FieldComparison)
		noteComp[noteID] = comparisons
		PrintNoteFields(writer, "HEAD", noteComp, false)
	}
}

// NoteActionCustomise creates an override file and allows to editing the Note
// definition file
func NoteActionCustomise(noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(os.Stdout, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	editFileName := ""
	fileName, _ := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, OverrideTuningSheets)
	if !overrideNote {
		//copy file
		err := system.CopyFile(fileName, ovFileName)
		if err != nil {
			system.ErrorExit("Problems while copying '%s' to '%s' - %v", fileName, ovFileName, err)
		}
		editFileName = ovFileName
	} else {
		system.InfoLog("Note override file already exists, using file '%s' as base for editing", ovFileName)
		editFileName = ovFileName
	}

	if editor == "" {
		editor = "/usr/bin/vim" // launch vim by default
	}
	cmd := exec.Command(editor, editFileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		system.ErrorExit("Failed to start launch editor %s: %v", editor, err)
	}
	if _, ok := tuneApp.IsNoteApplied(noteID); !ok {
		system.InfoLog("Do not forget to apply the just edited Note to get your changes to take effect\n")
	} else { // noteID already applied
		system.InfoLog("Your just edited Note is already applied. To get your changes to take effect, please 'revert' the Note and apply again.\n")
	}
	// if syscall.Exec returns 'nil' the execution of the program ends immediately
	// changed syscall.Exec to exec.Command because of the new 'lock' handling
}

// NoteActionCreate helps the customer to create an own Note definition
func NoteActionCreate(noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(os.Stdout, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err == nil {
		system.ErrorExit("Note '%s' already exists. Please use 'saptune note customise %s' instead to create an override file or choose another NoteID.", noteID, noteID)
	}
	fileName := fmt.Sprintf("%s%s", NoteTuningSheets, noteID)
	if _, err := os.Stat(fileName); err == nil {
		system.ErrorExit("Note '%s' already exists in %s. Please use 'saptune note customise %s' instead to create an override file or choose another NoteID.", noteID, NoteTuningSheets, noteID)
	}
	extraFileName := fmt.Sprintf("%s%s.conf", ExtraTuningSheets, noteID)
	if _, err := os.Stat(extraFileName); err == nil {
		system.ErrorExit("Note '%s' already exists in %s. Please use 'saptune note customise %s' instead to create an override file or choose another NoteID.", noteID, ExtraTuningSheets, noteID)
	}
	//if _, err := os.Stat(extraFileName); os.IsNotExist(err) {
	//copy template file
	err := system.CopyFile(templateFile, extraFileName)
	if err != nil {
		system.ErrorExit("Problems while copying '%s' to '%s' - %v", templateFile, extraFileName, err)
	}
	if editor == "" {
		editor = "/usr/bin/vim" // launch vim by default
	}
	cmd := exec.Command(editor, extraFileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		system.ErrorExit("Failed to start launch editor %s: %v", editor, err)
	}
}

// NoteActionShow shows the content of the Note definition file
func NoteActionShow(writer io.Writer, noteID, noteTuningSheets, extraTuningSheets string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	fileName, _ := getFileName(noteID, noteTuningSheets, extraTuningSheets)
	cont, err := ioutil.ReadFile(fileName)
	if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
	}
	fmt.Fprintf(writer, "\nContent of Note %s:\n%s\n", noteID, string(cont))
}

// NoteActionDelete deletes a custom Note definition file and
// the corresponding override file
func NoteActionDelete(reader io.Reader, writer io.Writer, noteID, noteTuningSheets, extraTuningSheets, ovTuningSheets string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}

	txtConfirm := fmt.Sprintf("Do you really want to delete Note (%s)?", noteID)
	fileName, extraNote := getFileName(noteID, noteTuningSheets, extraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, ovTuningSheets)

	// check, if note is active - applied
	if _, ok := tuneApp.IsNoteApplied(noteID); ok {
		system.InfoLog("The Note definition file you want to delete is currently in use, which means it is already applied.")
		system.InfoLog("So please 'revert' the Note first and then try deleting again.\n")
		system.ErrorExit("", 0)
	}

	if !extraNote && !overrideNote {
		system.ErrorExit("ATTENTION: The Note definition file you want to delete is a saptune internal (shipped) Note and can NOT be deleted. Exiting ...")
	}
	if !extraNote && overrideNote {
		// system note, override file exists
		txtConfirm = fmt.Sprintf("Note to delete is a saptune internal (shipped) Note, so it can NOT be deleted. But an override file for the Note exists.\nDo you want to remove the override file for Note %s?", noteID)
	}
	if extraNote && overrideNote {
		// custome note with override file
		txtConfirm = fmt.Sprintf("Note to delete is a customer/vendor specific Note.\nDo you really want to delete this Note (%s) and the corresponding override file?", noteID)
	}
	if extraNote && !overrideNote {
		// custome note
		txtConfirm = fmt.Sprintf("Note to delete is a customer/vendor specific Note.\nDo you really want to delete this Note (%s)?", noteID)
	}

	if readYesNo(txtConfirm, reader, writer) {
		deleteNote(fileName, ovFileName, overrideNote, extraNote)
	}
}

// NoteActionRename renames a custom Note definition file and
// the corresponding override file
func NoteActionRename(reader io.Reader, writer io.Writer, noteID, newNoteID, noteTuningSheets, extraTuningSheets, ovTuningSheets string, tuneApp *app.App) {
	if noteID == "" || newNoteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	if _, err := tuneApp.GetNoteByID(newNoteID); err == nil {
		system.ErrorExit("The new name '%s' for Note %s already exists, can't rename.", noteID, newNoteID)
	}

	txtConfirm := fmt.Sprintf("Do you really want to rename Note %s to %s?", noteID, newNoteID)
	fileName, extraNote := getFileName(noteID, noteTuningSheets, extraTuningSheets)
	newFileName := fmt.Sprintf("%s%s.conf", extraTuningSheets, newNoteID)
	if !extraNote {
		system.ErrorExit("The Note definition file you want to rename is a saptune internal (shipped) Note and can NOT be renamed. Exiting ...")
	}
	ovFileName, overrideNote := getovFile(noteID, ovTuningSheets)
	newovFileName := fmt.Sprintf("%s%s", ovTuningSheets, newNoteID)

	// check, if note is active - applied
	if _, ok := tuneApp.IsNoteApplied(noteID); ok {
		system.InfoLog("The Note definition file you want to rename is currently in use, which means it is already applied.")
		system.InfoLog("So please 'revert' the Note first and then try renaming again.\n")
		system.ErrorExit("", 0)
	}

	if extraNote && overrideNote {
		// custome note with override file
		txtConfirm = fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) and the corresponding override file to the new name '%s'?", noteID, newNoteID)
	}
	if extraNote && !overrideNote {
		// custome note
		txtConfirm = fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) to the new name '%s'?", noteID, newNoteID)
	}

	if readYesNo(txtConfirm, reader, writer) {
		renameNote(fileName, newFileName, ovFileName, newovFileName, overrideNote, extraNote)
	}
}

// NoteActionRevert reverts all parameter settings of a Note back to the
// state before 'apply'
func NoteActionRevert(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if err := tuneApp.RevertNote(noteID, true); err != nil {
		system.ErrorExit("Failed to revert note %s: %v", noteID, err)
	}
	fmt.Fprintf(writer, "Parameters tuned by the note have been successfully reverted.\n")
}

// NoteActionEnabled lists all enabled Note definitions as list separated
// by blanks
func NoteActionEnabled(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.NoteApplyOrder) != 0 {
		fmt.Fprintf(writer, "%s", strings.Join(tuneApp.NoteApplyOrder, " "))
	}
}
