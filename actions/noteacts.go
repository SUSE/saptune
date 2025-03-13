package actions

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"sort"
	"strings"
)

var templateFile = "/usr/share/saptune/NoteTemplate.conf"

// NoteAction  Note actions like apply, revert, verify asm.
func NoteAction(writer io.Writer, actionName, noteID, newNoteID string, tuneApp *app.App) {
	switch actionName {
	case "apply":
		NoteActionApply(writer, noteID, tuneApp)
	case "list":
		NoteActionList(writer, tuneApp)
	case "verify":
		NoteActionVerify(writer, noteID, tuneApp)
	case "simulate":
		system.WarningLog("the action 'note simulate' is deprecated!.\nsaptune will still handle this action in the current version, but it will be removed in future versions of saptune.")
		NoteActionSimulate(writer, noteID, tuneApp)
	case "customise", "customize":
		NoteActionCustomise(writer, noteID, tuneApp)
	case "edit":
		NoteActionEdit(writer, noteID, tuneApp)
	case "create":
		NoteActionCreate(writer, noteID, tuneApp)
	case "show":
		NoteActionShow(writer, noteID, tuneApp)
	case "delete":
		NoteActionDelete(os.Stdin, writer, noteID, tuneApp)
	case "refresh":
		NoteActionRefresh(writer, noteID, tuneApp)
	case "rename":
		NoteActionRename(os.Stdin, writer, noteID, newNoteID, tuneApp)
	case "revert":
		NoteActionRevert(writer, noteID, tuneApp)
	case "revertall":
		RevertAction(writer, "all", tuneApp)
	case "applied":
		NoteActionApplied(writer, tuneApp)
	case "enabled":
		NoteActionEnabled(writer, tuneApp)
	default:
		PrintHelpAndExit(writer, 1)
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
			system.NoticeLog("note '%s' already applied. Nothing to do", noteID)
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
func NoteActionList(writer io.Writer, tuneApp *app.App) {
	fmt.Fprintf(writer, "\nAll notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note, C denotes custom note, D denotes deprecated notes):\n")
	format := ""
	jnoteList := []system.JNoteListEntry{}
	jnoteListEntry := system.JNoteListEntry{}

	solutionNoteIDs := tuneApp.GetSortedSolutionEnabledNotes()
	for _, noteID := range tuneApp.GetSortedAllNotes() {
		noteObj := tuneApp.AllNotes[noteID]
		// setup the list format to print
		format, jnoteListEntry = setupNoteListFormat(noteID, solutionNoteIDs, tuneApp)
		// handle special highlighting in Note description
		// like the 'only' in SAP Note 1656250
		bonly := " " + setBoldText + "only" + resetBoldText + " "
		nname := strings.Replace(noteObj.Name(), " only ", bonly, 1)
		fmt.Fprintf(writer, format, noteID, nname)
		jnoteListEntry.NoteID = noteID
		jnoteListEntry.NoteDesc, jnoteListEntry.NoteVers, jnoteListEntry.NoteRdate, jnoteListEntry.NoteRef = note.GetNoteHeadData(noteObj)
		jnoteList = append(jnoteList, jnoteListEntry)
	}
	tuneApp.PrintNoteApplyOrder(writer)
	remember := bytes.Buffer{}
	if system.GetFlagVal("format") == "json" {
		writer = &remember
	}
	rememberMessage(writer)
	result := system.JNoteList{
		NotesList:  jnoteList,
		NotesOrder: tuneApp.NoteApplyOrder,
		Msg:        remember.String(),
	}
	system.Jcollect(result)
}

// setupNoteListFormat collects needed info and setup the list format
func setupNoteListFormat(noteID string, solutionNoteIDs []string, tuneApp *app.App) (string, system.JNoteListEntry) {
	jnoteListEntry := system.JNoteListEntryInit()
	format := "\t%s\t\t%s\n"
	if len(noteID) >= 8 {
		format = "\t%s\t%s\n"
	}
	if _, err := os.Stat(fmt.Sprintf("%s%s", OverrideTuningSheets, noteID)); err == nil {
		// override file exists
		format = " O" + format
		jnoteListEntry.NoteOverride = true
	}
	if _, err := os.Stat(fmt.Sprintf("%s%s.conf", ExtraTuningSheets, noteID)); err == nil {
		// custom note
		format = " C" + format
		jnoteListEntry.CustomNote = true
	}
	if _, err := os.Stat(fmt.Sprintf("%s%s", DeprecationSheets, noteID)); err == nil {
		// deprecated note
		format = " D" + format
		jnoteListEntry.DepNote = true
	}
	if i := sort.SearchStrings(solutionNoteIDs, noteID); i < len(solutionNoteIDs) && solutionNoteIDs[i] == noteID {
		j := tuneApp.PositionInNoteApplyOrder(noteID)
		if j < 0 { // noteID was reverted manually
			format = " " + setGreenText + "-" + format + resetTextColor
			jnoteListEntry.ManReverted = true
		} else {
			format = " " + setGreenText + "*" + format + resetTextColor
			jnoteListEntry.SolEnabled = true
		}
	}
	if i := sort.SearchStrings(tuneApp.TuneForNotes, noteID); i < len(tuneApp.TuneForNotes) && tuneApp.TuneForNotes[i] == noteID {
		format = " " + setGreenText + "+" + format + resetTextColor
		jnoteListEntry.ManEnabled = true
	}
	return format, jnoteListEntry
}

// NoteActionVerify compares all parameter settings from a Note definition
// against the system settings
func NoteActionVerify(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" || noteID == "applied" {
		VerifyAllParameters(writer, tuneApp)
	} else {
		result := system.JPNotes{
			Verifications: []system.JPNotesLine{},
			Attentions:    []system.JPNotesRemind{},
			NotesOrder:    []string{},
			SysCompliance: nil,
		}

		// Check system parameters against the specified note, no matter the note has been tuned for or not.
		conforming, comparisons, _, err := tuneApp.VerifyNote(noteID)
		if err != nil {
			system.Jcollect(result)
			system.ErrorExit("Failed to test the current system against the specified note: %v", err)
		}
		noteComp := make(map[string]map[string]note.FieldComparison)
		noteComp[noteID] = comparisons
		PrintNoteFields(writer, "HEAD", noteComp, true, &result)
		tuneApp.PrintNoteApplyOrder(writer)
		result.NotesOrder = tuneApp.NoteApplyOrder
		result.SysCompliance = &conforming
		system.Jcollect(result)
		if !conforming {
			system.ErrorExit("The parameters listed above have deviated from the specified note.\n", "colorPrint", setRedText, setBoldText, resetBoldText, resetTextColor)
		} else {
			fmt.Fprintf(writer, "%s%sThe system fully conforms to the specified note.%s%s\n", setGreenText, setBoldText, resetBoldText, resetTextColor)
		}
	}
}

// NoteActionSimulate shows all changes that will be applied to the system if
// the Note will be applied.
func NoteActionSimulate(writer io.Writer, noteID string, tuneApp *app.App) {
	result := system.JPNotes{}
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
		PrintNoteFields(writer, "HEAD", noteComp, false, &result)
		result.SysCompliance = nil
		system.Jcollect(result)
	}
}

// NoteActionCustomise creates an override file and allows to editing the Note
// definition override file
func NoteActionCustomise(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	editSrcFile := ""
	editDestFile := ""
	fileName, _ := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, OverrideTuningSheets)
	if !overrideNote {
		editSrcFile = fileName
		editDestFile = ovFileName
	} else {
		system.NoticeLog("Note override file already exists, using file '%s' as base for editing", ovFileName)
		editSrcFile = ovFileName
		editDestFile = ovFileName
	}

	changed, err := system.EditAndCheckFile(editSrcFile, editDestFile, noteID, "note")
	if err != nil {
		system.ErrorExit("Problems while editing note definition file '%s' - %v", editSrcFile, err)
	}
	if changed {
		if _, ok := tuneApp.IsNoteApplied(noteID); !ok {
			system.NoticeLog("Do not forget to apply the just edited Note to get your changes to take effect\n")
		} else { // noteID already applied
			system.NoticeLog("Your just edited Note is already applied. To get your changes to take effect, please 'revert' the Note and apply again.\n")
		}
	} else {
		system.NoticeLog("Nothing changed during the editor session, so no update of the note definition file '%s'", editSrcFile)
	}
}

// NoteActionEdit allows to editing the custom/vendor specific Note definition
// file and NOT the override file
func NoteActionEdit(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	fileName, extraNote := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, OverrideTuningSheets)
	if !extraNote {
		system.ErrorExit("The Note definition file you want to edit is a saptune internal (shipped) Note and can NOT be edited. Use 'saptune note customise' instead. Exiting ...")
	}

	changed, err := system.EditAndCheckFile(fileName, fileName, noteID, "note")
	if err != nil {
		system.ErrorExit("Problems while editing Note definition file '%s' - %v", fileName, err)
	}
	if changed {
		if _, ok := tuneApp.IsNoteApplied(noteID); !ok {
			system.NoticeLog("Do not forget to apply the just edited Note to get your changes to take effect\n")
		} else { // noteID already applied
			system.NoticeLog("Your just edited Note is already applied. To get your changes to take effect, please 'revert' the Note and apply again.\n")
		}
		if overrideNote {
			system.NoticeLog("Note override file '%s' exists. Please check, if the content of this file is still valid", ovFileName)
		}

	} else {
		system.NoticeLog("Nothing changed during the editor session, so no update of the note definition file '%s'", fileName)
	}
}

// NoteActionCreate helps the customer to create an own Note definition
func NoteActionCreate(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
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
		system.ErrorExit("Note '%s' already exists in %s. Please use 'saptune note edit %s' instead to modify this custom specific Note or 'saptune note customise %s' to create an override file or choose another NoteID.", noteID, ExtraTuningSheets, noteID)
	}

	changed, err := system.EditAndCheckFile(templateFile, extraFileName, noteID, "note")
	if err != nil {
		system.ErrorExit("Problems while editing note definition file '%s' - %v", extraFileName, err)
	}
	if !changed {
		system.NoticeLog("Nothing changed during the editor session, so no new, custom specific note definition file will be created.")
	} else {
		system.NoticeLog("Note '%s' created successfully. You can modify the content of your Note definition file by using 'saptune note edit %s' or create an override file by 'saptune note customise %s'.", noteID, noteID, noteID)
	}
}

// NoteActionShow shows the content of the Note definition file
func NoteActionShow(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	fileName, _ := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	cont, err := os.ReadFile(fileName)
	if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
	}
	fmt.Fprintf(writer, "\nContent of Note %s:\n%s\n", noteID, string(cont))
}

// NoteActionDelete deletes a custom Note definition file and
// the corresponding override file
func NoteActionDelete(reader io.Reader, writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}

	txtConfirm := fmt.Sprintf("Do you really want to delete Note (%s)?", noteID)
	fileName, extraNote := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, OverrideTuningSheets)

	// check, if note is active - applied
	if _, ok := tuneApp.IsNoteApplied(noteID); ok {
		system.ErrorExit("The Note definition file you want to delete is currently in use, which means it is already applied.\nSo please 'revert' the Note first and then try deleting again.")
	}

	if !extraNote && !overrideNote {
		system.ErrorExit("The Note definition file you want to delete is a saptune internal (shipped) Note and can NOT be deleted. Exiting ...")
	}

	if !extraNote && overrideNote {
		// system note, override file exists
		txtConfirm = fmt.Sprintf("Note to delete is a saptune internal (shipped) Note, so it can NOT be deleted. But an override file for the Note exists.\nDo you want to remove the override file for Note %s?", noteID)
	}
	if extraNote && overrideNote {
		// custom note with override file
		txtConfirm = fmt.Sprintf("Note to delete is a customer/vendor specific Note and an override file for the Note exists.\nDo you want to remove the override file for Note %s?", noteID)
	}
	if overrideNote {
		// remove override file
		if readYesNo(txtConfirm, reader, writer) {
			deleteDefFile(ovFileName)
		}
	}
	if extraNote {
		inSolution, _ := getNoteInSol(tuneApp, noteID)
		if inSolution != "" {
			system.ErrorExit("The Note definition file you want to delete is part of a Solution(s) (%s). Please fix the Solution first and then try deleting again.", inSolution)
		}
		// custom note
		txtConfirm = fmt.Sprintf("Note to delete is a customer/vendor specific Note.\nDo you really want to delete this Note (%s)?", noteID)
		// remove customer/vendor specific note definition file
		if readYesNo(txtConfirm, reader, writer) {
			deleteDefFile(fileName)
		}
	}
}

// NoteActionRename renames a custom Note definition file and
// the corresponding override file
func NoteActionRename(reader io.Reader, writer io.Writer, noteID, newNoteID string, tuneApp *app.App) {
	if noteID == "" || newNoteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		system.ErrorExit("%v", err)
	}
	if _, err := tuneApp.GetNoteByID(newNoteID); err == nil {
		system.ErrorExit("The new name '%s' for Note '%s' already exists, can't rename.", newNoteID, noteID)
	}

	txtConfirm := fmt.Sprintf("Do you really want to rename Note '%s' to '%s'?", noteID, newNoteID)
	fileName, extraNote := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	newFileName := fmt.Sprintf("%s%s.conf", ExtraTuningSheets, newNoteID)
	if !extraNote {
		system.ErrorExit("The Note definition file you want to rename is a saptune internal (shipped) Note and can NOT be renamed. Exiting ...")
	}
	ovFileName, overrideNote := getovFile(noteID, OverrideTuningSheets)
	newovFileName := fmt.Sprintf("%s%s", OverrideTuningSheets, newNoteID)

	// check, if note is active - applied
	if _, ok := tuneApp.IsNoteApplied(noteID); ok {
		system.ErrorExit("The Note definition file you want to rename is currently in use, which means it is already applied.\nSo please 'revert' the Note first and then try renaming again.")
	}
	inSolution, _ := getNoteInSol(tuneApp, noteID)
	if inSolution != "" {
		system.ErrorExit("The Note definition file you want to rename is part of a Solution(s) (%s). Please fix the Solution first and then try renaming again.", inSolution)
	}

	if extraNote && overrideNote {
		// custom note with override file
		txtConfirm = fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) and the corresponding override file to the new name '%s'?", noteID, newNoteID)
	}
	if extraNote && !overrideNote {
		// custom note
		txtConfirm = fmt.Sprintf("Note to rename is a customer/vendor specific Note.\nDo you really want to rename this Note (%s) to the new name '%s'?", noteID, newNoteID)
	}

	if readYesNo(txtConfirm, reader, writer) {
		renameDefFile(fileName, newFileName)
		if overrideNote {
			renameDefFile(ovFileName, newovFileName)
		}
	}
}

// NoteActionRevert reverts all parameter settings of a Note back to the
// state before 'apply'
func NoteActionRevert(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(writer, 1)
	}
	// 'ok' only used to control the log messages
	// call RevertNote in any case to get the chance of clean up
	_, ok := tuneApp.IsNoteApplied(noteID)
	if err := tuneApp.RevertNote(noteID, true); err != nil {
		system.ErrorExit("Failed to revert note %s: %v", noteID, err)
	}
	// if a solution is enabled (available in the configuration), check, if
	// this note is the last note in NoteApplyOrder, which is related to
	// this solution. If yes, remove solution for the configuration.
	solutionStillEnabled(tuneApp)

	if ok {
		system.InfoLog("Parameters tuned by the note '%s' have been successfully reverted.", noteID)
		fmt.Fprintf(writer, "Parameters tuned by the note have been successfully reverted.\n")
	} else {
		system.NoticeLog("Note '%s' is not applied, so nothing to revert.", noteID)
	}
}

// if a solution is enabled (available in the configuration), check, if
// there is a least one note in NoteApplyOrder, which is related to
// this solution. If no, remove solution for the configuration.
func solutionStillEnabled(tuneApp *app.App) {
	if len(tuneApp.TuneForSolutions) == 0 {
		return
	}
	for _, sol := range tuneApp.TuneForSolutions {
		solNoteAvail := false
		for _, solNote := range tuneApp.AllSolutions[sol] {
			if tuneApp.PositionInNoteApplyOrder(solNote) < 0 {
				continue
			} else {
				solNoteAvail = true // sol still valid
				break
			}
		}
		if !solNoteAvail {
			system.InfoLog("The last, still enabled Note got reverted and removed from the configuration, so remove the enabled Solution from the configuration too.")
			_ = tuneApp.RemoveSolFromConfig(sol)
		}
	}
}

// NoteActionEnabled lists all enabled Note definitions as list separated
// by blanks
func NoteActionEnabled(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.NoteApplyOrder) != 0 {
		fmt.Fprintf(writer, "%s", strings.Join(tuneApp.NoteApplyOrder, " "))
	}
	system.Jcollect(tuneApp.NoteApplyOrder)
}

// NoteActionApplied lists all applied Note definitions as list separated
// by blanks
func NoteActionApplied(writer io.Writer, tuneApp *app.App) {
	notesApplied := tuneApp.AppliedNotes()
	fmt.Fprintf(writer, "%s", notesApplied)
	system.Jcollect(strings.Split(notesApplied, " "))
}

// NoteActionRefresh re-applies Note parameter settings to the system
// if an already applied Note got changed.
func NoteActionRefresh(writer io.Writer, noteID string, tuneApp *app.App) {
	errCount := 0
	noteList := make([]string, 0)
	if noteID == "" || noteID == "applied" {
		if len(tuneApp.NoteApplyOrder) == 0 {
			system.NoticeLog("No notes enabled, nothing to refresh.\n")
			system.ErrorExit("", 0)
		}
		noteList = tuneApp.NoteApplyOrder
	} else {
		noteList = append(noteList, noteID)
	}
	for _, note := range noteList {
		if err := tuneApp.RefreshNote(note); err != nil {
			errCount++
			system.ErrorLog("Failed to refresh tuning for note '%s': %v", note, err)
		} else {
			system.NoticeLog("The note '%s' has been refreshed successfully.\n", note)
		}
	}
	if errCount != 0 {
		system.ErrorExit("At least the refresh of the tuning of one Notes was not successful. Please check.", 1)
	}
	rememberMessage(writer)
}
