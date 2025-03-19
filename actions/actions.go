package actions

import (
	"bufio"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"strings"
)

// define constants and variables for the whole package
const (
	SaptuneService     = "saptune.service"
	SapconfService     = "sapconf.service"
	TunedService       = "tuned.service"
	exitSaptuneStopped = 1
	exitNotTuned       = 3
	exitNotCompliant   = 4
)

// PackageArea is the package area with all notes and solutions shipped by
// the current installed saptune rpm
var PackageArea = "/usr/share/saptune/"

// WorkingArea is the working directory
var WorkingArea = "/var/lib/saptune/working/"

// StagingArea is the staging area
var StagingArea = "/var/lib/saptune/staging/"

// StagingSheets is the staging directory of the latest notes
var StagingSheets = "/var/lib/saptune/staging/latest/"

// NoteTuningSheets is the working directory of available sap notes
var NoteTuningSheets = "/var/lib/saptune/working/notes/"

// OverrideTuningSheets is the directory for the override files
var OverrideTuningSheets = "/etc/saptune/override/"

// ExtraTuningSheets is a directory located on file system for external parties to place their tuning option files.
var ExtraTuningSheets = "/etc/saptune/extra/"

// DeprecationSheets s the directory for the deprecated Notes and Solutions
var DeprecationSheets = "/usr/share/saptune/deprecated/"

// SolutionSheets is the working directory of available sap solutions
var SolutionSheets = "/var/lib/saptune/working/sols/"

// RPMVersion is the package version from package build process
var RPMVersion = "undef"

// RPMDate is the date of package build
// only used in individual build test packages, but NOT in our official
// built and released packages (not possible because of 'reproducible' builds)
var RPMDate = "undef"

// solutionSelector used in solutionacts and stagingacts
var solutionSelector = system.GetSolutionSelector()

// saptune configuration file
var saptuneSysconfig = system.SaptuneConfigFile()

// set colors for the table and list output
// var setYellowText = "\033[38;5;220m"
// var setCyanText = "\033[36m"
// var setUnderlinedText = "\033[4m"
var setGreenText = "\033[32m"
var setRedText = "\033[31m"
var setYellowText = "\033[33m"
var setBlueText = "\033[34m"
var setBoldText = "\033[1m"
var resetBoldText = "\033[22m"
var setStrikeText = "\033[9m"
var resetTextColor = "\033[0m"
var dfltColorScheme = "full-red-noncmpl"

// SelectAction selects the chosen action depending on the first command line
// argument
func SelectAction(writer io.Writer, stApp *app.App, saptuneVers string) {
	// switch off color and highlighting, if Stdout is not a terminal
	switchOffColor()
	system.JnotSupportedYet()

	// check for test packages
	if RPMDate != "undef" {
		system.NoticeLog("ATTENTION: You are running a test version (%s for SLES4SAP %d from %s) of saptune which is not supported for production use", RPMVersion, system.IfdefVers(), RPMDate)
	}

	switch system.CliArg(1) {
	case "daemon":
		DaemonAction(writer, system.CliArg(2), saptuneVers, stApp)
	case "service":
		ServiceAction(writer, system.CliArg(2), saptuneVers, stApp)
	case "note":
		NoteAction(writer, system.CliArg(2), system.CliArg(3), system.CliArg(4), stApp)
	case "solution":
		SolutionAction(writer, system.CliArg(2), system.CliArg(3), system.CliArg(4), stApp)
	case "configure":
		ConfigureAction(writer, system.CliArg(2), system.CliArgs(3), stApp)
	case "refresh":
		RefreshAction(writer, system.CliArg(2), stApp)
	case "revert":
		RevertAction(writer, system.CliArg(2), stApp)
	case "staging":
		StagingAction(system.CliArg(2), system.CliArgs(3), stApp)
	case "status":
		if system.CliArg(2) != "" {
			PrintHelpAndExit(writer, 1)
		}
		ServiceAction(writer, "status", saptuneVers, stApp)
	case "verify":
		VerifyAction(writer, system.CliArg(2), stApp)
	default:
		PrintHelpAndExit(writer, 1)
	}
}

// RefreshAction refreshes all applied Notes
func RefreshAction(writer io.Writer, actionName string, tuneApp *app.App) {
	if actionName != "applied" {
		PrintHelpAndExit(writer, 1)
	}
	NoteActionRefresh(writer, "", tuneApp)
}

// VerifyAction verifies all applied Notes
func VerifyAction(writer io.Writer, actionName string, tuneApp *app.App) {
	if actionName != "applied" {
		PrintHelpAndExit(writer, 1)
	}
	VerifyAllParameters(writer, tuneApp)
}

// RevertAction Revert all notes and solutions
func RevertAction(writer io.Writer, actionName string, tuneApp *app.App) {
	if actionName != "all" {
		PrintHelpAndExit(writer, 1)
	}
	reportSuc := false
	if len(tuneApp.NoteApplyOrder) != 0 {
		reportSuc = true
		system.InfoLog("Reverting all notes and solutions, this may take some time...")
		fmt.Fprintf(writer, "Reverting all notes and solutions, this may take some time...\n")
	}
	if err := tuneApp.RevertAll(true); err != nil {
		system.ErrorExit("Failed to revert notes: %v", err)
	}
	if reportSuc {
		system.InfoLog("Parameters tuned by the notes and solutions have been successfully reverted.")
		fmt.Fprintf(writer, "Parameters tuned by the notes and solutions have been successfully reverted.\n")
	}
}

// rememberMessage prints a reminder message
func rememberMessage(writer io.Writer) {
	active, err := system.SystemctlIsRunning(SaptuneService)
	if err != nil {
		system.ErrorExit("%v", err)
	}
	if !active {
		fmt.Fprintf(writer, "\nRemember: if you wish to automatically activate the solution's tuning options after a reboot, "+
			"you must enable and start saptune.service by running:"+
			"\n    saptune service enablestart\n")
	}
}

// VerifyAllParameters Verify that all system parameters do not deviate from any of the enabled solutions/notes.
func VerifyAllParameters(writer io.Writer, tuneApp *app.App) {
	result := system.JPNotes{
		Verifications: []system.JPNotesLine{},
		Attentions:    []system.JPNotesRemind{},
		NotesOrder:    []string{},
		SysCompliance: nil,
	}
	if len(tuneApp.NoteApplyOrder) == 0 {
		fmt.Fprintf(writer, "No notes or solutions enabled, nothing to verify.\n")
	} else {
		unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
		if err != nil {
			system.Jcollect(result)
			system.ErrorExit("Failed to inspect the current system: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true, &result)
		tuneApp.PrintNoteApplyOrder(writer)
		result.NotesOrder = tuneApp.NoteApplyOrder
		sysComp := len(unsatisfiedNotes) == 0
		result.SysCompliance = &sysComp
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "%s%sThe running system is currently well-tuned according to all of the enabled notes.%s%s\n", setGreenText, setBoldText, resetBoldText, resetTextColor)
		} else {
			system.Jcollect(result)
			system.ErrorExit("The parameters listed above have deviated from SAP/SUSE recommendations.", "colorPrint", setRedText, setBoldText, resetBoldText, resetTextColor)
		}
	}
	system.Jcollect(result)
}

// chkFileName returns the corresponding filename of a given definition file
// (note or solution)
// additional it returns a boolean value which is pointing out that
// the definition is a custom definition (extraDef = true) or an internal one
func chkFileName(defName, workingDir, extraDir string) (string, bool, error) {
	extraDef := false
	defType := "Note"
	if workingDir == SolutionSheets {
		defType = "Solution"
	}
	fileName := fmt.Sprintf("%s%s", workingDir, defName)
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		// Note/solution is NOT an internal Note/solution,
		// but may be a custom Note/solution
		chkName := defName
		if defType == "Note" {
			chkName = defName + ".conf"
		}
		if defName != "" {
			fileName = fmt.Sprintf("%s%s", extraDir, chkName)
		}
		if _, err = os.Stat(fileName); err == nil {
			extraDef = true
		}
	}
	return fileName, extraDef, err
}

// getFileName returns the corresponding filename of a given definition file
// (note or solution)
// additional it returns a boolean value which is pointing out that
// the definition is a custom definition (extraDef = true) or an internal one
func getFileName(defName, workingDir, extraDir string) (string, bool) {
	defType := "Note"
	if workingDir == SolutionSheets {
		defType = "Solution"
	}
	fileName, extraDef, err := chkFileName(defName, workingDir, extraDir)
	if os.IsNotExist(err) {
		system.ErrorExit("%s %s not found in %s or %s.", defType, defName, workingDir, extraDir)
	} else if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
	}
	return fileName, extraDef
}

// getovFile returns the corresponding override filename of a given
// definition (Note or solution)
// additional it returns a boolean value which is pointing out if the
// override file already exists (overrideNote = true) or not
func getovFile(defName, OverrideTuningSheets string) (string, bool) {
	overrideNote := true
	ovFileName := fmt.Sprintf("%s%s", OverrideTuningSheets, defName)
	if _, err := os.Stat(ovFileName); os.IsNotExist(err) {
		overrideNote = false
	} else if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", ovFileName, err)
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
			system.ErrorExit("Failed to read input: %v", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// renameDefFile will rename a definition file (Note or Solution) to an new name
func renameDefFile(fileName, newFileName string) {
	if err := os.Rename(fileName, newFileName); err != nil {
		system.ErrorExit("Failed to rename file '%s' to '%s' - %v", fileName, newFileName, err)
	} else {
		system.NoticeLog("File '%s' renamed successfully to '%s'", fileName, newFileName)
	}
}

// deleteDefFile will delete a definition file (Note or Solution)
func deleteDefFile(fileName string) {
	if err := os.Remove(fileName); err != nil {
		system.ErrorExit("Failed to remove file '%s' - %v", fileName, err)
	} else {
		system.NoticeLog("File '%s' removed successfully", fileName)
	}
}

// switchOffColor turns off color and highlighting, if Stdout is not a terminal
func switchOffColor() {
	// switch off color and highlighting, if Stdout is not a terminal
	// command line option --force-color will override the 'switch off'
	if !system.OutIsTerm(os.Stdout) && !system.IsFlagSet("force-color") {
		setGreenText = ""
		setRedText = ""
		setYellowText = ""
		setBlueText = ""
		setBoldText = ""
		resetBoldText = ""
		setStrikeText = ""
		resetTextColor = ""
	}
}

func cmdLineSyntax() string {
	return `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune [--format FORMAT] [--force-color] [--fun] daemon ( start | stop | status [--non-compliance-check] ) ATTENTION: deprecated
  saptune [--format FORMAT] [--force-color] [--fun] service ( start | stop | restart | takeover | enable | disable | enablestart | disablestop | status [--non-compliance-check] )
Tune system according to SAP and SUSE notes:
  saptune [--format FORMAT] [--force-color] [--fun] note ( list | verify | refresh | revertall | enabled | applied )
  saptune [--format FORMAT] [--force-color] [--fun] note ( apply | simulate | customise | create | edit | revert | show | delete ) NOTEID
  saptune [--format FORMAT] [--force-color] [--fun] note refresh [NOTEID|applied]
  saptune [--format FORMAT] [--force-color] [--fun] note verify [--colorscheme SCHEME] [--show-non-compliant] [NOTEID|applied]
  saptune [--format FORMAT] [--force-color] [--fun] note rename NOTEID NEWNOTEID
Tune system for all notes applicable to your SAP solution:
  saptune [--format FORMAT] [--force-color] [--fun] solution ( list | verify | enabled | applied )
  saptune [--format FORMAT] [--force-color] [--fun] solution ( apply | simulate | customise | create | edit | revert | show | delete ) SOLUTIONNAME
  saptune [--format FORMAT] [--force-color] [--fun] solution change [--force] SOLUTIONNAME
  saptune [--format FORMAT] [--force-color] [--fun] solution verify [--colorscheme SCHEME] [--show-non-compliant] [SOLUTIONNAME]
  saptune [--format FORMAT] [--force-color] [--fun] solution rename SOLUTIONNAME NEWSOLUTIONNAME
Staging control:
   saptune [--format FORMAT] [--force-color] [--fun] staging ( status | enable | disable | is-enabled | list )
   saptune [--format FORMAT] [--force-color] [--fun] staging ( analysis | diff ) [ ( NOTEID | SOLUTIONNAME.sol )... | all ]
   saptune [--format FORMAT] [--force-color] [--fun] staging release [--force|--dry-run] [ ( NOTEID | SOLUTIONNAME.sol )... | all ]
Config (re-)settings:
  saptune [--format FORMAT] [--force-color] [--fun] configure ( COLOR_SCHEME | SKIP_SYSCTL_FILES | IGNORE_RELOAD | DEBUG | TrentoASDP ) Value
  saptune [--format FORMAT] [--force-color] [--fun] configure ( reset | show )
Verify all applied Notes:
  saptune [--format FORMAT] [--force-color] [--fun] verify applied
Refresh all applied Notes:
  saptune [--format FORMAT] [--force-color] [--fun] refresh applied
Revert all parameters tuned by the SAP notes or solutions:
  saptune [--format FORMAT] [--force-color] [--fun] revert all
Remove the pending lock file from a former saptune call
  saptune [--format FORMAT] [--force-color] [--fun] lock remove
Call external script '/usr/sbin/saptune_check'
  saptune [--format FORMAT] [--force-color] [--fun] check
Print current saptune status:
  saptune [--format FORMAT] [--force-color] [--fun] status [--non-compliance-check]
Print current saptune version:
  saptune [--format FORMAT] [--force-color] [--fun] version
Print this message:
  saptune [--format FORMAT] [--force-color] [--fun] help

Deprecation list:
  all 'saptune daemon' actions
  'saptune note simulate'
  'saptune solution simulate'
  'Solution SAP-ASE.sol and related Note 1680803'
`
}

// PrintHelpAndExit prints the usage and exit
func PrintHelpAndExit(writer io.Writer, exitStatus int) {
	if system.GetFlagVal("format") == "json" {
		system.JInvalid(exitStatus)
	}
	fmt.Fprintf(writer, cmdLineSyntax())
	system.ErrorExit("", exitStatus)
}
