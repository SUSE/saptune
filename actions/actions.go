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

//define constants and variables for the whole package
const (
	SaptuneService     = "saptune.service"
	SapconfService     = "sapconf.service"
	TunedService       = "tuned.service"
	exitSaptuneStopped = 1
	exitNotTuned       = 3
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

// SolutionSheets is the working directory of available sap solutions
var SolutionSheets = "/var/lib/saptune/working/sols/"

// RPMVersion is the package version from package build process
var RPMVersion = "undef"

// RPMDate is the date of package build
// only used in individual build test packages, but NOT in our official
// built and released packages (not possible because of 'reproducible' builds)
var RPMDate = "undef"

// solutionSelector used in solutionacts and statgingacts
var solutionSelector = system.GetSolutionSelector()

// set colors for the table and list output
//var setBlueText = "\033[34m"
//var setCyanText = "\033[36m"
//var setUnderlinedText = "\033[4m"
var setGreenText = "\033[32m"
var setRedText = "\033[31m"
var setBoldText = "\033[1m"
var setStrikeText = "\033[9m"
var resetTextColor = "\033[0m"

// SelectAction selects the chosen action depending on the first command line
// argument
func SelectAction(stApp *app.App, saptuneVers string) {
	// switch off color and highlighting, if Stdout is not a terminal
	if !system.OutIsTerm(os.Stdout) {
		setGreenText = ""
		setRedText = ""
		setBoldText = ""
		setStrikeText = ""
		resetTextColor = ""
	}
	// check for test packages
	if RPMDate != "undef" {
		system.NoticeLog("ATTENTION: You are running a test version of saptune which is not supported for production use")
	}

	switch system.CliArg(1) {
	case "daemon":
		DaemonAction(os.Stdout, system.CliArg(2), saptuneVers, stApp)
	case "service":
		ServiceAction(system.CliArg(2), saptuneVers, stApp)
	case "note":
		NoteAction(system.CliArg(2), system.CliArg(3), system.CliArg(4), stApp)
	case "solution":
		SolutionAction(system.CliArg(2), system.CliArg(3), system.CliArg(4), stApp)
	case "revert":
		RevertAction(os.Stdout, system.CliArg(2), stApp)
	case "staging":
		StagingAction(system.CliArg(2), system.CliArgs(3), stApp)
	case "status":
		ServiceAction("status", saptuneVers, stApp)
	default:
		PrintHelpAndExit(os.Stdout, 1)
	}
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
	if len(tuneApp.NoteApplyOrder) == 0 {
		fmt.Fprintf(writer, "No notes or solutions enabled, nothing to verify.\n")
	} else {
		unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
		if err != nil {
			system.ErrorExit("Failed to inspect the current system: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true)
		tuneApp.PrintNoteApplyOrder(writer)
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "The running system is currently well-tuned according to all of the enabled notes.\n")
		} else {
			system.ErrorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
		}
	}
}

// getFileName returns the corresponding filename of a given definition file
// (note or solution)
// additional it returns a boolean value which is pointing out that
// the definition is a custom definition (extraDef = true) or an internal one
func getFileName(defName, workingDir, extraDir string) (string, bool) {
	extraDef := false
	defType := "Note"
	if workingDir == SolutionSheets {
		defType = "Solution"
	}
	fileName := fmt.Sprintf("%s%s", workingDir, defName)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		// Note/solution is NOT an internal Note/solution,
		// but may be a custom Note/solution
		extraDef = true
		chkName := defName
		if defType == "Note" {
			chkName = defName + ".conf"
		}
		_, files := system.ListDir(extraDir, "")
		for _, f := range files {
			if f == chkName {
				fileName = fmt.Sprintf("%s%s", extraDir, f)
			}
		}
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			system.ErrorExit("%s %s not found in %s or %s.", defType, defName, workingDir, extraDir)
		} else if err != nil {
			system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
		}
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
//func deleteDefFile(fileName, ovFileName string, overrideDef, extraDef bool) {
func deleteDefFile(fileName string) {
	if err := os.Remove(fileName); err != nil {
		system.ErrorExit("Failed to remove file '%s' - %v", fileName, err)
	} else {
		system.NoticeLog("File '%s' removed successfully", fileName)
	}
}

// PrintHelpAndExit prints the usage and exit
func PrintHelpAndExit(writer io.Writer, exitStatus int) {
	fmt.Fprintln(writer, `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | revertall | enabled | applied ]
  saptune note [ apply | simulate | verify | customise | create | edit | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled | applied ]
  saptune solution [ apply | simulate | verify | customise | create | edit | revert | show | delete ] SolutionName
  saptune solution rename SolutionName newSolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff | analysis | release ]
   saptune staging [ analysis | diff ] [ NoteID... | SolutionID... | all ]
   saptune staging release [--force|--dry-run] [ NoteID... | SolutionID... | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Print current saptune status:
  saptune status
Print current saptune version:
  saptune version
Print this message:
  saptune help`)
	system.ErrorExit("", exitStatus)
}
