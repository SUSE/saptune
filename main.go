package main

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

const (
	SapconfService        = "sapconf.service"
	TunedService          = "tuned.service"
	TunedProfileName      = "saptune"
	ExitTunedStopped      = 1
	ExitTunedWrongProfile = 2
	ExitNotTuned          = 3
	NoteTuningSheets      = "/usr/share/saptune/notes/"
	OverrideTunigSheets   = "/etc/saptune/override/"
	// ExtraTuningSheets is a directory located on file system for external parties to place their tuning option files.
	ExtraTuningSheets     = "/etc/saptune/extra/"
	SetGreenText          = "\033[32m"
	SetRedText            = "\033[31m"
	ResetTextColor        = "\033[0m"
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

var tuneApp *app.App                 // application configuration and tuning states
var tuningOptions note.TuningOptions // Collection of tuning options from SAP notes and 3rd party vendors.
var solutionSelector = runtime.GOARCH

func main() {
	if arg1 := cliArg(1); arg1 == "" || arg1 == "help" || arg1 == "--help" {
		PrintHelpAndExit(0)
	}
	// All other actions require super user privilege
	if os.Geteuid() != 0 {
		errorExit("Please run saptune with root privilege.")
		return
	}
	var saptune_log io.Writer
	saptune_log, err := os.OpenFile("/var/log/tuned/tuned.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err.Error())
	}
	saptune_writer := io.MultiWriter(os.Stderr, saptune_log)
	log.SetOutput(saptune_writer)
	if system.IsPagecacheAvailable() {
		solutionSelector = solutionSelector + "_PC"
	}
	archSolutions, exist := solution.AllSolutions[solutionSelector]
	if !exist {
		errorExit("The system architecture (%s) is not supported.", solutionSelector)
		return
	}
	// Initialise application configuration and tuning procedures
	tuningOptions = note.GetTuningOptions(NoteTuningSheets, ExtraTuningSheets)
	tuneApp = app.InitialiseApp("", "", tuningOptions, archSolutions)
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
		system.SystemctlDisableStop(SapconfService) // do not error exit on failure
		if err := system.WriteTunedAdmProfile("saptune"); err != nil {
			errorExit("%v", err)
		}
		if err := system.SystemctlEnableStart(TunedService); err != nil {
			errorExit("%v", err)
		}
		// tuned then calls `saptune daemon apply`
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
		if system.SystemctlIsRunning(TunedService) {
			fmt.Println("Daemon (tuned.service) is running.")
		} else {
			fmt.Fprintln(os.Stderr, "Daemon (tuned.service) is stopped. If you wish to start the daemon, run `saptune daemon start`.")
			os.Exit(ExitTunedStopped)
		}
		// Check tuned profile
		if system.GetTunedProfile() != TunedProfileName {
			fmt.Fprintln(os.Stderr, "tuned.service profile is incorrect. If you wish to correct it, run `saptune daemon start`.")
			os.Exit(ExitTunedWrongProfile)
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
			os.Exit(ExitNotTuned)
		}
	case "stop":
		fmt.Println("Stopping daemon (tuned.service), this may take several seconds...")
		if err := system.SystemctlDisableStop(TunedService); err != nil {
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
func PrintNoteFields(header string, noteComparisons map[string]map[string]note.NoteFieldComparison, printComparison bool) {

	var fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4 int
	// initialise
	printHead := ""
	sortkeys  := make([]string, 0, len(noteComparisons))
	remskeys  := make([]string, 0, len(noteComparisons))
	hasDiff   := false
	compliant := "yes"
	override  := ""
	format    := "\t%s : %s\n"
	noteField := ""
	reminder  := make(map[string]string)

	if printComparison {
		// verify
		fmtlen0 = 16
		fmtlen1 = 12
		fmtlen2 = 9
		fmtlen3 = 9
		fmtlen4 = 7
	} else {
		// simulate
		fmtlen1 = 9
		fmtlen2 = 9
		fmtlen3 = 14
		fmtlen4 = 9
	}

	// sort output
	for noteID, comparisons := range noteComparisons {
		for _, comparison := range comparisons {
			if len(comparison.ReflectMapKey) != 0 && comparison.ReflectFieldName != "OverrideParams" {
				if comparison.ReflectMapKey != "reminder" {
					sortkeys = append(sortkeys, noteID + "@" + comparison.ReflectMapKey)
				} else {
					remskeys = append(remskeys, noteID + "@" + comparison.ReflectMapKey)
				}
			}
		}
	}
	sort.Strings(sortkeys)
	for _, rem := range remskeys {
		sortkeys = append(sortkeys, rem)
	}

	// setup format values
	for _ , skey := range sortkeys {
		keyFields := strings.Split(skey, "@")
		noteID := keyFields[0]
		comparisons := noteComparisons[noteID]
		for _, comparison := range comparisons {
			if comparison.ReflectMapKey == "reminder" {
				continue
			}
			if printComparison {
				// verify
				if len(noteField) > fmtlen0 {
					fmtlen0 = len(noteField)
				}
				if len(comparison.ReflectMapKey) != 0 {
					if comparison.ReflectFieldName == "OverrideParams" && len(comparison.ActualValueJS) > fmtlen3 {
						fmtlen3 = len(comparison.ActualValueJS)
						continue
					}
					if len(comparison.ReflectMapKey) > fmtlen1 {
						fmtlen1 = len(comparison.ReflectMapKey)
					}
					if len(comparison.ExpectedValueJS) > fmtlen2 {
						fmtlen2 = len(comparison.ExpectedValueJS)
					}
					if len(comparison.ActualValueJS) > fmtlen4 {
						fmtlen4 = len(comparison.ActualValueJS)
					}
				}
				format = "   %-" + strconv.Itoa(fmtlen0) + "s | %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %3s\n"
			} else {
				// simulate
				if len(comparison.ReflectMapKey) != 0 {
					if len(comparison.ReflectMapKey) > fmtlen1 {
						fmtlen1 = len(comparison.ReflectMapKey)
					}
					if len(comparison.ActualValueJS) > fmtlen2 {
						fmtlen2 = len(comparison.ActualValueJS)
					}
					if len(comparison.ExpectedValueJS) > fmtlen3 {
						fmtlen3 = len(comparison.ExpectedValueJS)
					}
				}
				format = "   %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s| %-" + strconv.Itoa(fmtlen3) + "s| %-" + strconv.Itoa(fmtlen4) + "s\n"
			}
		}
	}

	// print
	noteID := ""
	for _ , skey := range sortkeys {
		keyFields := strings.Split(skey, "@")
		key := keyFields[1]
		printHead = ""
		if keyFields[0] != noteID {
			if noteID == "" {
				printHead = "yes"
			}
			noteID = keyFields[0]
			noteField = fmt.Sprintf("%s, %s", noteID, txtparser.GetINIFileVersion(noteComparisons[noteID]["ConfFilePath"].ActualValue.(string)))
		}

		comparison := noteComparisons[noteID][fmt.Sprintf("%s[%s]", "SysctlParams", key)]
		override = strings.Replace(noteComparisons[noteID][fmt.Sprintf("%s[%s]", "OverrideParams", key)].ExpectedValueJS, "\t", " ", -1)

		if comparison.ReflectMapKey == "reminder" {
			reminder[noteID] = reminder[noteID] + comparison.ExpectedValueJS
			continue
		}
		if !comparison.MatchExpectation {
			hasDiff = true
			compliant = "no"
		} else {
			compliant = "yes"
		}

		// print table header
		if printHead != "" {
			if header != "NONE" {
				fmt.Printf("%s - %s \n\n", noteID, tuningOptions[noteID].Name())
			}
			if printComparison {
				// verify
				fmt.Printf(format, "SAPNote, Version", "Parameter", "Expected", "Override", "Actual", "Compliant")
				for i := 0; i < fmtlen0+fmtlen1+fmtlen2+fmtlen3+fmtlen4+28 ; i++ {
					if i == 3+fmtlen0+1 || i == 3+fmtlen0+3+fmtlen1+1 || i == 3+fmtlen0+3+fmtlen1+4+fmtlen2 || i == 3+fmtlen0+3+fmtlen1+4+fmtlen2+2+fmtlen3+1 || i == 3+fmtlen0+3+fmtlen1+4+fmtlen2+2+fmtlen3+3+fmtlen4+1 {
						fmt.Printf("+")
					} else {
						fmt.Printf("-")
					}
				}
				fmt.Printf("\n")
			} else {
				// simulate
				fmt.Printf(format, "Parameter", "Value set", "Value expected", "Override")
				for i := 0; i < fmtlen1+fmtlen2+fmtlen3+fmtlen4+8 ; i++ {
					if i == 3+fmtlen1+1 || i == 3+fmtlen1+3+fmtlen2 || i == 3+fmtlen1+3+fmtlen2+1+fmtlen3+1 {
						fmt.Printf("+")
					} else {
						fmt.Printf("-")
					}
				}
				fmt.Printf("\n")
			}
		}

		// print table body
		if printComparison {
			// verify
			fmt.Printf(format, noteField, comparison.ReflectMapKey, strings.Replace(comparison.ExpectedValueJS, "\t", " ", -1), override, strings.Replace(comparison.ActualValueJS, "\t", " ", -1), compliant)
		} else {
			// simulate
			fmt.Printf(format, comparison.ReflectMapKey, strings.Replace(comparison.ActualValueJS, "\t", " ", -1), strings.Replace(comparison.ExpectedValueJS, "\t", " ", -1), override)
		}
	}
	// print footer
	if header != "NONE" && !hasDiff {
		fmt.Printf("\n   (no change)\n")
	}
	fmt.Printf("\n")
	for noteID, reminde := range reminder {
		if reminde != "" {
			reminderHead := fmt.Sprintf("Attention for SAP Note %s:\nHints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:\n", noteID)
			fmt.Printf("%s\n", SetRedText + reminderHead + reminde + ResetTextColor)
		}
	}
}

// Verify that all system parameters do not deviate from any of the enabled solutions/notes.
func VerifyAllParameters() {
	unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
	if err != nil {
		errorExit("Failed to inspect the current system: %v", err)
	}
	PrintNoteFields("NONE", comparisons, true)
	if len(unsatisfiedNotes) == 0 {
		fmt.Println("The running system is currently well-tuned according to all of the enabled notes.")
	} else {
		errorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
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
		if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
			fmt.Println("\nRemember: if you wish to automatically activate the solution's tuning options after a reboot," +
				"you must instruct saptune to configure \"tuned\" daemon by running:" +
				"\n    saptune daemon start")
		}
	case "list":
		fmt.Println("All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, O denotes override file exists for note):")
		solutionNoteIDs := tuneApp.GetSortedSolutionEnabledNotes()
		for _, noteID := range tuningOptions.GetSortedIDs() {
			if strings.HasSuffix(noteID, "_n2c") {
				continue
			}
			noteObj := tuningOptions[noteID]
			format := "\t%s\t\t%s\n"
			if len(noteID) >= 8 {
				format = "\t%s\t%s\n"
			}
			if _, err := os.Stat(fmt.Sprintf("%s%s", OverrideTunigSheets, noteID)); err == nil {
				format = " O" + format
			}
			if i := sort.SearchStrings(solutionNoteIDs, noteID); i < len(solutionNoteIDs) && solutionNoteIDs[i] == noteID {
				format = " " + SetGreenText + "*" + format + ResetTextColor
			} else if i := sort.SearchStrings(tuneApp.TuneForNotes, noteID); i < len(tuneApp.TuneForNotes) && tuneApp.TuneForNotes[i] == noteID {
				format = " " + SetGreenText + "+" + format + ResetTextColor
			}
			fmt.Printf(format, noteID, noteObj.Name())
		}
		if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
			fmt.Println("\nRemember: if you wish to automatically activate the solution's tuning options after a reboot," +
				"you must instruct saptune to configure \"tuned\" daemon by running:" +
				"\n    saptune daemon start")
		}
	case "verify":
		if noteID == "" {
			VerifyAllParameters()
		} else {
			// Check system parameters against the specified note, no matter the note has been tuned for or not.
			conforming, comparisons, _, err := tuneApp.VerifyNote(noteID)
			if err != nil {
				errorExit("Failed to test the current system against the specified note: %v", err)
			}
			noteComp := make(map[string]map[string]note.NoteFieldComparison)
			noteComp[noteID] = comparisons
			PrintNoteFields("HEAD", noteComp, true)
			if !conforming {
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
		if _, comparisons, _, err := tuneApp.VerifyNote(noteID); err != nil {
			errorExit("Failed to test the current system against the specified note: %v", err)
		} else {
			fmt.Printf("If you run `saptune note apply %s`, the following changes will be applied to your system:\n", noteID)
			noteComp := make(map[string]map[string]note.NoteFieldComparison)
			noteComp[noteID] = comparisons
			PrintNoteFields("HEAD", noteComp, true)
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
		fmt.Println("Please note: the reverted note may still show up in list of enabled notes, if an enabled solution refers to it.")
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
				fmt.Printf("\t%s\t%s\n", noteNumber, tuningOptions[noteNumber].Name())
			}
		}
		if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
			fmt.Println("\nRemember: if you wish to automatically activate the solution's tuning options after a reboot," +
				"you must instruct saptune to configure \"tuned\" daemon by running:" +
				"\n    saptune daemon start")
		}
	case "list":
		fmt.Println("All solutions (* denotes enabled solution):")
		for _, solName := range solution.GetSortedSolutionNames(solutionSelector) {
			format := "\t%-18s -"
			if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
				format = " " + SetGreenText + "*" + format
			}
			solNotes := ""
			for _, noteString := range solution.AllSolutions[solutionSelector][solName] {
				solNotes = solNotes + " " + noteString
			}
			format = format + solNotes + ResetTextColor + "\n"
			fmt.Printf(format, solName)
		}
		if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
			fmt.Println("\nRemember: if you wish to automatically activate the solution's tuning options after a reboot," +
				"you must instruct saptune to configure \"tuned\" daemon by running:" +
				"\n    saptune daemon start")
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
			PrintNoteFields("NONE", comparisons, true)
			if len(unsatisfiedNotes) == 0 {
				fmt.Println("The system fully conforms to the tuning guidelines of the specified SAP solution.")
			} else {
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
			PrintNoteFields("NONE", comparisons, true)
		}
	case "revert":
		if solName == "" {
			PrintHelpAndExit(1)
		}
		if err := tuneApp.RevertSolution(solName); err != nil {
			errorExit("Failed to revert tuning for solution %s: %v", solName, err)
		}
		fmt.Println("Parameters tuned by the notes referred by the SAP solution have been successfully reverted.")
	default:
		PrintHelpAndExit(1)
	}
}
