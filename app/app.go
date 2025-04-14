package app

import (
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"reflect"
	"sort"
)

// define saptunes main configuration variables
const (
	TuneForSolutionsKey = "TUNE_FOR_SOLUTIONS"
	TuneForNotesKey     = "TUNE_FOR_NOTES"
	NoteApplyOrderKey   = "NOTE_APPLY_ORDER"
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

// define saptunes main configuration file
var sysconfigSaptuneFile = system.SaptuneConfigFile()

// InitialiseApp load application configuration. Panic on error.
func InitialiseApp(sysconfigPrefix, stateDirPrefix string, allNotes map[string]note.Note, allSolutions map[string]solution.Solution) (app *App) {
	app = &App{
		SysconfigPrefix: sysconfigPrefix,
		State:           &State{StateDirPrefix: stateDirPrefix},
		AllNotes:        allNotes,
		AllSolutions:    allSolutions,
	}
	sysconf, err := txtparser.ParseSysconfigFile(path.Join(app.SysconfigPrefix, sysconfigSaptuneFile), true)
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
	// Never ever sort app.NoteApplyOrder !
	return
}

// SaveConfig save configuration to file /etc/sysconfig/saptune.
func (app *App) SaveConfig() error {
	sysconf, err := txtparser.ParseSysconfigFile(path.Join(app.SysconfigPrefix, sysconfigSaptuneFile), true)
	if err != nil {
		return err
	}
	sysconf.SetStrArray(TuneForSolutionsKey, app.TuneForSolutions)
	sysconf.SetStrArray(TuneForNotesKey, app.TuneForNotes)
	sysconf.SetStrArray(NoteApplyOrderKey, app.NoteApplyOrder)
	return os.WriteFile(path.Join(app.SysconfigPrefix, sysconfigSaptuneFile), []byte(sysconf.ToText()), 0644)
}

// removeFromConfig removes NoteID from the variables in the configuration
// changes TuneForNotes and NoteApplyOrder
func (app *App) removeFromConfig(noteID string) {
	i := sort.SearchStrings(app.TuneForNotes, noteID)
	if i < len(app.TuneForNotes) && app.TuneForNotes[i] == noteID {
		// remove noteID from the configuration 'TuneForNotes'
		app.TuneForNotes = append(app.TuneForNotes[0:i], app.TuneForNotes[i+1:]...)
	}
	i = app.PositionInNoteApplyOrder(noteID)
	if i < 0 {
		system.WarningLog("noteID '%s' not found in configuration 'NoteApplyOrder'", noteID)
	} else if i < len(app.NoteApplyOrder) && app.NoteApplyOrder[i] == noteID {
		// remove noteID from the configuration 'NoteApplyOrder'
		app.NoteApplyOrder = append(app.NoteApplyOrder[0:i], app.NoteApplyOrder[i+1:]...)
	}
}

// handleCounterParts will save the counterpart values of parameters
func handleCounterParts(currentState interface{}) bool {
	forceApply := false
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
	return forceApply
}

// TuneAll tune for all currently enabled solutions and notes.
func (app *App) TuneAll() error {
	for _, noteID := range app.NoteApplyOrder {
		if _, err := os.Stat(app.State.GetPathToNote(noteID)); err == nil {
			// state file for note already exists
			continue
		}
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

// RevertAll revert all tuned parameters (both solutions and additional notes),
// and clear stored states, but NOT NoteApplyOrder.
func (app *App) RevertAll(permanent bool) error {
	allErrs := make([]error, 0)

	// Simply revert all notes from serialised states
	otherNotes, err := app.State.List()
	if err == nil {
		for _, otherNoteID := range otherNotes {
			// check, if empty state file exists
			if content, err := os.ReadFile(app.State.GetPathToNote(otherNoteID)); err == nil && len(content) == 0 {
				// remove empty state file
				_ = app.State.Remove(otherNoteID)
				continue
			}
			if err := app.RevertNote(otherNoteID, permanent); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	} else {
		allErrs = append(allErrs, err)
	}
	if permanent {
		app.TuneForNotes = make([]string, 0)
		app.TuneForSolutions = make([]string, 0)
		app.NoteApplyOrder = make([]string, 0)
		if err := app.SaveConfig(); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return fmt.Errorf("Failed to revert one or more SAP notes/solutions: %v", allErrs)
}

// VerifyAll inspect the system and verify all parameters against all enabled
// notes/solutions.
// The note comparison results will always contain all fields from all notes.
func (app *App) VerifyAll(chkApplied bool) (unsatisfiedNotes []string, comparisons map[string]map[string]note.FieldComparison, err error) {
	unsatisfiedNotes = make([]string, 0)
	comparisons = make(map[string]map[string]note.FieldComparison)
	for _, noteID := range app.NoteApplyOrder {
		// Collect field comparison results from all enabled notes
		if chkApplied {
			// Collect field comparison results from all applied (tuned) notes
			if _, ok := app.IsNoteApplied(noteID); !ok {
				// note not applied
				continue
			}
		}
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
