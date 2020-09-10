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
)

// constant definitions
const (
	TunedService    = "tuned.service"
	saptuneV1       = "/usr/sbin/saptune_v1"
	logFile         = "/var/log/saptune/saptune.log"
	exitNotYetTuned = 5
)

var tuneApp *app.App                             // application configuration and tuning states
var tuningOptions note.TuningOptions             // Collection of tuning options from SAP notes and 3rd party vendors.
var debugSwitch = os.Getenv("SAPTUNE_DEBUG")     // Switch Debug on ("1") or off ("0" - default)
var verboseSwitch = os.Getenv("SAPTUNE_VERBOSE") // Switch verbose mode on ("on" - default) or off ("off")

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
	// check, if DEBUG is set in /etc/sysconfig/saptune
	if debugSwitch == "" {
		debugSwitch = sconf.GetString("DEBUG", "0")
	}
	if verboseSwitch == "" {
		verboseSwitch = sconf.GetString("VERBOSE", "on")
	}

	if arg1 := system.CliArg(1); arg1 == "" || arg1 == "help" || arg1 == "--help" {
		actions.PrintHelpAndExit(0)
	}
	if arg1 := system.CliArg(1); arg1 == "version" || arg1 == "--version" {
		fmt.Printf("current active saptune version is '%s'\n", SaptuneVersion)
		system.ErrorExit("", 0)
	}

	// All other actions require super user privilege
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "Please run saptune with root privilege.\n")
		system.ErrorExit("", 1)
	}

	// cleanup runtime file
	note.CleanUpRun()

	// activate logging
	system.LogInit(logFile, debugSwitch, verboseSwitch)

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
		system.ErrorExit("There are 'old' solutions or notes defined in file '/etc/sysconfig/saptune'. Seems there were some steps missed during the migration from saptune version 1 to version 2. Please check. Refer to saptune-migrate(7) for more information")
	}
}

// checkForTuned checks for enabled and/or running tuned and prints out
// a warning message
func checkForTuned() {
	if system.SystemctlIsEnabled(TunedService) || system.SystemctlIsRunning(TunedService) {
		system.WarningLog("ATTENTION: tuned service is enabled/active, so we may encounter conflicting tuning values")
	}
}
