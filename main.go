package main

import (
	"fmt"
	"github.com/SUSE/saptune/actions"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"os"
	"os/exec"
	"strings"
)

// constant definitions
const (
	saptuneV1 = "/usr/sbin/saptune_v1"
	saptcheck = "/usr/sbin/saptune_check"
	logFile   = "/var/log/saptune/saptune.log"
)

var tuneApp *app.App                 // application configuration and tuning states
var tuningOptions note.TuningOptions // Collection of tuning options from SAP notes and 3rd party vendors.
// Switch to control log reaction
var logSwitch = map[string]string{"verbose": os.Getenv("SAPTUNE_VERBOSE"), "debug": os.Getenv("SAPTUNE_DEBUG"), "error": os.Getenv("SAPTUNE_ERROR")}

// SaptuneVersion is the saptune version from /etc/sysconfig/saptune
var SaptuneVersion = ""

func main() {
	system.InitOut(logSwitch)
	if !system.ChkCliSyntax() {
		actions.PrintHelpAndExit(os.Stdout, 1)
	}

	// get saptune version and log switches from saptune sysconfig file
	SaptuneVersion = checkSaptuneConfigFile(os.Stderr, app.SysconfigSaptuneFile, logSwitch)

	arg1 := system.CliArg(1)
	if arg1 == "version" || system.IsFlagSet("version") {
		fmt.Printf("current active saptune version is '%s'\n", SaptuneVersion)
		system.Jcollect(SaptuneVersion)
		system.ErrorExit("", 0)
	}
	if arg1 == "help" || system.IsFlagSet("help") {
		system.JnotSupportedYet()
		actions.PrintHelpAndExit(os.Stdout, 0)
	}
	if arg1 == "" {
		actions.PrintHelpAndExit(os.Stdout, 1)
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
	system.InfoLog("saptune (%s) started with '%s'", actions.RPMVersion, strings.Join(os.Args, " "))

	if arg1 == "lock" {
		if arg2 := system.CliArg(2); arg2 == "remove" {
			system.JnotSupportedYet()
			system.ReleaseSaptuneLock()
			system.InfoLog("command line triggered remove of lock file '/run/.saptune.lock'\n")
			system.ErrorExit("", 0)
		} else {
			actions.PrintHelpAndExit(os.Stdout, 1)
		}
	}
	callSaptuneCheckScript(arg1)

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
		system.ErrorExit("Wrong saptune version in file '/etc/sysconfig/saptune': %s", SaptuneVersion, 128)
	}

	solutionSelector := system.GetSolutionSelector()
	archSolutions, exist := solution.AllSolutions[solutionSelector]
	system.AddGap(os.Stdout)
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
	actions.CheckOrphanedOverrides()
	actions.SelectAction(os.Stdout, tuneApp, SaptuneVersion)
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
	active, _ := system.SystemctlIsRunning(actions.TunedService)
	enabled, _ := system.SystemctlIsEnabled(actions.TunedService)
	if enabled || active {
		system.WarningLog("ATTENTION: tuned service is active, so we may encounter conflicting tuning values")
	}
}

// callSaptuneCheckScript will simply call the saptune_check script
// it's done before the saptune lock is set, but after the check for
// running as root
func callSaptuneCheckScript(arg string) {
	if arg == "check" {
		system.JnotSupportedYet()
		// call external scrip saptune_check
		cmd := exec.Command(saptcheck)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			system.ErrorExit("command '%+s' failed with error '%v'\n", saptcheck, err)
		} else {
			system.ErrorExit("", 0)
		}
	}
}

// checkWorkingArea checks, if solution and note configs exist in the working
// area
// if not, copy the definition files from the package area into the working area
// Should be covered by package installation but better safe than sorry
func checkWorkingArea() {
	refresh := false
	files := map[string]string{"note": actions.NoteTuningSheets, "solution": actions.SolutionSheets}
	for obj, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			// missing working area /var/lib/saptune/working/{notes,sols}
			refresh = true
			fmt.Println()
			system.WarningLog("missing the %ss in the working area, so copy %s definitions from package area to working area", obj, obj)
			if err := os.MkdirAll(file, 0755); err != nil {
				system.ErrorExit("Problems creating directory '%s' - '%v'", file, err)
			}
			if obj == "solution" {
				obj = "sol"
			}
			// package area /usr/share/saptune/{notes,sols}
			packedObjs := fmt.Sprintf("%s%ss/", actions.PackageArea, obj)
			_, files := system.ListDir(packedObjs, "")
			for _, f := range files {
				src := fmt.Sprintf("%s%s", packedObjs, f)
				dest := fmt.Sprintf("%s%s", file, f)
				if err := system.CopyFile(src, dest); err != nil {
					system.ErrorLog("Problems copying '%s' to '%s', continue with next file ...", src, dest)
				}
			}
		}
	}
	if refresh {
		// refresh
		solution.Refresh()
	}
}

// checkSaptuneConfigFile checks the config file /etc/sysconfig/saptune
// if it exists, if it contains all needed variables and for some variables
// checks, if the values is valid
// returns the saptune version and changes some log switches
func checkSaptuneConfigFile(writer io.Writer, saptuneConf string, lswitch map[string]string) string {
	missingKey := []string{}
	keyList := []string{app.TuneForSolutionsKey, app.TuneForNotesKey, app.NoteApplyOrderKey, "SAPTUNE_VERSION", "STAGING", "COLOR_SCHEME", "SKIP_SYSCTL_FILES", "IGNORE_RELOAD"}
	sconf, err := txtparser.ParseSysconfigFile(saptuneConf, false)
	if err != nil {
		fmt.Fprintf(writer, "Error: Unable to read file '%s': %v\n", saptuneConf, err)
		system.ErrorExit("", 128)
	}
	// check, if all needed variables are available in the saptune
	// config file
	for _, key := range keyList {
		if !sconf.IsKeyAvail(key) {
			missingKey = append(missingKey, key)
		}
	}
	if len(missingKey) != 0 {
		fmt.Fprintf(writer, "Error: File '%s' is broken. Missing variables '%s'\n", saptuneConf, strings.Join(missingKey, ", "))
		system.ErrorExit("", 128)
	}
	txtparser.GetSysctlExcludes(sconf.GetString("SKIP_SYSCTL_FILES", ""))
	stageVal := sconf.GetString("STAGING", "")
	if stageVal != "true" && stageVal != "false" {
		fmt.Fprintf(writer, "Error: Variable 'STAGING' from file '%s' contains a wrong value '%s'. Needs to be 'true' or 'false'\n", saptuneConf, stageVal)
		system.ErrorExit("", 128)
	}

	// set values read from the config file
	saptuneVers := sconf.GetString("SAPTUNE_VERSION", "")
	if saptuneVers != "1" && saptuneVers != "2" && saptuneVers != "3" {
		system.ErrorExit("Wrong saptune version in file '/etc/sysconfig/saptune': %s", SaptuneVersion, 128)
	}

	// Switch Debug on ("on") or off ("off" - default)
	// Switch verbose mode on ("on" - default) or off ("off")
	// Switch error mode on ("on" - default) or off ("off")
	// check, if DEBUG, ERROR or VERBOSE is set in /etc/sysconfig/saptune
	if lswitch["debug"] == "" {
		lswitch["debug"] = sconf.GetString("DEBUG", "off")
	}
	if lswitch["verbose"] == "" {
		lswitch["verbose"] = sconf.GetString("VERBOSE", "on")
	}
	if lswitch["error"] == "" {
		lswitch["error"] = sconf.GetString("ERROR", "on")
	}
	return saptuneVers
}
