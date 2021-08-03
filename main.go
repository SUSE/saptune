package main

import (
	"fmt"
	"github.com/SUSE/saptune/actions"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"os/exec"
	"strings"
)

// constant definitions
const (
	TunedService    = "tuned.service"
	saptuneV1       = "/usr/sbin/saptune_v1"
	logFile         = "/var/log/saptune/saptune.log"
	exitNotYetTuned = 5
)

var tuneApp *app.App                 // application configuration and tuning states
var tuningOptions note.TuningOptions // Collection of tuning options from SAP notes and 3rd party vendors.
// Switch to control log reaction
var logSwitch = map[string]string{"verbose": os.Getenv("SAPTUNE_VERBOSE"), "debug": os.Getenv("SAPTUNE_DEBUG")}

// SaptuneVersion is the saptune version from /etc/sysconfig/saptune
var SaptuneVersion = ""

func main() {
	// get saptune version
	sconf, err := txtparser.ParseSysconfigFile("/etc/sysconfig/saptune", true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to read file '/etc/sysconfig/saptune': %v\n", err)
		system.ErrorExit("", 1)
	}
	SaptuneVersion = sconf.GetString("SAPTUNE_VERSION", "")
	// Switch Debug on ("1") or off ("0" - default)
	// Switch verbose mode on ("on" - default) or off ("off")
	// check, if DEBUG or VERBOSE is set in /etc/sysconfig/saptune
	if logSwitch["debug"] == "" {
		logSwitch["debug"] = sconf.GetString("DEBUG", "0")
	}
	if logSwitch["verbose"] == "" {
		logSwitch["verbose"] = sconf.GetString("VERBOSE", "on")
	}

	arg1 := system.CliArg(1)
	if arg1 == "version" || system.IsFlagSet("version") {
		fmt.Printf("current active saptune version is '%s'\n", SaptuneVersion)
		system.ErrorExit("", 0)
	}
	if arg1 == "" || arg1 == "help" || system.IsFlagSet("help") {
		actions.PrintHelpAndExit(os.Stdout, 0)
	}

	// All other actions require super user privilege
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "Please run saptune with root privilege.\n")
		system.ErrorExit("", 1)
	}

	// activate logging
	system.LogInit(logFile, logSwitch)
	// now system.ErrorExit can write to log and os.Stderr. No longer extra
	// care is needed.
	system.InfoLog("saptune started with '%s'", strings.Join(os.Args, " "))

	if arg1 == "lock" {
		if arg2 := system.CliArg(2); arg2 == "remove" {
			system.ReleaseSaptuneLock()
			system.InfoLog("command line triggered remove of lock file '/run/.saptune.lock'\n")
			system.ErrorExit("", 0)
		} else {
			actions.PrintHelpAndExit(os.Stdout, 0)
		}
	}

	// only one instance of saptune should run
	// check and set saptune lock file
	system.SaptuneLock()
	defer system.ReleaseSaptuneLock()

	// cleanup runtime files
	system.CleanUpRun()
	// additional clear ignore flag for the sapconf/saptune service deadlock
	os.Remove("/run/.saptune.ignore")

	//check, running config exists
	checkWorkingArea()

	switch SaptuneVersion {
	case "1":
		cmd := exec.Command(saptuneV1, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			system.ErrorExit("command '%+s %+v' failed with error '%v'\n", saptuneV1, os.Args, err)
		} else {
			system.ErrorExit("", 0)
		}
	case "2", "3":
		break
	default:
		system.ErrorExit("Wrong saptune version in file '/etc/sysconfig/saptune': %s", SaptuneVersion)
	}

	solutionSelector := system.GetSolutionSelector()
	archSolutions, exist := solution.AllSolutions[solutionSelector]
	if !exist {
		system.ErrorExit("The system architecture (%s) is not supported.", solutionSelector)
		return
	}
	// Initialise application configuration and tuning procedures
	tuningOptions = note.GetTuningOptions(actions.NoteTuningSheets, actions.ExtraTuningSheets)
	tuneApp = app.InitialiseApp("", "", tuningOptions, archSolutions)

	checkUpdateLeftOvers()
	if err := tuneApp.NoteSanityCheck(); err != nil {
		system.ErrorExit("Error during NoteSanityCheck - '%v'\n", err)
	}
	checkForTuned()
	actions.SelectAction(tuneApp, SaptuneVersion)
	system.ErrorExit("", 0)
}

// checkUpdateLeftOvers checks for left over files from the migration of
// saptune version 1 to saptune version 2
func checkUpdateLeftOvers() {
	// check for the /etc/tuned/saptune/tuned.conf file created during
	// the package update from saptune v1 to saptune v2/3
	// give a Warning but go ahead tuning the system
	if system.CheckForPattern("/etc/tuned/saptune/tuned.conf", "#stv1tov2#") {
		system.WarningLog("found file '/etc/tuned/saptune/tuned.conf' left over from the migration of saptune version 1 to saptune version 3. Please check and remove this file as it may work against the settings of some SAP Notes. For more information refer to the man page saptune-migrate(7)")
	}

	// check if old solution or notes are applied
	if tuneApp != nil && (len(tuneApp.NoteApplyOrder) == 0 && (len(tuneApp.TuneForNotes) != 0 || len(tuneApp.TuneForSolutions) != 0)) {
		system.ErrorExit("There are 'old' solutions or notes defined in file '/etc/sysconfig/saptune'. Seems there were some steps missed during the migration from saptune version 1 to version 3. Please check. Refer to saptune-migrate(7) for more information")
	}
}

// checkForTuned checks for enabled and/or running tuned and prints out
// a warning message
func checkForTuned() {
	if system.SystemctlIsEnabled(TunedService) || system.SystemctlIsRunning(TunedService) {
		system.WarningLog("ATTENTION: tuned service is active, so we may encounter conflicting tuning values")
	}
}

// checkWorkingArea checks, if solution and note configs exist in the working
// area
// if not, copy the definition files from the package area into the working area
// Should be covered by package installation but better safe than sorry
func checkWorkingArea() {
	if _, err := os.Stat(actions.NoteTuningSheets); os.IsNotExist(err) {
		// missing working area /var/lib/saptune/working/notes/
		system.WarningLog("missing the notes in the working area, so copy note definitions from package area to working area")
		if err := os.MkdirAll(actions.NoteTuningSheets, 0755); err != nil {
			system.ErrorExit("Problems creating directory '%s' - '%v'", actions.NoteTuningSheets, err)
			return
		}
		packedNotes := fmt.Sprintf("%snotes/", actions.PackageArea)
		_, files := system.ListDir(packedNotes, "")
		for _, f := range files {
			src := fmt.Sprintf("%s%s", packedNotes, f)
			dest := fmt.Sprintf("%s%s", actions.NoteTuningSheets, f)
			if err := system.CopyFile(src, dest); err != nil {
				system.ErrorLog("Problems copying '%s' to '%s', continue with next file ...", src, dest)
			}
		}
	}
	if _, err := os.Stat(actions.SolutionSheets); os.IsNotExist(err) {
		// missing working area /var/lib/saptune/working/sols/
		system.WarningLog("missing the solutions in the working area, so copy solution definitions from package area to working area")
		if err := os.MkdirAll(actions.SolutionSheets, 0755); err != nil {
			system.ErrorExit("Problems creating directory '%s' - '%v'", actions.SolutionSheets, err)
			return
		}
		packedSols := fmt.Sprintf("%ssols/", actions.PackageArea)
		_, files := system.ListDir(packedSols, "")
		for _, f := range files {
			src := fmt.Sprintf("%s%s", packedSols, f)
			dest := fmt.Sprintf("%s%s", actions.SolutionSheets, f)
			if err := system.CopyFile(src, dest); err != nil {
				system.ErrorLog("Problems copying '%s' to '%s', continue with next file ...", src, dest)
			}
		}
	}
}
