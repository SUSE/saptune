package app

import (
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
)

// define saptunes main configuration file and variables
const (
	SysconfigSaptuneFile = "/etc/sysconfig/saptune"
	TuneForSolutionsKey  = "TUNE_FOR_SOLUTIONS"
	TuneForNotesKey      = "TUNE_FOR_NOTES"
	NoteApplyOrderKey    = "NOTE_APPLY_ORDER"
)

// App defines the application configuration and serialised state information.
type App struct {
	SysconfigPrefix  string
	AllNotes         map[string]note.Note         // all notes
	AllSolutions     map[string]solution.Solution // all solutions
	TuneForSolutions []string                     // list of solution names to tune, must always be sorted in ascending order.
	TuneForNotes     []string                     // list of additional notes to tune, must always be sorted in ascending order.
	NoteApplyOrder   []string                     // list of notes in applied order. Do NOT sort.
	State            *State                       // examine and manage serialised notes.
}

// InitialiseApp load application configuration. Panic on error.
func InitialiseApp(sysconfigPrefix, stateDirPrefix string, allNotes map[string]note.Note, allSolutions map[string]solution.Solution) (app *App) {
	app = &App{
		SysconfigPrefix: sysconfigPrefix,
		State:           &State{StateDirPrefix: stateDirPrefix},
		AllNotes:        allNotes,
		AllSolutions:    allSolutions,
	}
	sysconf, err := txtparser.ParseSysconfigFile(path.Join(app.SysconfigPrefix, SysconfigSaptuneFile), true)
	if err == nil {
		app.TuneForSolutions = sysconf.GetStringArray(TuneForSolutionsKey, []string{})
		app.TuneForNotes = sysconf.GetStringArray(TuneForNotesKey, []string{})
		app.NoteApplyOrder = sysconf.GetStringArray(NoteApplyOrderKey, []string{})
	} else {
		app.TuneForSolutions = []string{}
		app.TuneForNotes = []string{}
		app.NoteApplyOrder = []string{}
	}
	sort.Strings(app.TuneForSolutions)
	sort.Strings(app.TuneForNotes)
	return
}

// PrintNoteApplyOrder prints out the order of the currently applied notes
func (app *App) PrintNoteApplyOrder(writer io.Writer) {
	if len(app.NoteApplyOrder) != 0 {
		fmt.Fprintf(writer, "\ncurrent order of applied notes is: %s\n\n", strings.Join(app.NoteApplyOrder, " "))
	}
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

// SaveConfig save configuration to file /etc/sysconfig/saptune.
func (app *App) SaveConfig() error {
	sysconf, err := txtparser.ParseSysconfigFile(path.Join(app.SysconfigPrefix, SysconfigSaptuneFile), true)
	if err != nil {
		return err
	}
	sysconf.SetStrArray(TuneForSolutionsKey, app.TuneForSolutions)
	sysconf.SetStrArray(TuneForNotesKey, app.TuneForNotes)
	sysconf.SetStrArray(NoteApplyOrderKey, app.NoteApplyOrder)
	return ioutil.WriteFile(path.Join(app.SysconfigPrefix, SysconfigSaptuneFile), []byte(sysconf.ToText()), 0644)
}

// GetSortedSolutionEnabledNotes returns the number of all solution-enabled
// SAP notes, sorted.
func (app *App) GetSortedSolutionEnabledNotes() (allNoteIDs []string) {
	allNoteIDs = make([]string, 0, 0)
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

// GetSolutionByName return the solution corresponding to the name,
// or an error if it does not exist.
func (app *App) GetSolutionByName(name string) (solution.Solution, error) {
	if n, exists := app.AllSolutions[name]; exists {
		return n, nil
	}
	return nil, fmt.Errorf(`solution name "%s" is not recognised by saptune.
Run "saptune solution list" for a complete list of supported solutions,
and then please double check your input and /etc/sysconfig/saptune`, name)
}

// TuneNote apply tuning for a note.
// If the note is not yet covered by one of the enabled solutions,
// the note number will be added into the list of additional notes.
func (app *App) TuneNote(noteID string) error {
	forceApply := false
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
	}
	// to prevent double noteIDs in the apply order list
	i := app.PositionInNoteApplyOrder(noteID)
	if i < 0 { // noteID not yet available
		app.NoteApplyOrder = append(app.NoteApplyOrder, noteID)
	}
	if err := app.SaveConfig(); err != nil {
		return err
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
		return fmt.Errorf("Failed to examine system for the current status of note %s - %v", noteID, err)
	}
	if reflect.TypeOf(currentState).String() == "note.INISettings" {
		// in case of vm.dirty parameters save additionally the
		// counterpart values to be able to revert the values
		addkey := ""
		_, exist := reflect.TypeOf(currentState).FieldByName("SysctlParams")
		if exist {
			for _, mkey := range reflect.ValueOf(currentState).FieldByName("SysctlParams").MapKeys() {
				switch mkey.String() {
				case "vm.dirty_background_bytes":
					addkey = "vm.dirty_background_ratio"
				case "vm.dirty_bytes":
					addkey = "vm.dirty_ratio"
				case "vm.dirty_background_ratio":
					addkey = "vm.dirty_background_bytes"
				case "vm.dirty_ratio":
					addkey = "vm.dirty_bytes"
				case "force_latency":
					forceApply = true
				}
				if addkey != "" {
					//currentState.(note.INISettings).SysctlParams[addkey], _ = system.GetSysctlString(addkey)
					addkeyval, _ := system.GetSysctlString(addkey)
					//func (v Value) SetMapIndex(key, val Value)
					reflect.ValueOf(currentState).FieldByName("SysctlParams").SetMapIndex(reflect.ValueOf(addkey), reflect.ValueOf(addkeyval))
				}
			}
		}
	}
	if err = app.State.Store(noteID, currentState, false); err != nil {
		return fmt.Errorf("Failed to save current state of note %s - %v", noteID, err)
	}

	optimised, err := currentState.Optimise()
	if err != nil {
		return fmt.Errorf("Failed to calculate optimised parameters for note %s - %v", noteID, err)
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
		return fmt.Errorf("Failed to apply note %s - %v", noteID, err)
	}

	return nil
}

// TuneSolution apply tuning for a solution.
// If the solution is not yet enabled, the name will be added into the list
// of tuned solution names.
// If the solution covers any of the additional notes, those notes will be removed.
func (app *App) TuneSolution(solName string) (removedExplicitNotes []string, err error) {
	removedExplicitNotes = make([]string, 0, 0)
	sol, err := app.GetSolutionByName(solName)
	if err != nil {
		return
	}
	if i := sort.SearchStrings(app.TuneForSolutions, solName); !(i < len(app.TuneForSolutions) && app.TuneForSolutions[i] == solName) {
		app.TuneForSolutions = append(app.TuneForSolutions, solName)
		sort.Strings(app.TuneForSolutions)
		if err = app.SaveConfig(); err != nil {
			return
		}
	}
	for _, noteID := range sol {
		// Remove solution's notes from additional notes list.
		if i := sort.SearchStrings(app.TuneForNotes, noteID); i < len(app.TuneForNotes) && app.TuneForNotes[i] == noteID {
			app.TuneForNotes = append(app.TuneForNotes[0:i], app.TuneForNotes[i+1:]...)
			removedExplicitNotes = append(removedExplicitNotes, noteID)
			if err = app.SaveConfig(); err != nil {
				return
			}
		}
		if err = app.TuneNote(noteID); err != nil {
			return
		}
	}
	return
}

// TuneAll tune for all currently enabled solutions and notes.
func (app *App) TuneAll() error {
	for _, noteID := range app.NoteApplyOrder {
		if _, err := app.GetNoteByID(noteID); err != nil {
			_ = system.ErrorLog(err.Error())
			continue
		}
		if err := app.TuneNote(noteID); err != nil {
			return err
		}
	}
	return nil
}

// RevertNote revert parameters tuned by the note and clear its stored states.
func (app *App) RevertNote(noteID string, permanent bool) error {
	noteTemplate, err := app.GetNoteByID(noteID)
	if err != nil {
		return err
	}

	// Remove from configuration
	if permanent {
		i := sort.SearchStrings(app.TuneForNotes, noteID)
		if i < len(app.TuneForNotes) && app.TuneForNotes[i] == noteID {
			app.TuneForNotes = append(app.TuneForNotes[0:i], app.TuneForNotes[i+1:]...)
		}
		i = app.PositionInNoteApplyOrder(noteID)
		if i < 0 {
			system.WarningLog("noteID '%s' not found in configuration 'NoteApplyOrder'", noteID)
		} else if i < len(app.NoteApplyOrder) && app.NoteApplyOrder[i] == noteID {
			// remove noteID from the configuration 'NoteApplyOrder'
			app.NoteApplyOrder = append(app.NoteApplyOrder[0:i], app.NoteApplyOrder[i+1:]...)
		}
		if err := app.SaveConfig(); err != nil {
			return err
		}
	}

	// Revert parameters using the file record
	// Workaround for Go JSON package's stubbornness, Go developers are not willing to fix their code in this occasion.
	var noteReflectValue = reflect.New(reflect.TypeOf(noteTemplate))
	var noteIface interface{} = noteReflectValue.Interface()
	if err = app.State.Retrieve(noteID, &noteIface); err == nil {
		var noteRecovered note.Note = noteIface.(note.Note)
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

// RevertSolution permanently revert notes tuned by the solution and
// clear their stored states.
func (app *App) RevertSolution(solName string) error {
	sol, err := app.GetSolutionByName(solName)
	if err != nil {
		return err
	}
	// Remove from configuration
	i := sort.SearchStrings(app.TuneForSolutions, solName)
	if i < len(app.TuneForSolutions) && app.TuneForSolutions[i] == solName {
		app.TuneForSolutions = append(app.TuneForSolutions[0:i], app.TuneForSolutions[i+1:]...)
		if err := app.SaveConfig(); err != nil {
			return err
		}
	}
	// The tricky part: figure out which notes are to be reverted, do not revert manually enabled notes.
	notesDoNotRevert := make(map[string]struct{})
	for _, noteID := range app.TuneForNotes {
		notesDoNotRevert[noteID] = struct{}{}
	}
	// Do not revert notes that are referred to by other enabled solutions
	for _, otherSolName := range app.TuneForSolutions {
		if otherSolName != solName {
			otherSolNotes, err := app.GetSolutionByName(otherSolName)
			if err != nil {
				return err
			}
			for _, noteID := range otherSolNotes {
				notesDoNotRevert[noteID] = struct{}{}
			}
		}
	}
	// Now revert the (sol notes - manually enabled - other sol notes)
	noteErrs := make([]error, 0, 0)
	for _, noteID := range sol {
		if _, found := notesDoNotRevert[noteID]; found {
			continue // skip this one
		}
		if err := app.RevertNote(noteID, true); err != nil {
			if err != nil {
				noteErrs = append(noteErrs, err)
			}
		}
	}
	if len(noteErrs) == 0 {
		return nil
	}
	return fmt.Errorf("Failed to revert one or more SAP notes that belong to the solution: %v", noteErrs)
}

// RevertAll revert all tuned parameters (both solutions and additional notes),
// and clear stored states.
func (app *App) RevertAll(permanent bool) error {
	allErrs := make([]error, 0, 0)

	// Simply revert all notes from serialised states
	otherNotes, err := app.State.List()
	if err == nil {
		for _, otherNoteID := range otherNotes {
			if err := app.RevertNote(otherNoteID, permanent); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	} else {
		allErrs = append(allErrs, err)
	}
	if permanent {
		app.TuneForNotes = make([]string, 0, 0)
		app.TuneForSolutions = make([]string, 0, 0)
		if err := app.SaveConfig(); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return fmt.Errorf("Failed to revert one or more SAP notes/solutions: %v", allErrs)
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

// VerifySolution inspect the system and verify that all parameters conform
// to all of the notes associated to the solution.
// The note comparison results will always contain all fields from all notes.
func (app *App) VerifySolution(solName string) (unsatisfiedNotes []string, comparisons map[string]map[string]note.FieldComparison, err error) {
	unsatisfiedNotes = make([]string, 0, 0)
	comparisons = make(map[string]map[string]note.FieldComparison)
	sol, err := app.GetSolutionByName(solName)
	if err != nil {
		return nil, nil, err
	}
	for _, noteID := range sol {
		conforming, noteComparisons, _, err := app.VerifyNote(noteID)
		if err != nil {
			return nil, nil, err
		} else if !conforming {
			unsatisfiedNotes = append(unsatisfiedNotes, noteID)
		}
		comparisons[noteID] = noteComparisons
	}
	return
}

// VerifyAll inspect the system and verify all parameters against all enabled
// notes/solutions.
// The note comparison results will always contain all fields from all notes.
func (app *App) VerifyAll() (unsatisfiedNotes []string, comparisons map[string]map[string]note.FieldComparison, err error) {
	unsatisfiedNotes = make([]string, 0, 0)
	comparisons = make(map[string]map[string]note.FieldComparison)
	for _, solName := range app.TuneForSolutions {
		// Collect field comparison results from solution notes
		unsatisfiedSolNotes, noteComparisons, err := app.VerifySolution(solName)
		if err != nil {
			return nil, nil, err
		} else if len(unsatisfiedSolNotes) > 0 {
			unsatisfiedNotes = append(unsatisfiedNotes, unsatisfiedSolNotes...)
		}
		for noteName, noteComparisonResult := range noteComparisons {
			comparisons[noteName] = noteComparisonResult
		}
	}
	for _, noteID := range app.TuneForNotes {
		// Collect field comparison results from additionally tuned notes
		conforming, noteComparisons, _, err := app.VerifyNote(noteID)
		if err != nil {
			return nil, nil, err
		} else if !conforming {
			unsatisfiedNotes = append(unsatisfiedNotes, noteID)
		}
		comparisons[noteID] = noteComparisons
	}
	return
}
