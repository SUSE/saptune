package app

import (
	"fmt"
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
)

// PrintNoteApplyOrder prints out the order of the currently enabled notes
func (app *App) PrintNoteApplyOrder(writer io.Writer) {
	if len(app.NoteApplyOrder) != 0 {
		fmt.Fprintf(writer, "\ncurrent order of enabled notes is: %s\n\n", strings.Join(app.NoteApplyOrder, " "))
	}
}

// AppliedNotes returns the currently applied Notes in 'NoteApplyOrder' order
func (app *App) AppliedNotes() string {
	var notesApplied string
	for _, note := range app.NoteApplyOrder {
		if _, ok := app.IsNoteApplied(note); ok {
			notesApplied = fmt.Sprintf("%s%s ", notesApplied, note)
		}
	}
	notesApplied = strings.TrimSpace(notesApplied)
	return notesApplied
}

// PositionInNoteApplyOrder returns the position of the note within the slice.
// for a given noteID get the position in the slice NoteApplyOrder
// do not sort the slice
func (app *App) PositionInNoteApplyOrder(noteID string) int {
	for cnt, note := range app.NoteApplyOrder {
		if note == noteID {
			return cnt
		}
	}
	return -1 //not found
}

// GetSortedSolutionEnabledNotes returns the number of all solution-enabled
// SAP notes, sorted.
func (app *App) GetSortedSolutionEnabledNotes() (allNoteIDs []string) {
	allNoteIDs = make([]string, 0)
	for _, sol := range app.TuneForSolutions {
		for _, noteID := range app.AllSolutions[sol] {
			if i := sort.SearchStrings(allNoteIDs, noteID); !(i < len(allNoteIDs) && allNoteIDs[i] == noteID) {
				allNoteIDs = append(allNoteIDs, noteID)
				sort.Strings(allNoteIDs)
			}
		}
	}
	return
}

// GetSortedAllNotes returns all SAP notes, sorted.
func (app *App) GetSortedAllNotes() []string {
	allNoteIDs := make([]string, 0)
	for noteID := range app.AllNotes {
		allNoteIDs = append(allNoteIDs, noteID)
	}
	sort.Strings(allNoteIDs)
	return allNoteIDs
}

// IsNoteApplied checks, if a note is applied or not
// return true, if note is already applied
// return false, if not is NOT applied
func (app *App) IsNoteApplied(noteID string) (string, bool) {
	rval := ""
	ret := false
	// check against state file first is still ok, don't change
	// because of TuneAll/RevertAll to cover system reboot
	// NoteApplyOrder is filled, but state files are removed
	sfile, err := os.Stat(app.State.GetPathToNote(noteID))
	if err == nil {
		// state file for note already exists
		// check, if note is part of NOTE_APPLY_ORDER
		if app.PositionInNoteApplyOrder(noteID) < 0 { // noteID not yet available
			// bsc#1167618
			// check, if state file is empty - seems to be a
			// left-over of the update from saptune V1 to V2
			if sfile.Size() == 0 {
				// remove old, left-over state file and go
				// forward to apply the note
				os.Remove(app.State.GetPathToNote(noteID))
			} else {
				// data mismatch, do not apply the note
				system.WarningLog("note '%s' is not listed in 'NOTE_APPLY_ORDER', but a non-empty state file exists. To prevent configuration mismatch, please revert note '%s' first and try again.", noteID, noteID)
				rval = "mismatch"
				ret = true
			}
		} else { // note applied
			ret = true
		}
	}
	return rval, ret
}

// NoteSanityCheck checks, if for all notes listed in
// NoteApplyOrder and TuneForNotes a note definition file exists.
// if not, remove the NoteID from the variables, save the new config and
// inform the user
func (app *App) NoteSanityCheck() error {
	// app.NoteApplyOrder, app.TuneForNotes
	errs := make([]error, 0)
	for _, note := range app.NoteApplyOrder {
		if _, exists := app.AllNotes[note]; exists {
			// note definition file for NoteID exists
			continue
		}
		// bsc#1149205
		// noteID available in apply order list, but no note definition
		// file found. May be removed or renamed.
		system.ErrorLog("The Note ID '%s' is not recognized by saptune, but it is listed in the apply order list.\nMay be the associated Note definition file was removed or renamed via command line without previously reverting the Note.\nSaptune will now remove the NoteID from the apply order list to prevent further confusion.", note)
		app.removeFromConfig(note)
		if err := app.SaveConfig(); err != nil {
			errs = append(errs, err)
		}

		// idea to first check for existence of section file and then
		// check the state file will NOT work in case that the apply
		// was done with a previous saptune version where NO section
		// file handling exists
		fileName := fmt.Sprintf("/run/saptune/sections/%s.sections", note)
		// check, if empty state file exists
		if content, err := os.ReadFile(app.State.GetPathToNote(note)); err == nil && len(content) == 0 {
			// remove empty state file
			_ = app.State.Remove(note)
			if _, err := os.Stat(fileName); err == nil {
				// section file exists, remove
				_ = os.Remove(fileName)
			}
		} else if err == nil {
			// non-empty state file
			if _, err := os.Stat(fileName); err == nil {
				// section file exists, try revert
				// without section file a revert is
				// impossible as the fall back, the
				// Note definition file no longer exists
				_ = app.RevertNote(note, true)
			} else {
				// non empty state file, but no chance to revert
				_ = app.State.Remove(note)
			}
		} else if _, err := os.Stat(fileName); err == nil {
			// no state file, but section file exists, remove
			_ = os.Remove(fileName)
		}
	}

	err := sap.PrintErrors(errs)
	return err
}

// GetNoteByID return the note corresponding to the number, or an error
// if the note does not exist.
func (app *App) GetNoteByID(id string) (note.Note, error) {
	if n, exists := app.AllNotes[id]; exists {
		return n, nil
	}
	return nil, fmt.Errorf(`the Note ID "%s" is not recognised by saptune.
Run "saptune note list" for a complete list of supported notes.
and then please double check your input and /etc/sysconfig/saptune`, id)
}

// TuneNote apply tuning for a note.
// If the note is not yet covered by one of the enabled solutions,
// the note number will be added into the list of additional notes.
func (app *App) TuneNote(noteID string) error {
	savConf := false
	aNote, err := app.GetNoteByID(noteID)
	if err != nil {
		return err
	}
	solNotes := app.GetSortedSolutionEnabledNotes()
	searchInSol := sort.SearchStrings(solNotes, noteID)
	searchInNote := sort.SearchStrings(app.TuneForNotes, noteID)
	if !(searchInSol < len(solNotes) && solNotes[searchInSol] == noteID) && !(searchInNote < len(app.TuneForNotes) && app.TuneForNotes[searchInNote] == noteID) {
		// Note is not covered by any of the existing solution, hence adding it into the additions' list
		app.TuneForNotes = append(app.TuneForNotes, noteID)
		sort.Strings(app.TuneForNotes)
		savConf = true
	}
	// to prevent double noteIDs in the apply order list
	i := app.PositionInNoteApplyOrder(noteID)
	if i < 0 { // noteID not yet available
		app.NoteApplyOrder = append(app.NoteApplyOrder, noteID)
		savConf = true
	}
	if savConf {
		if err := app.SaveConfig(); err != nil {
			return err
		}
	}

	// check, if system already complies with the requirements.
	// set values for later use
	conforming, _, valApplyList, err := app.VerifyNote(noteID)
	if err != nil {
		return err
	}

	// Save current state for the Note in any case
	currentState, err := aNote.Initialise()
	if err != nil {
		system.ErrorLog("Failed to examine system for the current status of note %s - %v", noteID, err)
		return err
	}
	forceApply := handleCounterParts(currentState)
	if err = app.State.Store(noteID, currentState, false); err != nil {
		system.ErrorLog("Failed to save current state of note %s - %v", noteID, err)
		return err
	}

	optimised, err := currentState.Optimise()
	if err != nil {
		system.ErrorLog("Failed to calculate optimised parameters for note %s - %v", noteID, err)
		return err
	}
	if len(valApplyList) != 0 {
		optimised = optimised.(note.INISettings).SetValuesToApply(valApplyList)
	}

	if conforming && !forceApply {
		// Do not apply the Note, if the system already complies with
		// the requirements.
		return nil
	}
	if err := optimised.Apply(); err != nil {
		system.ErrorLog("Failed to apply note %s - %v", noteID, err)
		return err
	}

	return nil
}

// RevertNote revert parameters tuned by the note and clear its stored states.
func (app *App) RevertNote(noteID string, permanent bool) error {

	noteTemplate, err := app.GetNoteByID(noteID)
	if err != nil {
		// to revert an applied note even if the corresponding
		// note definition file is no longer available, but the
		// saved state info can be found
		// helpful for cleanup
		noteTemplate = note.INISettings{
			ConfFilePath:    "",
			ID:              "",
			DescriptiveName: "",
		}
	}

	// Remove from configuration
	if permanent {
		app.removeFromConfig(noteID)
		if err := app.SaveConfig(); err != nil {
			return err
		}
	}

	// Revert parameters using the file record
	// Workaround for Go JSON package's stubbornness, Go developers are not willing to fix their code in this occasion.
	var noteReflectValue = reflect.New(reflect.TypeOf(noteTemplate))
	var noteIface interface{} = noteReflectValue.Interface()
	if err := app.State.Retrieve(noteID, &noteIface); err == nil {
		//var noteRecovered note.Note = noteIface.(note.Note)
		var noteRecovered = noteIface.(note.Note)
		if reflect.TypeOf(noteRecovered).String() == "*note.INISettings" {
			noteRecovered = noteRecovered.(*note.INISettings).SetValuesToApply([]string{"revert"})
		}

		if err := noteRecovered.Apply(); err != nil {
			return err
		} else if err := app.State.Remove(noteID); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// VerifyNote inspect the system and verify that all parameters conform
// to the note's guidelines.
// The note comparison results will always contain all fields, no matter
// the note is currently conforming or not.
func (app *App) VerifyNote(noteID string) (conforming bool, comparisons map[string]note.FieldComparison, valApplyList []string, err error) {
	theNote, err := app.GetNoteByID(noteID)
	if err != nil {
		return
	}
	if reflect.TypeOf(theNote).String() == "note.INISettings" {
		// workaround to prevent storing of parameter state files
		// during verify
		theNote = theNote.(note.INISettings).SetValuesToApply([]string{"verify"})
	}
	// Run optimisation routine and compare it against current status
	inspectedNote, err := theNote.Initialise()
	if err != nil {
		return false, nil, nil, err
	}

	// to get Apply work:
	optimisedNote, err := theNote.Initialise()
	if err != nil {
		return false, nil, nil, err
	}
	// if used inspectedNote as before, inspectedNote and optimisedNote
	// will have the same content after 'Optimise()'
	// so CompareNoteFields wont find a difference and NO Apply will done
	//optimisedNote, err := inspectedNote.Optimise()
	optimisedNote, err = optimisedNote.Optimise()
	if err != nil {
		return false, nil, nil, err
	}
	if reflect.TypeOf(theNote).String() == "note.INISettings" {
		// remove workaround to not affect the 'comparison' result
		inspectedNote = inspectedNote.(note.INISettings).SetValuesToApply(make([]string, 0))
		optimisedNote = optimisedNote.(note.INISettings).SetValuesToApply(make([]string, 0))
	}
	conforming, comparisons, valApplyList = note.CompareNoteFields(inspectedNote, optimisedNote)
	return
}

// RefreshNote re-applies parameter values, if a note definition file or
// an override file has changed
func (app *App) RefreshNote(noteID string) error {
	paramApplyList := []string{}
	// check, if note is already applied
	// if not, apply note instead of refresh
	ok, err := app.checkNoteAppliedState(noteID)
	if !ok || err != nil {
		// note not applied before, but apply worked without error.
		// err is nil
		// or
		// error during check (e.g. during apply of note)
		return err
	}
	// note already applied, starting refresh
	system.NoticeLog("note '%s' already applied, refreshing.", noteID)

	// get parameters and values, which differ from the current system
	// settings, needed for Apply later.
	// ValuesToApply created by 'verify' may miss changed or added parameter
	// because they simply conforms to the current system. And it can
	// contain more parameter than the 'changedParameter' in the note,
	// because other notes later applied may have changed additional
	// parameter values.
	// But we need only the really 'changedParameter' per note, so no
	// 'ValuesToApply' from 'verify' needed.
	//
	// But comparisons will contain the optimized values for the parameters
	// which are the values to change in the parameter file.
	// taking 'override' and 'untouched' into account
	_, comparisons, _, _ := app.VerifyNote(noteID)

	//fileName := reflect.ValueOf(app.AllNotes[noteID]).FieldByName("ConfFilePath").String()
	fileName := app.AllNotes[noteID].(note.INISettings).ConfFilePath

	// Because parameter of the note file could be changed or added,
	// but conforms to the current system, so not listed in ValuesToApply
	// of 'verify'.
	// So first get changed parameter from note definition file and/or
	// override file, later create ValuesToApply.
	// Build a list of parameter, flags (changed in note, changed in
	// override, added, deleted, ..) and values
	changedParameter := collectChangedParameterInfo(noteID, fileName, comparisons, app)
	system.DebugLog("'%d' parameter changed in note '%s'", len(changedParameter), noteID)
	system.DebugLog("changed parameter are - '%+v'\n", changedParameter)
	if len(changedParameter) == 0 {
		system.NoticeLog("Nothing changed in note '%s', nothing to do.", noteID)
		return nil
	}

	// adjust the related state files (section, saved_state, parameter)
	// and build the 'ApplyList'
	paramApplyList, err = adjustStateFiles(noteID, app, changedParameter, comparisons)
	if err != nil {
		return err
	}

	// prepare apply
	refreshed, err := note.GetVendInfo("vend", noteID)
	if len(paramApplyList) != 0 {
		refreshed = refreshed.(note.INISettings).SetValuesToApply(paramApplyList)
	}
	// apply changed parameter values
	if err := refreshed.Apply(); err != nil {
		return err
	}

	return err
}

// check, if note is already applied
// if not call 'note apply' and exit
func (app *App) checkNoteAppliedState(noteID string) (bool, error) {
	var err error
	str, ok := app.IsNoteApplied(noteID)
	if !ok {
		// note not applied
		system.NoticeLog("note '%s' not yet applied, redirecting to 'saptune note apply'", noteID)
		err = app.TuneNote(noteID)
		if err != nil {
			system.ErrorLog("Failed to tune for note %s: %v", noteID, err)
		}
	} else if str != "" {
		// mismatch, do not apply or re-apply
		err = fmt.Errorf(`configuration mismatch detected. No apply or re-apply/refresh of note '%s'`, noteID)
	}
	return ok, err
}
