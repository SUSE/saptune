package actions

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
)

// ServiceAction handles service actions like start, stop, status, enable, disable
// it controlles the systemd saptune.service
//func ServiceAction(actionName string, tuneApp *app.App) {
func ServiceAction(actionName, saptuneVersion string, tApp *app.App) {
	switch actionName {
	case "start":
		ServiceActionStart(false, tApp)
	case "apply":
		// This action name is only used by saptune service, hence it is not advertised to end user.
		ServiceActionApply(tApp)
	case "enable":
		ServiceActionEnable()
	case "enablestart":
		ServiceActionStart(true, tApp)
	case "status":
		ServiceActionStatus(os.Stdout, tApp, saptuneVersion)
	case "stop":
		ServiceActionStop(false)
	case "revert":
		// This action name is only used by saptune service, hence it is not advertised to end user.
		ServiceActionRevert(tApp)
	case "stopdisable":
		ServiceActionStop(true)
	case "disable":
		ServiceActionDisable()
	default:
		PrintHelpAndExit(1)
	}
}

// ServiceActionStart starts the saptune service
// enable service before start, if enableService is true
func ServiceActionStart(enableService bool, tuneApp *app.App) {
	var err error
	system.InfoLog("Starting 'saptune.service', this may take some time...")
	// disable and stop 'sapconf.service'
	if system.IsServiceAvailable(SapconfService) {
		if err = system.SystemctlDisableStop(SapconfService); err != nil {
			system.ErrorExit("%v", err)
		}
	}
	// enable and/or start 'saptune.service'
	if enableService {
		err = system.SystemctlEnableStart(SaptuneService)
		system.InfoLog("Service 'saptune.service' has been enabled and started.")
	} else {
		err = system.SystemctlStart(SaptuneService)
		system.InfoLog("Service 'saptune.service' has been started.")
	}
	if err != nil {
		system.ErrorExit("%v", err)
	}
	// saptune.service then calls `saptune service apply` to
	// tune the system
	if len(tuneApp.TuneForSolutions) == 0 && len(tuneApp.TuneForNotes) == 0 {
		system.InfoLog("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
	}
	if !system.SystemctlIsEnabled(SaptuneService) {
		system.InfoLog("\nRemember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable saptune.service by running:\n    saptune service enable\n")
	}
}

// ServiceActionApply is only used by saptune service, hence it is not
// advertised to the end user. It is used to tune the system after reboot
func ServiceActionApply(tuneApp *app.App) {
	// service should fail, if sapconf.service is enabled or has exited
	// but 'active' file is available
	// /var/lib/sapconf/act_profile in sle12
	// /var/run/sapconf/active in sle15
	if system.SystemctlIsEnabled(SapconfService) || system.CmdIsAvailable("/var/lib/sapconf/act_profile") || system.CmdIsAvailable("/var/run/sapconf/active") {
		system.ErrorExit("ATTENTION: found an active sapconf, so refuse any action")
	}
	system.InfoLog("saptune is now tuning the system...")
	if err := tuneApp.TuneAll(); err != nil {
		system.ErrorExit("%v", err)
	}
}

// ServiceActionEnable enables the saptune service
func ServiceActionEnable() {
	system.InfoLog("Enable 'saptune.service'")
	// disable and stop 'sapconf.service'
	if system.IsServiceAvailable(SapconfService) {
		if err := system.SystemctlDisableStop(SapconfService); err != nil {
			system.ErrorExit("%v", err)
		}
	}
	// enable 'saptune.service'
	if err := system.SystemctlEnable(SaptuneService); err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog("Service 'saptune.service' has been enabled.")
	if !system.SystemctlIsRunning(SaptuneService) {
		system.InfoLog("Service 'saptune.service' is not running. Please use `saptune service start` to start the service and tune the system")
	}
}

// ServiceActionStatus checks the status of the saptune service
func ServiceActionStatus(writer io.Writer, tuneApp *app.App, saptuneVersion string) {
	// check for running saptune.service
	if system.SystemctlIsRunning(SaptuneService) {
		system.InfoLog("Service 'saptune.service' is running.")
	} else {
		system.ErrorLog("Service 'saptune.service' is stopped. If you wish to start the service, run `saptune service start`.")
		system.ErrorExit("", exitSaptuneStopped)
	}
	if !system.SystemctlIsEnabled(SaptuneService) {
		system.InfoLog("Service 'saptune.service' is disabled.")
		system.InfoLog("Remember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable saptune.service by running:\n    saptune service enable")
	} else {
		system.InfoLog("Service 'saptune.service' is enabled.")
	}

	// print saptune version
	system.InfoLog("current active saptune version is '%s'\n", saptuneVersion)
	// print saptune rpm version and date
	system.InfoLog("installed saptune version is '%s' from '%s'\n", RPMVersion, RPMDate)

	// Check for any enabled note/solution
	if len(tuneApp.TuneForSolutions) > 0 || len(tuneApp.TuneForNotes) > 0 {
		fmt.Fprintf(writer, "The system has been tuned for the following solutions and notes:")
		for _, sol := range tuneApp.TuneForSolutions {
			fmt.Fprintf(writer, "\t"+sol)
		}
		for _, noteID := range tuneApp.TuneForNotes {
			fmt.Fprintf(writer, "\t"+noteID)
		}
		// list order of enabled notes
		tuneApp.PrintNoteApplyOrder(writer)
	} else {
		system.ErrorLog("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
		system.ErrorExit("", exitNotTuned)
	}
}

// ServiceActionStop stops the saptune service
// disable service before stop, if disableService is true
func ServiceActionStop(disableService bool) {
	var err error
	system.InfoLog("Stopping 'saptune.service', this may take some time...")
	// disable and/or stop 'saptune.service'
	if disableService {
		err = system.SystemctlDisableStop(SaptuneService)
		system.InfoLog("Service 'saptune.service' has been disabled and stopped.")
	} else {
		err = system.SystemctlStop(SaptuneService)
		system.InfoLog("Service 'saptune.service' has been stopped.")
	}
	if err != nil {
		system.ErrorExit("%v", err)
	}
	// saptune.service then calls `saptune daemon revert` to
	// revert all tuned parameter
	system.InfoLog("All tuned parameters have been reverted to default.")
}

// ServiceActionRevert is only used by saptune service, hence it is not
// advertised to the end user. It is used to revert all the tuned parameters
// right before a system reboot
func ServiceActionRevert(tuneApp *app.App) {
	// service should fail, if sapconf.service is enabled or has exited
	// but 'active' file is available
	// /var/lib/sapconf/act_profile in sle12
	// /var/run/sapconf/active in sle15
	if system.SystemctlIsEnabled(SapconfService) || system.CmdIsAvailable("/var/lib/sapconf/act_profile") || system.CmdIsAvailable("/var/run/sapconf/active") {
		system.ErrorExit("ATTENTION: found an active sapconf, so refuse any action")
	}
	system.InfoLog("saptune is now reverting all settings...")
	if err := tuneApp.RevertAll(false); err != nil {
		system.ErrorExit("%v", err)
	}
}

// ServiceActionDisable disables the saptune service
func ServiceActionDisable() {
	system.InfoLog("Disable 'saptune.service'")
	// disable 'saptune.service'
	if err := system.SystemctlDisable(SaptuneService); err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog("Service 'saptune.service' has been disabled.")
	if system.SystemctlIsRunning(SaptuneService) {
		system.InfoLog("Service 'saptune.service' still running. Please use `saptune service stop` to stop the service and revert the tuned parameter")
	}
}

// DaemonAction handles daemon actions like start, stop, status asm.
// still available for compatibility reasons
func DaemonAction(actionName, saptuneVersion string, tuneApp *app.App) {
	system.WarningLog("ATTENTION: the argument 'daemon' is deprecated!. saptune will forward the request to 'saptune service %s'.\nFor the future please use 'saptune service %s'.", actionName, actionName)
	switch actionName {
	case "start":
		ServiceActionStart(false, tuneApp)
	case "status":
		ServiceActionStatus(os.Stdout, tuneApp, saptuneVersion)
	case "stop":
		ServiceActionStop(false)
	default:
		PrintHelpAndExit(1)
	}
}
