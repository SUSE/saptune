package main

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// constant definitions
const (
	SapconfService        = "sapconf.service"
	TunedService          = "tuned.service"
	TunedProfileName      = "saptune"
	NoteTuningSheets      = "/usr/share/saptune/notes/"
	OverrideTuningSheets  = "/etc/saptune/override/"
	ExtraTuningSheets     = "/etc/saptune/extra/" // ExtraTuningSheets is a directory located on file system for external parties to place their tuning option files.
	exitTunedStopped      = 1
	exitTunedWrongProfile = 2
	exitNotTuned          = 3
	saptuneV1             = "/usr/sbin/saptune_v1"
	setGreenText          = "\033[32m"
	setRedText            = "\033[31m"
	resetTextColor        = "\033[0m"
	footnote1X86          = "[1] setting is not supported by the system"
	footnote1IBM          = "[1] setting is not relevant for the system"
	footnote2             = "[2] setting is not available on the system"
	footnote3             = "[3] value is only checked, but NOT set"
	footnote4             = "[4] cpu idle state settings differ"
)

// PrintHelpAndExit Print the usage and exit
func PrintHelpAndExit(exitStatus int) {
	fmt.Println(`saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify ]
  saptune note [ apply | simulate | verify | customise | create | revert ] NoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify ]
  saptune solution [ apply | simulate | verify | revert ] SolutionName
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Print current saptune version:
  saptune version
Print this message:
  saptune help`)
	os.Exit(exitStatus)
}

// Print the message to stderr and exit 1.
func errorExit(template string, stuff ...interface{}) {
	exState := 1
	fieldType := ""
	field := len(stuff) - 1
	if field >= 0 {
		fieldType = reflect.TypeOf(stuff[field]).String()
	}
	if fieldType == "*exec.ExitError" {
		// get return code of failed command, if available
		if exitError, ok := stuff[field].(*exec.ExitError); ok {
			exState = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		}
	}
	system.ErrorLog(template+"\n", stuff...)
	os.Exit(exState)
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
var footnote1 = footnote1X86         // set 'unsupported' footnote regarding the architecture
var solutionSelector = runtime.GOARCH

func main() {
	if runtime.GOARCH == "ppc64le" {
		footnote1 = footnote1IBM
	}

	// get saptune version
	sconf, err := txtparser.ParseSysconfigFile("/etc/sysconfig/saptune", true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to read file '/etc/sysconfig/saptune': %v\n", err)
		os.Exit(1)
	}
	saptuneVersion := sconf.GetString("SAPTUNE_VERSION", "")

	if arg1 := cliArg(1); arg1 == "" || arg1 == "help" || arg1 == "--help" {
		PrintHelpAndExit(0)
	}
	if arg1 := cliArg(1); arg1 == "version" || arg1 == "--version" {
		fmt.Printf("current active saptune version is '%s'\n", saptuneVersion)
		os.Exit(0)
	}

	// All other actions require super user privilege
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "Please run saptune with root privilege.\n")
		os.Exit(1)
	}

	// activate logging
	system.LogInit()

	switch saptuneVersion {
	case "1":
		cmd := exec.Command(saptuneV1, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			errorExit("command '%+s %+v' failed with error '%v'\n", saptuneV1, os.Args, err)
		} else {
			os.Exit(0)
		}
	case "2":
		break
	default:
		errorExit("Wrong saptune version in file '/etc/sysconfig/saptune': %s", saptuneVersion)
	}

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

	checkUpdateLeftOvers()

	switch cliArg(1) {
	case "daemon":
		DaemonAction(cliArg(2))
	case "note":
		NoteAction(cliArg(2), cliArg(3))
	case "solution":
		SolutionAction(cliArg(2), cliArg(3))
	case "revert":
		RevertAction(cliArg(2))
	default:
		PrintHelpAndExit(1)
	}
}

// checkUpdateLeftOvers checks for left over files from the migration of
// saptune version 1 to saptune version 2
func checkUpdateLeftOvers() {
	// check for the /etc/tuned/saptune/tuned.conf file created during
	// the package update from saptune v1 to saptune v2
	// give a Warning but go ahead tuning the system
	if system.CheckForPattern("/etc/tuned/saptune/tuned.conf", "#stv1tov2#") {
		system.WarningLog("found file '/etc/tuned/saptune/tuned.conf' left over from the migration of saptune version 1 to saptune version 2. Please check and remove this file as it may work against the settings of some SAP Notes. For more information refer to the man page saptune-migrate(7)")
	}

	// check if old solution or notes are applied
	if len(tuneApp.NoteApplyOrder) == 0 && (len(tuneApp.TuneForNotes) != 0 || len(tuneApp.TuneForSolutions) != 0) {
		errorExit("There are 'old' solutions or notes defined in file '/etc/sysconfig/saptune'. Seems there were some steps missed during the migration from saptune version 1 to version 2. Please check. Refer to saptune-migrate(7) for more information")
	}
}

// RevertAction Revert all notes and solutions
func RevertAction(actionName string) {
	if actionName != "all" {
		PrintHelpAndExit(1)
	}
	fmt.Println("Reverting all notes and solutions, this may take some time...")
	if err := tuneApp.RevertAll(true); err != nil {
		errorExit("Failed to revert notes: %v", err)
		//panic(err)
	}
	fmt.Println("Parameters tuned by the notes and solutions have been successfully reverted.")
}

// DaemonAction handles daemon actions like start, stop, status asm.
func DaemonAction(actionName string) {
	switch actionName {
	case "start":
		fmt.Println("Starting daemon (tuned.service), this may take several seconds...")
		system.SystemctlDisableStop(SapconfService) // do not error exit on failure
		if err := system.TunedAdmProfile("saptune"); err != nil {
			errorExit("%v", err)
		}
		if err := system.SystemctlEnableStart(TunedService); err != nil {
			errorExit("%v", err)
		}
		// Check tuned profile
		if system.GetTunedAdmProfile() != TunedProfileName {
			system.ErrorLog("tuned.service profile is incorrect. Please check tuned logs for more information")
			os.Exit(exitTunedWrongProfile)
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
			os.Exit(exitTunedStopped)
		}
		// Check tuned profile
		if system.GetTunedProfile() != TunedProfileName {
			fmt.Fprintln(os.Stderr, "tuned.service profile is incorrect. If you wish to correct it, run `saptune daemon start`.")
			os.Exit(exitTunedWrongProfile)
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
			os.Exit(exitNotTuned)
		}
	case "stop":
		fmt.Println("Stopping daemon (tuned.service), this may take several seconds...")
		if err := system.TunedAdmOff(); err != nil {
			errorExit("%v", err)
		}
		if err := system.SystemctlDisableStop(TunedService); err != nil {
			errorExit("%v", err)
		}
		// tuned then calls `saptune daemon revert`
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

// PrintNoteFields Print mismatching fields in the note comparison result.
func PrintNoteFields(header string, noteComparisons map[string]map[string]note.FieldComparison, printComparison bool) {

	var fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4 int
	// initialise
	compliant := "yes"
	printHead := ""
	noteField := ""
	sortkeys := make([]string, 0, len(noteComparisons))
	remskeys := make([]string, 0, len(noteComparisons))
	footnote := make([]string, 4, 4)
	reminder := make(map[string]string)
	override := ""
	comment := ""
	hasDiff := false
	format := "\t%s : %s\n"

	if printComparison {
		// verify
		fmtlen0 = 16
		fmtlen1 = 12
		fmtlen2 = 9
		fmtlen3 = 9
		fmtlen4 = 7
	} else {
		// simulate
		fmtlen1 = 12
		fmtlen2 = 10
		fmtlen3 = 15
		fmtlen4 = 9
	}

	// sort output
	for noteID, comparisons := range noteComparisons {
		for _, comparison := range comparisons {
			if comparison.ReflectFieldName == "Inform" {
				// skip inform map to avoid double entries in verify table
				continue
			}
			if len(comparison.ReflectMapKey) != 0 && comparison.ReflectFieldName != "OverrideParams" {
				if comparison.ReflectMapKey != "reminder" {
					sortkeys = append(sortkeys, noteID+"§"+comparison.ReflectMapKey)
				} else {
					remskeys = append(remskeys, noteID+"§"+comparison.ReflectMapKey)
				}
			}
		}
	}
	sort.Strings(sortkeys)
	for _, rem := range remskeys {
		sortkeys = append(sortkeys, rem)
	}

	// setup format values
	for _, skey := range sortkeys {
		keyFields := strings.Split(skey, "§")
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
				format = "   %-" + strconv.Itoa(fmtlen0) + "s | %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %2s\n"
			} else {
				// simulate
				if len(comparison.ReflectMapKey) != 0 {
					if comparison.ReflectFieldName == "OverrideParams" && len(comparison.ActualValueJS) > fmtlen4 {
						fmtlen4 = len(comparison.ActualValueJS)
						continue
					}
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
				format = "   %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %2s\n"
			}
		}
	}

	// print
	noteID := ""
	for _, skey := range sortkeys {
		comment = ""
		keyFields := strings.Split(skey, "§")
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

		// prepare footnote
		switch comparison.ActualValue {
		case "all:none":
			compliant = compliant + " [1]"
			comment = comment + " [1]"
			footnote[0] = footnote1
		case "NA":
			compliant = compliant + " [2]"
			comment = comment + " [2]"
			footnote[1] = footnote2
		}
		if strings.Contains(comparison.ReflectMapKey, "rpm") || strings.Contains(comparison.ReflectMapKey, "grub") {
			compliant = compliant + " [3]"
			comment = comment + " [3]"
			footnote[2] = footnote3
		}

		// check inform map for special settings
		// ANGI: future - check for 'nil', if using noteComparisons[noteID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ActualValue.(string) in general
		if comparison.ReflectMapKey == "force_latency" && noteComparisons[noteID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ActualValue.(string) == "hasDiffs" {
			compliant = "no [4]"
			comment = comment + " [4]"
			footnote[3] = footnote4
		}

		// print table header
		if printHead != "" {
			if header != "NONE" {
				fmt.Printf("\n%s - %s \n\n", noteID, tuningOptions[noteID].Name())
			}
			if printComparison {
				// verify
				fmt.Printf(format, "SAPNote, Version", "Parameter", "Expected", "Override", "Actual", "Compliant")
				for i := 0; i < fmtlen0+fmtlen1+fmtlen2+fmtlen3+fmtlen4+28; i++ {
					if i == 3+fmtlen0+1 || i == 3+fmtlen0+3+fmtlen1+1 || i == 3+fmtlen0+3+fmtlen1+4+fmtlen2 || i == 3+fmtlen0+3+fmtlen1+4+fmtlen2+2+fmtlen3+1 || i == 3+fmtlen0+3+fmtlen1+4+fmtlen2+2+fmtlen3+3+fmtlen4+1 {
						fmt.Printf("+")
					} else {
						fmt.Printf("-")
					}
				}
				fmt.Printf("\n")
			} else {
				// simulate
				fmt.Printf(format, "Parameter", "Value set", "Value expected", "Override", "Comment")
				for i := 0; i < fmtlen1+fmtlen2+fmtlen3+fmtlen4+28; i++ {
					if i == 3+fmtlen1+1 || i == 3+fmtlen1+3+fmtlen2+1 || i == 3+fmtlen1+3+fmtlen2+3+fmtlen3+1 || i == 3+fmtlen1+3+fmtlen2+3+fmtlen3+3+fmtlen4+1 {
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
			fmt.Printf(format, comparison.ReflectMapKey, strings.Replace(comparison.ActualValueJS, "\t", " ", -1), strings.Replace(comparison.ExpectedValueJS, "\t", " ", -1), override, comment)
		}
	}
	// print footer
	if header != "NONE" && !hasDiff {
		fmt.Printf("\n   (no change)\n")
	}
	for _, fn := range footnote {
		if fn != "" {
			fmt.Printf("\n %s", fn)
		}
	}
	fmt.Printf("\n\n")
	for noteID, reminde := range reminder {
		if reminde != "" {
			reminderHead := fmt.Sprintf("Attention for SAP Note %s:\nHints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:\n", noteID)
			fmt.Printf("%s\n", setRedText+reminderHead+reminde+resetTextColor)
		}
	}
}

// VerifyAllParameters Verify that all system parameters do not deviate from any of the enabled solutions/notes.
func VerifyAllParameters() {
	if len(tuneApp.NoteApplyOrder) == 0 {
		fmt.Println("No notes or solutions enabled, nothing to verify.")
	} else {
		unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
		if err != nil {
			errorExit("Failed to inspect the current system: %v", err)
		}
		PrintNoteFields("NONE", comparisons, true)
		tuneApp.PrintNoteApplyOrder()
		if len(unsatisfiedNotes) == 0 {
			fmt.Println("The running system is currently well-tuned according to all of the enabled notes.")
		} else {
			errorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
		}
	}
}

// NoteAction  Note actions like apply, revert, verify asm.
func NoteAction(actionName, noteID string) {
	switch actionName {
	case "apply":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		// Do not apply the note, if it was applied before
		// Otherwise, the state file (serialised parameters) will be
		// overwritten, and it will no longer be possible to revert the
		// note to the state before it was tuned.
		_, err := os.Stat(tuneApp.State.GetPathToNote(noteID))
		if err == nil {
			// state file for note already exists
			// do not apply the note again
			system.InfoLog("note '%s' already applied. Nothing to do", noteID)
			os.Exit(0)
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
		fmt.Println("\nAll notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note):")
		solutionNoteIDs := tuneApp.GetSortedSolutionEnabledNotes()
		for _, noteID := range tuningOptions.GetSortedIDs() {
			noteObj := tuningOptions[noteID]
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
			fmt.Printf(format, noteID, noteObj.Name())
		}
		tuneApp.PrintNoteApplyOrder()
		if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
			fmt.Println("Remember: if you wish to automatically activate the solution's tuning options after a reboot," +
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
			noteComp := make(map[string]map[string]note.FieldComparison)
			noteComp[noteID] = comparisons
			PrintNoteFields("HEAD", noteComp, true)
			tuneApp.PrintNoteApplyOrder()
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
			noteComp := make(map[string]map[string]note.FieldComparison)
			noteComp[noteID] = comparisons
			PrintNoteFields("HEAD", noteComp, false)
		}
	case "customise":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		if _, err := tuneApp.GetNoteByID(noteID); err != nil {
			errorExit("%v", err)
		}
		editFileName := ""
		fileName := fmt.Sprintf("%s%s", NoteTuningSheets, noteID)
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			_, files := system.ListDir(ExtraTuningSheets, "")
			for _, f := range files {
				if strings.HasPrefix(f, noteID) {
					fileName = fmt.Sprintf("%s%s", ExtraTuningSheets, f)
				}
			}
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				errorExit("Note %s not found in %s or %s.", noteID, NoteTuningSheets, ExtraTuningSheets)
			} else if err != nil {
				errorExit("Failed to read file '%s' - %v", fileName, err)
			}
		} else if err != nil {
			errorExit("Failed to read file '%s' - %v", fileName, err)
		}
		ovFileName := fmt.Sprintf("%s%s", OverrideTuningSheets, noteID)
		if _, err := os.Stat(ovFileName); os.IsNotExist(err) {
			//copy file
			src, err := os.Open(fileName)
			if err != nil {
				errorExit("Can not open file '%s' - %v", fileName, err)
			}
			defer src.Close()
			dst, err := os.OpenFile(ovFileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				errorExit("Can not create file '%s' - %v", ovFileName, err)
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			if err != nil {
				errorExit("Problems while copying '%s' to '%s' - %v", fileName, ovFileName, err)
			}
			editFileName = ovFileName
		} else if err == nil {
			system.InfoLog("Note override file already exists, using file '%s' as base for editing\n", ovFileName)
			editFileName = ovFileName
		} else {
			errorExit("Failed to read file '%s' - %v", ovFileName, err)
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "/usr/bin/vim" // launch vim by default
		}
		//if err := syscall.Exec(editor, []string{editor, fileName}, os.Environ()); err != nil {
		if err := syscall.Exec(editor, []string{editor, editFileName}, os.Environ()); err != nil {
			errorExit("Failed to start launch editor %s: %v", editor, err)
		}
	case "create":
		if noteID == "" {
			PrintHelpAndExit(1)
		}
		if _, err := tuneApp.GetNoteByID(noteID); err == nil {
			errorExit("Note '%s' already exists. Please use 'saptune note customise %s' instead to create an override file or choose another NoteID.", noteID, noteID)
		}
		fileName := fmt.Sprintf("%s%s", NoteTuningSheets, noteID)
		if _, err := os.Stat(fileName); err == nil {
			errorExit("Note '%s' already exists in %s. Please use 'saptune note customise %s' instead to create an override file or choose another NoteID.", noteID, NoteTuningSheets, noteID)
		}
		extraFileName := fmt.Sprintf("%s%s.conf", ExtraTuningSheets, noteID)
		if _, err := os.Stat(extraFileName); err == nil {
			errorExit("Note '%s' already exists in %s. Please use 'saptune note customise %s' instead to create an override file or choose another NoteID.", noteID, ExtraTuningSheets, noteID)
		}
		templateFile := "/usr/share/saptune/NoteTemplate.conf"
		//if _, err := os.Stat(extraFileName); os.IsNotExist(err) {
		//copy template file
		src, err := os.Open(templateFile)
		if err != nil {
			errorExit("Can not open file '%s' - %v", templateFile, err)
		}
		defer src.Close()
		dst, err := os.OpenFile(extraFileName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			errorExit("Can not create file '%s' - %v", extraFileName, err)
		}
		defer dst.Close()
		_, err = io.Copy(dst, src)
		if err != nil {
			errorExit("Problems while copying '%s' to '%s' - %v", templateFile, extraFileName, err)
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "/usr/bin/vim" // launch vim by default
		}
		if err := syscall.Exec(editor, []string{editor, extraFileName}, os.Environ()); err != nil {
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

// SolutionAction  Solution actions like apply, revert, verify asm.
func SolutionAction(actionName, solName string) {
	switch actionName {
	case "apply":
		if solName == "" {
			PrintHelpAndExit(1)
		}
		if len(tuneApp.TuneForSolutions) > 0 {
			// already one solution applied.
			// do not apply another solution. Does not make sense
			system.InfoLog("There is already one solution applied. Applying another solution is NOT supported.")
			os.Exit(0)
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
		fmt.Println("\nAll solutions (* denotes enabled solution, O denotes override file exists for solution, D denotes deprecated solutions):")
		for _, solName := range solution.GetSortedSolutionNames(solutionSelector) {
			format := "\t%-18s -"
			if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
				format = " " + setGreenText + "*" + format
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
			format = format + solNotes + resetTextColor + "\n"
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
			PrintNoteFields("NONE", comparisons, false)
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
