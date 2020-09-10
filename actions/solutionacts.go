package actions

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"sort"
)

// SolutionAction  Solution actions like apply, revert, verify asm.
func SolutionAction(actionName, solName string, tuneApp *app.App) {
	switch actionName {
	case "apply":
		SolutionActionApply(os.Stdout, solName, tuneApp)
	case "list":
		SolutionActionList(os.Stdout, tuneApp)
	case "verify":
		SolutionActionVerify(os.Stdout, solName, tuneApp)
	case "simulate":
		SolutionActionSimulate(os.Stdout, solName, tuneApp)
	case "revert":
		SolutionActionRevert(os.Stdout, solName, tuneApp)
	case "enabled":
		SolutionActionEnabled(os.Stdout, tuneApp)
	default:
		PrintHelpAndExit(1)
	}
}

// SolutionActionApply applies parameter settings defined by the solution
// to the system
func SolutionActionApply(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(1)
	}
	if len(tuneApp.TuneForSolutions) > 0 {
		// already one solution applied.
		// do not apply another solution. Does not make sense
		system.ErrorLog("There is already one solution applied. Applying another solution is NOT supported.")
		system.ErrorExit("", 1)
	}
	removedAdditionalNotes, err := tuneApp.TuneSolution(solName)
	if err != nil {
		system.ErrorExit("Failed to tune for solution %s: %v", solName, err)
	}
	fmt.Fprintf(writer, "All tuning options for the SAP solution have been applied successfully.\n")
	if len(removedAdditionalNotes) > 0 {
		fmt.Fprintf(writer, "\nThe following previously-enabled notes are now tuned by the SAP solution:\n")
		for _, noteNumber := range removedAdditionalNotes {
			fmt.Fprintf(writer, "\t%s\t%s\n", noteNumber, tuneApp.AllNotes[noteNumber].Name())
		}
	}
	rememberMessage(writer)
}

// SolutionActionList lists all available solution definitions
func SolutionActionList(writer io.Writer, tuneApp *app.App) {
	setColor := false
	solutionSelector := system.GetSolutionSelector()
	fmt.Fprintf(writer, "\nAll solutions (* denotes enabled solution, O denotes override file exists for solution, D denotes deprecated solutions):\n")
	for _, solName := range solution.GetSortedSolutionNames(solutionSelector) {
		format := "\t%-18s -"
		if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
			format = " " + setGreenText + "*" + format
			setColor = true
		}
		if len(solution.OverrideSolutions[solutionSelector][solName]) != 0 {
			//override solution
			format = " O" + format
		}

		solNotes := ""
		for _, noteString := range solution.AllSolutions[solutionSelector][solName] {
			solNotes = solNotes + " " + noteString
		}
		if _, ok := solution.DeprecSolutions[solutionSelector][solName]; ok {
			format = " D" + format
		}
		format = format + solNotes
		if setColor {
			format = format + resetTextColor
		}
		format = format + "\n"
		//fmt.Printf(format, solName)
		fmt.Fprintf(writer, format, solName)
	}
	rememberMessage(writer)
}

// SolutionActionVerify compares all parameter settings from a solution
// definition against the system settings
func SolutionActionVerify(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		VerifyAllParameters(writer, tuneApp)
	} else {
		// Check system parameters against the specified solution, no matter the solution has been tuned for or not.
		unsatisfiedNotes, comparisons, err := tuneApp.VerifySolution(solName)
		if err != nil {
			system.ErrorExit("Failed to test the current system against the specified SAP solution: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true)
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "The system fully conforms to the tuning guidelines of the specified SAP solution.\n")
		} else {
			system.ErrorExit("The parameters listed above have deviated from the specified SAP solution recommendations.\n")
		}
	}
}

// SolutionActionSimulate shows all changes that will be applied to the system if
// the solution will be applied.
func SolutionActionSimulate(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(1)
	}
	// Run verify and print out all fields of the note
	if _, comparisons, err := tuneApp.VerifySolution(solName); err != nil {
		system.ErrorExit("Failed to test the current system against the specified note: %v", err)
	} else {
		fmt.Fprintf(writer, "If you run `saptune solution apply %s`, the following changes will be applied to your system:\n", solName)
		PrintNoteFields(writer, "NONE", comparisons, false)
	}
}

// SolutionActionRevert reverts all parameter settings of a solution back to
// the state before 'apply'
func SolutionActionRevert(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(1)
	}
	if err := tuneApp.RevertSolution(solName); err != nil {
		system.ErrorExit("Failed to revert tuning for solution %s: %v", solName, err)
	}
	fmt.Fprintf(writer, "Parameters tuned by the notes referred by the SAP solution have been successfully reverted.\n")
}

// SolutionActionEnabled prints out the enabled solution definition
func SolutionActionEnabled(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.TuneForSolutions) != 0 {
		fmt.Fprintf(writer, "%s", tuneApp.TuneForSolutions[0])
	}
}
