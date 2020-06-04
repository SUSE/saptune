package main

import (
	"bufio"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
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
	logFile               = "/var/log/tuned/tuned.log"
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
	footnote5             = "[5] expected value does not contain a supported scheduler"
	footnote6             = "[6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details"
	footnote7             = "[7] parameter value is untouched by default"
)

// PrintHelpAndExit Print the usage and exit
func PrintHelpAndExit(exitStatus int) {
	fmt.Println(`saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | enabled ]
  saptune note [ apply | simulate | verify | customise | create | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled ]
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
	_ = system.ErrorLog(template+"\n", stuff...)
	os.Exit(exState)
}

// Return the i-th command line parameter, or empty string if it is not specified.
func cliArg(i int) string {
	if len(os.Args) >= i+1 {
		return os.Args[i]
	}
	return ""
}

var tuneApp *app.App                             // application configuration and tuning states
var tuningOptions note.TuningOptions             // Collection of tuning options from SAP notes and 3rd party vendors.
var footnote1 = footnote1X86                     // set 'unsupported' footnote regarding the architecture
var debugSwitch = os.Getenv("SAPTUNE_DEBUG")     // Switch Debug on ("1") or off ("0" - default)
var verboseSwitch = os.Getenv("SAPTUNE_VERBOSE") // Switch verbose mode on ("on" - default) or off ("off")
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
	// check, if DEBUG is set in /etc/sysconfig/saptune
	if debugSwitch == "" {
		debugSwitch = sconf.GetString("DEBUG", "0")
	}
	if verboseSwitch == "" {
		verboseSwitch = sconf.GetString("VERBOSE", "on")
	}

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

	// cleanup runtime file
	note.CleanUpRun()

	// activate logging
	system.LogInit(logFile, debugSwitch, verboseSwitch)

	switch saptuneVersion {
	case "1":
		cmd := exec.Command(saptuneV1, os.Args[1:]...)
		cmd.Stdin = os.Stdin
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
	selectAction()

}

// selectAction selects the choosen action depending on the first command line
// argument
func selectAction() {
	switch cliArg(1) {
	case "daemon":
		DaemonAction(cliArg(2))
	case "note":
		NoteAction(cliArg(2), cliArg(3), cliArg(4))
	case "solution":
		SolutionAction(cliArg(2), cliArg(3))
	case "revert":
		RevertAction(os.Stdout, cliArg(2), tuneApp)
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
	if tuneApp != nil && (len(tuneApp.NoteApplyOrder) == 0 && (len(tuneApp.TuneForNotes) != 0 || len(tuneApp.TuneForSolutions) != 0)) {
		errorExit("There are 'old' solutions or notes defined in file '/etc/sysconfig/saptune'. Seems there were some steps missed during the migration from saptune version 1 to version 2. Please check. Refer to saptune-migrate(7) for more information")
	}
}

// RevertAction Revert all notes and solutions
func RevertAction(writer io.Writer, actionName string, tuneApp *app.App) {
	if actionName != "all" {
		PrintHelpAndExit(1)
	}
	fmt.Fprintf(writer, "Reverting all notes and solutions, this may take some time...\n")
	if err := tuneApp.RevertAll(true); err != nil {
		errorExit("Failed to revert notes: %v", err)
		//panic(err)
	}
	fmt.Fprintf(writer, "Parameters tuned by the notes and solutions have been successfully reverted.\n")
}

// DaemonAction handles daemon actions like start, stop, status asm.
func DaemonAction(actionName string) {
	switch actionName {
	case "start":
		DaemonActionStart()
	case "apply":
		// This action name is only used by tuned script, hence it is not advertised to end user.
		if err := tuneApp.TuneAll(); err != nil {
			panic(err)
		}
	case "status":
		DaemonActionStatus()
	case "stop":
		DaemonActionStop()
	case "revert":
		// This action name is only used by tuned script, hence it is not advertised to end user.
		if err := tuneApp.RevertAll(false); err != nil {
			panic(err)
		}
	default:
		PrintHelpAndExit(1)
	}
}

// DaemonActionStart starts the tuned service
func DaemonActionStart() {
	fmt.Println("Starting daemon (tuned.service), this may take several seconds...")
	if system.IsServiceAvailable(SapconfService) {
		if err := system.SystemctlDisableStop(SapconfService); err != nil {
			errorExit("%v", err)
		}
	}
	if err := system.TunedAdmProfile("saptune"); err != nil {
		errorExit("%v", err)
	}
	if err := system.SystemctlEnableStart(TunedService); err != nil {
		errorExit("%v", err)
	}
	// Check tuned profile
	if system.GetTunedAdmProfile() != TunedProfileName {
		_ = system.ErrorLog("tuned.service profile is incorrect. Please check tuned logs for more information")
		// defined exit value needed for yast module
		os.Exit(exitTunedWrongProfile)
	}
	// tuned then calls `saptune daemon apply`
	fmt.Println("Daemon (tuned.service) has been enabled and started.")
	if len(tuneApp.TuneForSolutions) == 0 && len(tuneApp.TuneForNotes) == 0 {
		fmt.Println("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
	}
}

// DaemonActionStatus checks the status of the tuned service
func DaemonActionStatus() {
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
}

// DaemonActionStop stops the tuned service
func DaemonActionStop() {
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
}

// PrintNoteFields Print mismatching fields in the note comparison result.
//func PrintNoteFields(header string, noteComparisons map[string]map[string]note.FieldComparison, printComparison bool) {
func PrintNoteFields(writer io.Writer, header string, noteComparisons map[string]map[string]note.FieldComparison, printComparison bool) {

	// initialise
	compliant := "yes"
	printHead := ""
	noteField := ""
	footnote := make([]string, 7, 7)
	reminder := make(map[string]string)
	override := ""
	comment := ""
	hasDiff := false

	// sort output
	sortkeys := sortNoteComparisonsOutput(noteComparisons)

	// setup table format values
	fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4, format := setupTableFormat(sortkeys, noteField, noteComparisons, printComparison)

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
			//noteField = fmt.Sprintf("%s, %s", noteID, txtparser.GetINIFileVersion(noteComparisons[noteID]["ConfFilePath"].ActualValue.(string)))
			noteField = fmt.Sprintf("%s, %s", noteID, txtparser.GetINIFileVersionSectionEntry(noteComparisons[noteID]["ConfFilePath"].ActualValue.(string), "version"))
		}

		override = strings.Replace(noteComparisons[noteID][fmt.Sprintf("%s[%s]", "OverrideParams", key)].ExpectedValueJS, "\t", " ", -1)
		comparison := noteComparisons[noteID][fmt.Sprintf("%s[%s]", "SysctlParams", key)]
		if comparison.ReflectMapKey == "reminder" {
			reminder[noteID] = reminder[noteID] + comparison.ExpectedValueJS
			continue
		}
		if !comparison.MatchExpectation {
			if comparison.ExpectedValue.(string) == "" {
				compliant = "yes"
			} else {
				hasDiff = true
				compliant = "no "
			}
		} else {
			compliant = "yes"
		}
		if comparison.ActualValue.(string) == "all:none" {
			compliant = " - "
		}

		// check inform map for special settings
		inform := ""
		if noteComparisons[noteID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ActualValue != nil {
			inform = noteComparisons[noteID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ActualValue.(string)
			if inform == "" && noteComparisons[noteID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ExpectedValue != nil {
				inform = noteComparisons[noteID][fmt.Sprintf("%s[%s]", "Inform", comparison.ReflectMapKey)].ExpectedValue.(string)
			}
		}

		// prepare footnote
		compliant, comment, footnote = prepareFootnote(comparison, compliant, comment, inform, footnote)

		// print table header
		if printHead != "" {
			printHeadline(writer, header, noteID, tuningOptions)
			printTableHeader(writer, format, fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4, printComparison)
		}

		// print table body
		if printComparison {
			// verify
			fmt.Fprintf(writer, format, noteField, comparison.ReflectMapKey, strings.Replace(comparison.ExpectedValueJS, "\t", " ", -1), override, strings.Replace(comparison.ActualValueJS, "\t", " ", -1), compliant)
		} else {
			// simulate
			fmt.Fprintf(writer, format, comparison.ReflectMapKey, strings.Replace(comparison.ActualValueJS, "\t", " ", -1), strings.Replace(comparison.ExpectedValueJS, "\t", " ", -1), override, comment)
		}
	}
	// print footer
	printTableFooter(writer, header, footnote, reminder, hasDiff)
}

// sortNoteComparisonsOutput sorts the output of the Note comparison
// the reminder section should be the last one
func sortNoteComparisonsOutput(noteCompare map[string]map[string]note.FieldComparison) []string {
	skeys := make([]string, 0, len(noteCompare))
	rkeys := make([]string, 0, len(noteCompare))
	// sort output
	for noteID, comparisons := range noteCompare {
		for _, comparison := range comparisons {
			if comparison.ReflectFieldName == "Inform" {
				// skip inform map to avoid double entries in verify table
				continue
			}
			if len(comparison.ReflectMapKey) != 0 && comparison.ReflectFieldName != "OverrideParams" {
				if comparison.ReflectMapKey != "reminder" {
					skeys = append(skeys, noteID+"§"+comparison.ReflectMapKey)
				} else {
					rkeys = append(rkeys, noteID+"§"+comparison.ReflectMapKey)
				}
			}
		}
	}
	sort.Strings(skeys)
	for _, rem := range rkeys {
		skeys = append(skeys, rem)
	}
	return skeys
}

// setupTableFormat sets the format of the table columns dependent on the content
func setupTableFormat(skeys []string, noteField string, noteCompare map[string]map[string]note.FieldComparison, printComp bool) (int, int, int, int, int, string) {
	var fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4 int
	format := "\t%s : %s\n"
	// define start values for the column width
	if printComp {
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

	for _, skey := range skeys {
		keyFields := strings.Split(skey, "§")
		noteID := keyFields[0]
		comparisons := noteCompare[noteID]
		for _, comparison := range comparisons {
			if comparison.ReflectMapKey == "reminder" {
				continue
			}
			if printComp {
				// verify
				if len(noteField) > fmtlen0 {
					fmtlen0 = len(noteField)
				}
				// 3:override, 1:mapkey, 2:expval, 4:actval
				fmtlen3, fmtlen1, fmtlen2, fmtlen4 = setWidthOfColums(comparison, fmtlen3, fmtlen1, fmtlen2, fmtlen4)
				format = "   %-" + strconv.Itoa(fmtlen0) + "s | %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %2s\n"
			} else {
				// simulate
				// 4:override, 1:mapkey, 3:expval, 2:actval
				fmtlen4, fmtlen1, fmtlen3, fmtlen2 = setWidthOfColums(comparison, fmtlen4, fmtlen1, fmtlen3, fmtlen2)
				format = "   %-" + strconv.Itoa(fmtlen1) + "s | %-" + strconv.Itoa(fmtlen2) + "s | %-" + strconv.Itoa(fmtlen3) + "s | %-" + strconv.Itoa(fmtlen4) + "s | %2s\n"
			}
		}
	}
	return fmtlen0, fmtlen1, fmtlen2, fmtlen3, fmtlen4, format
}

// printHeadline prints a headline for the table
func printHeadline(writer io.Writer, header, id string, tuningOpts note.TuningOptions) {
	if header != "NONE" {
		nName := ""
		if len(tuningOpts) > 0 {
			nName = tuningOpts[id].Name()
		}
		fmt.Fprintf(writer, "\n%s - %s \n\n", id, nName)
	}
}

// printTableHeader prints the header of the table
func printTableHeader(writer io.Writer, format string, col0, col1, col2, col3, col4 int, printComp bool) {
	if printComp {
		// verify
		fmt.Fprintf(writer, format, "SAPNote, Version", "Parameter", "Expected", "Override", "Actual", "Compliant")
		for i := 0; i < col0+col1+col2+col3+col4+28; i++ {
			if i == 3+col0+1 || i == 3+col0+3+col1+1 || i == 3+col0+3+col1+4+col2 || i == 3+col0+3+col1+4+col2+2+col3+1 || i == 3+col0+3+col1+4+col2+2+col3+3+col4+1 {
				fmt.Fprintf(writer, "+")
			} else {
				fmt.Fprintf(writer, "-")
			}
		}
		fmt.Fprintf(writer, "\n")
	} else {
		// simulate
		fmt.Fprintf(writer, format, "Parameter", "Value set", "Value expected", "Override", "Comment")
		for i := 0; i < col1+col2+col3+col4+28; i++ {
			if i == 3+col1+1 || i == 3+col1+3+col2+1 || i == 3+col1+3+col2+3+col3+1 || i == 3+col1+3+col2+3+col3+3+col4+1 {
				fmt.Fprintf(writer, "+")
			} else {
				fmt.Fprintf(writer, "-")
			}
		}
		fmt.Fprintf(writer, "\n")
	}
}

// prepareFootnote prepares the content of the last column and the
// corresponding footnotes
func prepareFootnote(comparison note.FieldComparison, compliant, comment, inform string, footnote []string) (string, string, []string) {
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
	if comparison.ReflectMapKey == "force_latency" && inform == "hasDiffs" {
		compliant = "no [4]"
		comment = comment + " [4]"
		footnote[3] = footnote4
	}
	var isSched = regexp.MustCompile(`^IO_SCHEDULER_\w+$`)
	if isSched.MatchString(comparison.ReflectMapKey) && inform == "NA" {
		compliant = compliant + " [5]"
		comment = comment + " [5]"
		footnote[4] = footnote5
	}
	if strings.Contains(comparison.ReflectMapKey, "grub") {
		compliant = compliant + " [6]"
		comment = comment + " [6]"
		footnote[5] = footnote6
	}
	if comparison.ExpectedValue == "" {
		compliant = compliant + " [7]"
		comment = comment + " [7]"
		footnote[6] = footnote7
	}
	return compliant, comment, footnote
}

// printTableFooter prints the footer of the table
// footnotes and reminder section
func printTableFooter(writer io.Writer, header string, footnote []string, reminder map[string]string, hasDiff bool) {
	if header != "NONE" && !hasDiff {
		fmt.Fprintf(writer, "\n   (no change)\n")
	}
	for _, fn := range footnote {
		if fn != "" {
			fmt.Fprintf(writer, "\n %s", fn)
		}
	}
	fmt.Fprintf(writer, "\n\n")
	for noteID, reminde := range reminder {
		if reminde != "" {
			reminderHead := fmt.Sprintf("Attention for SAP Note %s:\nHints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:\n", noteID)
			fmt.Fprintf(writer, "%s\n", setRedText+reminderHead+reminde+resetTextColor)
		}
	}
}

// setWidthOfColums sets the width of the columns for verify and simulate
// depending on the highest number of characters of the content to be
// displayed
// c1:override, c2:mapkey, c3:expval, c4:actval
func setWidthOfColums(compare note.FieldComparison, c1, c2, c3, c4 int) (int, int, int, int) {
	if len(compare.ReflectMapKey) != 0 {
		if compare.ReflectFieldName == "OverrideParams" && len(compare.ActualValueJS) > c1 {
			c1 = len(compare.ActualValueJS)
			return c1, c2, c3, c4
		}
		if len(compare.ReflectMapKey) > c2 {
			c2 = len(compare.ReflectMapKey)
		}
		if len(compare.ExpectedValueJS) > c3 {
			c3 = len(compare.ExpectedValueJS)
		}
		if len(compare.ActualValueJS) > c4 {
			c4 = len(compare.ActualValueJS)
		}
	}
	return c1, c2, c3, c4
}

// VerifyAllParameters Verify that all system parameters do not deviate from any of the enabled solutions/notes.
func VerifyAllParameters(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.NoteApplyOrder) == 0 {
		fmt.Fprintf(writer, "No notes or solutions enabled, nothing to verify.\n")
	} else {
		unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
		if err != nil {
			errorExit("Failed to inspect the current system: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true)
		tuneApp.PrintNoteApplyOrder(writer)
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "The running system is currently well-tuned according to all of the enabled notes.\n")
		} else {
			errorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
		}
	}
}

// NoteAction  Note actions like apply, revert, verify asm.
func NoteAction(actionName, noteID, newNoteID string) {
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
		NoteActionCustomise(noteID)
	case "create":
		NoteActionCreate(noteID)
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
		PrintHelpAndExit(1)
	}
}

// NoteActionApply applies Note parameter settings to the system
func NoteActionApply(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(1)
	}
	// Do not apply the note, if it was applied before
	// Otherwise, the state file (serialised parameters) will be
	// overwritten, and it will no longer be possible to revert the
	// note to the state before it was tuned.
	sfile, err := os.Stat(tuneApp.State.GetPathToNote(noteID))
	if err == nil {
		// state file for note already exists
		// check, if note is part of NOTE_APPLY_ORDER
		if tuneApp.PositionInNoteApplyOrder(noteID) < 0 { // noteID not yet available
			// bsc#1167618
			// check, if state file is empty - seems to be a
			// left-over of the update from saptune V1 to V2
			if sfile.Size() == 0 {
				// remove old, left-over state file and go
				// forward to apply the note
				os.Remove(tuneApp.State.GetPathToNote(noteID))
			} else {
				// data mismatch, do not apply the note
				system.WarningLog("note '%s' is not listed in 'NOTE_APPLY_ORDER', but a non-empty state file exists. To prevent configuration mismatch, please revert note '%s' first and apply again.", noteID, noteID)
				os.Exit(0)
			}
		} else {
			system.InfoLog("note '%s' already applied. Nothing to do", noteID)
			os.Exit(0)
		}
	}
	if err := tuneApp.TuneNote(noteID); err != nil {
		errorExit("Failed to tune for note %s: %v", noteID, err)
	}
	fmt.Fprintf(writer, "The note has been applied successfully.\n")
	if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
		fmt.Fprintf(writer, "\nRemember: if you wish to automatically activate the solution's tuning options after a reboot,"+
			"you must instruct saptune to configure \"tuned\" daemon by running:"+
			"\n    saptune daemon start\n")
	}
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
	if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
		fmt.Fprintf(writer, "Remember: if you wish to automatically activate the solution's tuning options after a reboot,"+
			"you must instruct saptune to configure \"tuned\" daemon by running:"+
			"\n    saptune daemon start\n")
	}
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
			errorExit("Failed to test the current system against the specified note: %v", err)
		}
		noteComp := make(map[string]map[string]note.FieldComparison)
		noteComp[noteID] = comparisons
		PrintNoteFields(writer, "HEAD", noteComp, true)
		tuneApp.PrintNoteApplyOrder(writer)
		if !conforming {
			errorExit("The parameters listed above have deviated from the specified note.\n")
		} else {
			fmt.Fprintf(writer, "The system fully conforms to the specified note.\n")
		}
	}
}

// NoteActionSimulate shows all changes that will be applied to the system if
// the Note will be applied.
func NoteActionSimulate(writer io.Writer, noteID string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(1)
	}
	// Run verify and print out all fields of the note
	if _, comparisons, _, err := tuneApp.VerifyNote(noteID); err != nil {
		errorExit("Failed to test the current system against the specified note: %v", err)
	} else {
		fmt.Fprintf(writer, "If you run `saptune note apply %s`, the following changes will be applied to your system:\n", noteID)
		noteComp := make(map[string]map[string]note.FieldComparison)
		noteComp[noteID] = comparisons
		PrintNoteFields(writer, "HEAD", noteComp, false)
	}
}

// NoteActionCustomise creates an override file and allows to editing the Note
// definition file
func NoteActionCustomise(noteID string) {
	if noteID == "" {
		PrintHelpAndExit(1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		errorExit("%v", err)
	}
	editFileName := ""
	fileName, _ := getFileName(noteID, NoteTuningSheets, ExtraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, OverrideTuningSheets)
	if !overrideNote {
		//copy file
		err := system.CopyFile(fileName, ovFileName)
		if err != nil {
			errorExit("Problems while copying '%s' to '%s' - %v", fileName, ovFileName, err)
		}
		editFileName = ovFileName
	} else {
		system.InfoLog("Note override file already exists, using file '%s' as base for editing", ovFileName)
		editFileName = ovFileName
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "/usr/bin/vim" // launch vim by default
	}
	i := tuneApp.PositionInNoteApplyOrder(noteID)
	if i < 0 { // noteID not yet available
		system.InfoLog("Do not forget to apply the just edited Note to get your changes to take effect\n")
	} else { // noteID already applied
		system.InfoLog("Your just edited Note is already applied. To get your changes to take effect, please 'revert' the Note and apply again.\n")
	}
	if err := syscall.Exec(editor, []string{editor, editFileName}, os.Environ()); err != nil {
		errorExit("Failed to start launch editor %s: %v", editor, err)
	}
	// if syscall.Exec returns 'nil' the execution of the program ends immediately
}

// NoteActionCreate helps the customer to create an own Note definition
func NoteActionCreate(noteID string) {
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
	err := system.CopyFile(templateFile, extraFileName)
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
}

// NoteActionShow shows the content of the Note definition file
func NoteActionShow(writer io.Writer, noteID, noteTuningSheets, extraTuningSheets string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		errorExit("%v", err)
	}
	fileName, _ := getFileName(noteID, noteTuningSheets, extraTuningSheets)
	cont, err := ioutil.ReadFile(fileName)
	if err != nil {
		errorExit("Failed to read file '%s' - %v", fileName, err)
	}
	fmt.Fprintf(writer, "\nContent of Note %s:\n%s\n", noteID, string(cont))
}

// NoteActionDelete deletes a custom Note definition file and
// the corresponding override file
func NoteActionDelete(reader io.Reader, writer io.Writer, noteID, noteTuningSheets, extraTuningSheets, ovTuningSheets string, tuneApp *app.App) {
	if noteID == "" {
		PrintHelpAndExit(1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		errorExit("%v", err)
	}

	txtConfirm := fmt.Sprintf("Do you really want to delete Note (%s)?", noteID)
	fileName, extraNote := getFileName(noteID, noteTuningSheets, extraTuningSheets)
	ovFileName, overrideNote := getovFile(noteID, ovTuningSheets)

	// check, if note is active - applied
	i := tuneApp.PositionInNoteApplyOrder(noteID)
	if i >= 0 { // noteID already applied
		system.InfoLog("The Note definition file you want to delete is currently in use, which means it is already applied.")
		system.InfoLog("So please 'revert' the Note first and then try deleting again.\n")
		os.Exit(0)
	}

	if !extraNote && !overrideNote {
		errorExit("ATTENTION: The Note definition file you want to delete is a saptune internal (shipped) Note and can NOT be deleted. Exiting ...")
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
		PrintHelpAndExit(1)
	}
	if _, err := tuneApp.GetNoteByID(noteID); err != nil {
		errorExit("%v", err)
	}
	if _, err := tuneApp.GetNoteByID(newNoteID); err == nil {
		errorExit("The new name '%s' for Note %s already exists, can't rename.", noteID, newNoteID)
	}

	txtConfirm := fmt.Sprintf("Do you really want to rename Note %s to %s?", noteID, newNoteID)
	fileName, extraNote := getFileName(noteID, noteTuningSheets, extraTuningSheets)
	newFileName := fmt.Sprintf("%s%s.conf", extraTuningSheets, newNoteID)
	if !extraNote {
		errorExit("The Note definition file you want to rename is a saptune internal (shipped) Note and can NOT be renamed. Exiting ...")
	}
	ovFileName, overrideNote := getovFile(noteID, ovTuningSheets)
	newovFileName := fmt.Sprintf("%s%s", ovTuningSheets, newNoteID)

	// check, if note is active - applied
	i := tuneApp.PositionInNoteApplyOrder(noteID)
	if i >= 0 { // noteID already applied
		system.InfoLog("The Note definition file you want to rename is currently in use, which means it is already applied.")
		system.InfoLog("So please 'revert' the Note first and then try renaming again.\n")
		os.Exit(0)
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
		PrintHelpAndExit(1)
	}
	if err := tuneApp.RevertNote(noteID, true); err != nil {
		errorExit("Failed to revert note %s: %v", noteID, err)
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

// SolutionAction  Solution actions like apply, revert, verify asm.
func SolutionAction(actionName, solName string) {
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
		os.Exit(1)
	}
	removedAdditionalNotes, err := tuneApp.TuneSolution(solName)
	if err != nil {
		errorExit("Failed to tune for solution %s: %v", solName, err)
	}
	fmt.Fprintf(writer, "All tuning options for the SAP solution have been applied successfully.\n")
	if len(removedAdditionalNotes) > 0 {
		fmt.Fprintf(writer, "\nThe following previously-enabled notes are now tuned by the SAP solution:\n")
		for _, noteNumber := range removedAdditionalNotes {
			fmt.Fprintf(writer, "\t%s\t%s\n", noteNumber, tuneApp.AllNotes[noteNumber].Name())
		}
	}
	if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
		fmt.Fprintf(writer, "\nRemember: if you wish to automatically activate the solution's tuning options after a reboot,"+
			"you must instruct saptune to configure \"tuned\" daemon by running:"+
			"\n    saptune daemon start\n")
	}
}

// SolutionActionList lists all available solution definitions
func SolutionActionList(writer io.Writer, tuneApp *app.App) {
	setColor := false
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
	if !system.SystemctlIsRunning(TunedService) || system.GetTunedProfile() != TunedProfileName {
		fmt.Fprintf(writer, "\nRemember: if you wish to automatically activate the solution's tuning options after a reboot,"+
			"you must instruct saptune to configure \"tuned\" daemon by running:"+
			"\n    saptune daemon start\n")
	}
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
			errorExit("Failed to test the current system against the specified SAP solution: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true)
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "The system fully conforms to the tuning guidelines of the specified SAP solution.\n")
		} else {
			errorExit("The parameters listed above have deviated from the specified SAP solution recommendations.\n")
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
		errorExit("Failed to test the current system against the specified note: %v", err)
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
		errorExit("Failed to revert tuning for solution %s: %v", solName, err)
	}
	fmt.Fprintf(writer, "Parameters tuned by the notes referred by the SAP solution have been successfully reverted.\n")
}

// SolutionActionEnabled prints out the enabled solution definition
func SolutionActionEnabled(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.TuneForSolutions) != 0 {
		fmt.Fprintf(writer, "%s", tuneApp.TuneForSolutions[0])
	}
}

// getFileName returns the corresponding filename of a given noteID
// additional it returns a boolean value which is pointing out that the Note
// the Note is a custom Note (extraNote = true) or an internal one
func getFileName(noteID, NoteTuningSheets, ExtraTuningSheets string) (string, bool) {
	extraNote := false
	fileName := fmt.Sprintf("%s%s", NoteTuningSheets, noteID)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		// Note is NOT an internal Note, but may be a custom Note
		extraNote = true
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
	return fileName, extraNote
}

// getovFile returns the corresponding override filename of a given noteID
// additional it returns a boolean value which is pointing out if the
// override file already exists (overrideNote = true) or not
func getovFile(noteID, OverrideTuningSheets string) (string, bool) {
	overrideNote := true
	ovFileName := fmt.Sprintf("%s%s", OverrideTuningSheets, noteID)
	if _, err := os.Stat(ovFileName); os.IsNotExist(err) {
		overrideNote = false
	} else if err != nil {
		errorExit("Failed to read file '%s' - %v", ovFileName, err)
	}
	return ovFileName, overrideNote
}

// readYesNo asks the user for yes/no answer.
// "y", "Y", "yes", "YES", and "Yes" following by "enter" count as confirmation
// "n", "N", "no", "NO", and "No" following by "enter" count as non-confirmation
func readYesNo(s string, in io.Reader, out io.Writer) bool {
	reader := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "%s [y/n]: ", s)
		response, err := reader.ReadString('\n')
		if err != nil {
			errorExit("Failed to read input: %v", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// renameNote will rename a Note to an new name
func renameNote(fileName, newFileName, ovFileName, newovFileName string, overrideNote, extraNote bool) {
	if overrideNote {
		if err := os.Rename(ovFileName, newovFileName); err != nil {
			errorExit("Failed to rename file '%s' to '%s' - %v", ovFileName, newovFileName, err)
		}
	}
	if extraNote {
		if err := os.Rename(fileName, newFileName); err != nil {
			errorExit("Failed to rename file '%s' to '%s' - %v", fileName, newFileName, err)
		}
	}
}

// deleteNote will delete a Note
func deleteNote(fileName, ovFileName string, overrideNote, extraNote bool) {
	if overrideNote {
		if err := os.Remove(ovFileName); err != nil {
			errorExit("Failed to remove file '%s' - %v", ovFileName, err)
		}
	}
	if extraNote {
		if err := os.Remove(fileName); err != nil {
			errorExit("Failed to remove file '%s' - %v", fileName, err)
		}
	}
}
