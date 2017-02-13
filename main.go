package main

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/app"
	"github.com/HouzuoGuo/saptune/sap/note"
	"github.com/HouzuoGuo/saptune/sap/solution"
	"github.com/HouzuoGuo/saptune/system"
	"os"
	"runtime"
	"sort"
	"syscall"
)

const (
	SAPCONF_SERVICE          = "sapconf.service"
	TUNED_SERVICE            = "tuned.service"
	TUNED_PROFILE_NAME       = "saptune"
	EXIT_TUNED_STOPPED       = 1
	EXIT_TUNED_WRONG_PROFILE = 2
	EXIT_NOT_TUNED           = 3
	VENDOR_DIR               = "/etc/saptune/extra/"
	VENDOR_FILE              = "HPE-Recommended_OS_settings.conf"
)

func PrintHelpAndExit(exitStatus int) {
	fmt.Println(`saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify ]
  saptune note [ apply | simulate | verify | customise | revert ] NoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify ]
  saptune solution [ apply | simulate | verify | revert ] SolutionName
`)
	os.Exit(exitStatus)
}

// Print the message to stderr and exit 1.
func errorExit(template string, stuff ...interface{}) {
	fmt.Fprintf(os.Stderr, template+"\n", stuff...)
	os.Exit(1)
}

// Return the i-th command line parameter, or empty string if it is not specified.
func cliArg(i int) string {
	if len(os.Args) >= i+1 {
		return os.Args[i]
	}
	return ""
}

var tuneApp *app.App

func main() {
	defer func() {
		if err := recover(); err != nil {
			errorExit("Critical error: %v", err)
		}
	}()
	if arg1 := cliArg(1); arg1 == "" || arg1 == "help" || arg1 == "--help" {
		PrintHelpAndExit(0)
	}
	// All other actions require super user privilege
	if os.Geteuid() != 0 {
		errorExit("Please run saptune with root privilege.")
		return
	}
	archSolutions, exist := solution.AllSolutions[runtime.GOARCH]
	if !exist {
		errorExit("The system architecture (%s) is not supported.", runtime.GOARCH)
		return
	}
	tuneApp = app.InitialiseApp("", "", note.AllNotes, archSolutions)
	switch cliArg(1) {
	case "daemon":
		DaemonAction(cliArg(2))
	case "note":
		NoteAction(cliArg(2), cliArg(3))
	case "solution":
		SolutionAction(cliArg(2), cliArg(3))
	default:
		PrintHelpAndExit(1)
	}
}

func DaemonAction(actionName string) {
	switch actionName {
	case "start":
		fmt.Println("Starting daemon (tuned.service), this may take several seconds...")
		system.SystemctlDisableStop(SAPCONF_SERVICE) // do not error exit on failure
		if err := system.SystemctlEnableStart(TUNED_SERVICE); err != nil {
			errorExit("%v", err)
		}
		if err := system.TunedAdmProfile("saptune"); err != nil {
			errorExit("%v", err)
		}
		// tuned then calls `sapconf daemon apply`
		fmt.Println("Daemon (tuned.service) has been enabled and started.")
		if len(tuneApp.TuneForSolutions) == 0 && len(tuneApp.TuneForNotes) == 0 {
			fmt.Println("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
		}
	case "apply":
		// This action name is only used by tuned script, hence it is not advertised to end user.
		if err := tuneApp.TuneAll(); err != nil {
			panic(err)
		}
	case "status":
		// Check daemon
		if system.SystemctlIsRunning(TUNED_SERVICE) {
			fmt.Println("Daemon (tuned.service) is running.")
		} else {
			fmt.Fprintln(os.Stderr, "Daemon (tuned.service) is stopped. If you wish to start the daemon, run `saptune daemon start`.")
			os.Exit(EXIT_TUNED_STOPPED)
		}
		// Check tuned profile
		if system.GetTunedProfile() != TUNED_PROFILE_NAME {
			fmt.Fprintln(os.Stderr, "tuned.service profile is incorrect. If you wish to correct it, run `saptune daemon start`.")
			os.Exit(EXIT_TUNED_WRONG_PROFILE)
		}
		// Check for any enabled note/solution
		if len(tuneApp.TuneForSolutions) > 0 || len(tuneApp.TuneForNotes) > 0 {
			fmt.Println("The system has been tuned for the following solutions and notes:")
			for _, sol := range tuneApp.TuneForSolutions {
				fmt.Println("\t" + sol)
			}
			for _, noteID := range tuneApp.TuneForNotes {
				fmt.Println("\t" + noteID)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
			os.Exit(EXIT_NOT_TUNED)
		}
	case "stop":
		fmt.Println("Stopping daemon (tuned.service), this may take several seconds...")
		if err := system.SystemctlDisableStop(TUNED_SERVICE); err != nil {
			errorExit("%v", err)
		}
		// tuned then calls `sapconf daemon revert`
		fmt.Println("Daemon (tuned.service) has been disabled and stopped.")
		fmt.Println("All tuned parameters have been reverted to default.")
	case "revert":
		// This action name is only used by tuned script, hence it is not advertised to end user.
		if err := tuneApp.RevertAll(false); err != nil {
			panic(err)
		}
	default:
		PrintHelpAndExit(1)
	}
}

// Print mismatching fields in the note comparison result.
func PrintNoteFields(noteID string, comparisons map[string]note.NoteFieldComparison, printComparison bool) {
	fmt.Printf("%s - %s -\n", noteID, note.AllNotes[noteID].Name())
	hasDiff := false
	for fieldName, fieldComparison := range comparisons {
		if !fieldComparison.MatchExpectation {
			hasDiff = true
			if printComparison {
				fmt.Printf("\t%s Expected: %s\n", fieldName, fieldComparison.ExpectedValueJS)
				fmt.Printf("\t%s Actual  : %s\n", fieldName, fieldComparison.ActualValueJS)
			} else {
				fmt.Printf("\t%s : %s\n", fieldName, fieldComparison.ExpectedValueJS)
			}
		}
	}
	if !hasDiff {
		fmt.Printf("\t(no change)\n")
	}
	if _, err := os.Stat(VENDOR_DIR + VENDOR_FILE); err == nil {
		fmt.Println("Attention: HPE vendor specific optimization")
	}
}

// Verify that all system parameters do not deviate from any of the enabled solutions/notes.
func VerifyAllParameters() {
	unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
	if err != nil {
		errorExit("Failed to inspect the current system: %v", err)
	}
	if len(unsatisfiedNotes) == 0 {
		fmt.Println("The running system is currently well-tuned according to all of the enabled notes.")
	} else {
		for _, unsatisfiedNoteID := range unsatisfiedNotes {
			PrintNoteFields(unsatisfiedNoteID, comparisons[unsatisfiedNoteID], true)
		}
		errorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
	}
}

// Apply all vendor customisations from /etc/saptune/extra
func ApplyVendorSettings() {
	if _, files, err := system.ListDir(note.VENDOR_DIR); err != nil || len(files) == 0 {
		// Nothing under vendor dir
		return
	}
	if _, err := os.Stat(note.VENDOR_DIR); err == nil {
		fmt.Println("Applying additional vendor customisations...")
		vendorSettings := note.VendorSettings{}
		if err := vendorSettings.Apply(); err != nil {
			fmt.Println("Error: ", err.Error())
		}
	}
}

func NoteAction(actionName, noteID string) {
	switch actionName {
	case "apply":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		if err := tuneApp.TuneNote(noteID); err != nil {
			errorExit("Failed to tune for note %s: %v", noteID, err)
		}
		fmt.Println("The note has been applied successfully.")
		ApplyVendorSettings()
	case "list":
		fmt.Println("All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions):")
		solutionNoteIDs := tuneApp.GetSortedSolutionEnabledNotes()
		for _, noteID := range note.GetSortedNoteIDs() {
			noteObj := note.AllNotes[noteID]
			format := "\t%s\t%s\n"
			if i := sort.SearchStrings(solutionNoteIDs, noteID); i < len(solutionNoteIDs) && solutionNoteIDs[i] == noteID {
				format = "*" + format
			} else if i := sort.SearchStrings(tuneApp.TuneForNotes, noteID); i < len(tuneApp.TuneForNotes) && tuneApp.TuneForNotes[i] == noteID {
				format = "+" + format
			}
			fmt.Printf(format, noteID, noteObj.Name())
		}
	case "verify":
		if noteID == "" {
			VerifyAllParameters()
		} else {
			// Check system parameters against the specified note, no matter the note has been tuned for or not.
			if conforming, comparisons, err := tuneApp.VerifyNote(noteID); err != nil {
				errorExit("Failed to test the current system against the specified note: %v", err)
			} else if !conforming {
				PrintNoteFields(noteID, comparisons, true)
				errorExit("The parameters listed above have deviated from the specified note.\n")
			} else {
				fmt.Println("The system fully conforms to the specified note.")
			}
		}
	case "simulate":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		// Run verify and print out all fields of the note
		if _, comparisons, err := tuneApp.VerifyNote(noteID); err != nil {
			errorExit("Failed to test the current system against the specified note: %v", err)
		} else {
			fmt.Printf("If you run `saptune note apply %s`, the following changes will be applied to your system:\n", noteID)
			PrintNoteFields(noteID, comparisons, false)
		}
	case "customise":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		if _, err := tuneApp.GetNoteByID(noteID); err != nil {
			errorExit("%v", err)
		}
		fileName := fmt.Sprintf("/etc/sysconfig/saptune-note-%s", noteID)
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			errorExit("Note %s does not require additional customisation input.", noteID)
		} else if err != nil {
			errorExit("Failed to read file '%s' - %v", fileName, err)
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "/usr/bin/vim" // launch vim by default
		}
		if err := syscall.Exec(editor, []string{editor, fileName}, os.Environ()); err != nil {
			errorExit("Failed to start launch editor %s: %v", editor, err)
		}
	case "revert":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		if err := tuneApp.RevertNote(noteID, true); err != nil {
			errorExit("Failed to revert note %s: %v", noteID, err)
		}
		fmt.Println("Parameters tuned by the note have been successfully reverted.")
		fmt.Println("Plese note: the reverted note may still show up in list of enabled notes, if an enabled solution refers to it.")
		fmt.Println("Plese note: vendor customisations (if any) cannot be reverted automatically.")
	default:
		PrintHelpAndExit(1)
	}
}

func SolutionAction(actionName, solName string) {
	switch actionName {
	case "apply":
		if solName == "" {
			PrintHelpAndExit(1)
		}
		removedAdditionalNotes, err := tuneApp.TuneSolution(solName)
		if err != nil {
			errorExit("Failed to tune for solution %s: %v", solName, err)
		}
		fmt.Println("All tuning options for the SAP solution have been applied successfully.")
		if len(removedAdditionalNotes) > 0 {
			fmt.Println("The following previously-enabled notes are now tuned by the SAP solution:")
			for _, noteNumber := range removedAdditionalNotes {
				fmt.Printf("\t%s\t%s\n", noteNumber, note.AllNotes[noteNumber].Name())
			}
		}
		ApplyVendorSettings()
	case "list":
		fmt.Println("All solutions (* denotes enabled solution):")
		for _, solName := range solution.GetSortedSolutionNames(runtime.GOARCH) {
			format := "\t%s\n"
			if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
				format = "*" + format
			}
			fmt.Printf(format, solName)
		}
	case "verify":
		if solName == "" {
			VerifyAllParameters()
		} else {
			// Check system parameters against the specified solution, no matter the solution has been tuned for or not.
			unsatisfiedNotes, comparisons, err := tuneApp.VerifySolution(solName)
			if err != nil {
				errorExit("Failed to test the current system against the specified SAP solution: %v", err)
			}
			if len(unsatisfiedNotes) == 0 {
				fmt.Println("The system fully conforms to the tuning guidelines of the specified SAP solution.")
			} else {
				for _, unsatisfiedNoteID := range unsatisfiedNotes {
					PrintNoteFields(unsatisfiedNoteID, comparisons[unsatisfiedNoteID], true)
				}
				errorExit("The parameters listed above have deviated from the specified SAP solution recommendations.\n")
			}
		}
	case "simulate":
		if solName == "" {
			PrintHelpAndExit(1)
		}
		// Run verify and print out all fields of the note
		if _, comparisons, err := tuneApp.VerifySolution(solName); err != nil {
			errorExit("Failed to test the current system against the specified note: %v", err)
		} else {
			fmt.Printf("If you run `saptune solution apply %s`, the following changes will be applied to your system:\n", solName)
			for noteID, noteComparison := range comparisons {
				PrintNoteFields(noteID, noteComparison, false)
			}
		}
	case "revert":
		if solName == "" {
			PrintHelpAndExit(1)
		}
		if err := tuneApp.RevertSolution(solName); err != nil {
			errorExit("Failed to revert tuning for solution %s: %v", solName, err)
		}
		fmt.Println("Parameters tuned by the notes referred by the SAP solution have been successfully reverted.")
		fmt.Println("Plese note: vendor customisations (if any) cannot be reverted automatically.")
	default:
		PrintHelpAndExit(1)
	}
}
