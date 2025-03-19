package app

import (
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"sort"
)

// IsSolutionEnabled returns true, if the solution is enabled or false, if not
// enbaled means - part of TuneForSolutions
func (app *App) IsSolutionEnabled(sol string) bool {
	i := sort.SearchStrings(app.TuneForSolutions, sol)
	if i < len(app.TuneForSolutions) && app.TuneForSolutions[i] == sol {
		return true
	}
	return false
}

// IsSolutionApplied returns true, if the solution is (partial) applied
// or false, if not
func (app *App) IsSolutionApplied(sol string) (string, bool) {
	state := ""
	ret := false
	if len(app.TuneForSolutions) != 0 {
		if app.TuneForSolutions[0] == sol {
			noteOK := 0
			noteCnt := 0
			for _, note := range app.AllSolutions[sol] {
				noteCnt = noteCnt + 1
				if _, ok := app.IsNoteApplied(note); ok {
					noteOK = noteOK + 1
				}
			}
			if noteOK == noteCnt {
				ret = true
				state = "fully"
			} else if noteOK != 0 {
				ret = true
				state = "partial"
			}
		}
	}
	return state, ret
}

// AppliedSolution returns the currently applied Solution
func (app *App) AppliedSolution() (string, string) {
	solApplied := ""
	state := ""
	if len(app.TuneForSolutions) != 0 {
		solName := app.TuneForSolutions[0]
		if st, ok := app.IsSolutionApplied(solName); ok {
			solApplied = solName
			state = st
		}
	}
	return solApplied, state
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

// TuneSolution apply tuning for a solution.
// If the solution is not yet enabled, the name will be added into the list
// of tuned solution names.
// If the solution covers any of the additional notes, those notes will be removed.
func (app *App) TuneSolution(solName string) (removedExplicitNotes []string, err error) {
	removedExplicitNotes = make([]string, 0)
	sol, err := app.GetSolutionByName(solName)
	if err != nil {
		return
	}
	// store note list of the currently active/applied solution definition
	if err = solution.StoreActiveSolNoteInfo(sol, solName); err != nil {
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
		if _, ok := app.IsNoteApplied(noteID); ok {
			continue
		}
		if err = app.TuneNote(noteID); err != nil {
			return
		}
	}
	return
}

// RemoveSolFromConfig removes the given solution from the configuration
func (app *App) RemoveSolFromConfig(solName string) error {
	i := sort.SearchStrings(app.TuneForSolutions, solName)
	if i < len(app.TuneForSolutions) && app.TuneForSolutions[i] == solName {
		app.TuneForSolutions = append(app.TuneForSolutions[0:i], app.TuneForSolutions[i+1:]...)
		if err := app.SaveConfig(); err != nil {
			return err
		}
	}
	return nil
}

// RevertSolution permanently revert notes tuned by the solution and
// clear their stored states.
func (app *App) RevertSolution(solName string) error {
	// Read and remove run time info needed e.g. for refresh
	sol, err := solution.GetActiveSolNoteInfo(solName, true)
	if err != nil {
		// fallback, if runtime info is not available
		// e.g. solution apply was from saptune version < 3.2
		sol, err = app.GetSolutionByName(solName)
		if err != nil {
			return err
		}
	}
	// Remove from configuration
	if err := app.RemoveSolFromConfig(solName); err != nil {
		return err
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
	noteErrs := make([]error, 0)
	for _, noteID := range sol {
		if _, found := notesDoNotRevert[noteID]; found {
			continue // skip this one
		}
		if err := app.RevertNote(noteID, true); err != nil {
			noteErrs = append(noteErrs, err)
		}
	}
	if len(noteErrs) == 0 {
		return nil
	}
	return fmt.Errorf("Failed to revert one or more SAP notes that belong to the solution: %v", noteErrs)
}

// VerifySolution inspect the system and verify that all parameters conform
// to all of the notes associated to the solution.
// The note comparison results will always contain all fields from all notes.
func (app *App) VerifySolution(solName string) (unsatisfiedNotes []string, comparisons map[string]map[string]note.FieldComparison, err error) {
	unsatisfiedNotes = make([]string, 0)
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
