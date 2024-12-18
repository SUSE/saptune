package actions

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io"
	"os"
	"strings"
)

var mandatoryConfigKeys = []string{app.TuneForSolutionsKey, app.TuneForNotesKey, app.NoteApplyOrderKey, "SAPTUNE_VERSION", "STAGING", "COLOR_SCHEME", "SKIP_SYSCTL_FILES", "IGNORE_RELOAD"}
var changeableConfigKeys = []string{"COLOR_SCHEME", "SKIP_SYSCTL_FILES", "IGNORE_RELOAD", "DEBUG", "TrentoASDP"}

// MandKeyList returns a list of mandatory configuration parameter, which need
// to be available in the saptune configuration file
func MandKeyList() []string {
	return mandatoryConfigKeys
}

// ChangeKeyList returns a list of configuration parameter, which can be set
// or changed by the customer
func ChangeKeyList() []string {
	return changeableConfigKeys
}

// ConfigureAction changes entries in the main saptune configuration file
// Replaces the direct editing of the config file
//
// saptune configure STAGING -- not needed because of 'saptune staging enable'
func ConfigureAction(writer io.Writer, configEntry string, configVals []string, tuneApp *app.App) {
	if len(configVals) == 0 && !(configEntry == "reset" || configEntry == "show") {
		// missing value to be configured
		PrintHelpAndExit(writer, 1)
	}
	switch configEntry {
	case "COLOR_SCHEME":
		ConfigureActionSetColorScheme(configVals[0])
	case "SKIP_SYSCTL_FILES":
		ConfigureActionSetSkipSysctlFiles(configVals)
	case "IGNORE_RELOAD":
		ConfigureActionSetIgnoreReload(configVals[0])
	case "DEBUG":
		ConfigureActionSetDebug(configVals[0])
	case "TrentoASDP":
		ConfigureActionSetTrentoASDP(configVals[0])
	case "reset":
		ConfigureActionReset(os.Stdin, writer, tuneApp)
	case "show":
		ConfigureActionShow(writer)
	default:
		PrintHelpAndExit(writer, 1)
	}
}

// ConfigureActionSetColorScheme sets the color scheme
func ConfigureActionSetColorScheme(configVal string) {
	switch configVal {
	case "", "full-green-zebra", "cmpl-green-zebra", "full-blue-zebra", "cmpl-blue-zebra", "full-red-noncmpl", "red-noncmpl", "full-yellow-noncmpl", "yellow-noncmpl":
		writeConfigEntry("COLOR_SCHEME", configVal)
	default:
		system.ErrorExit("wrong value '%s' for config variable 'COLOR_SCHEME'. Please check.", configVal)
	}
}

// ConfigureActionSetIgnoreReload sets the variable IGNORE_RELOAD
func ConfigureActionSetIgnoreReload(configVal string) {
	switch configVal {
	case "yes", "no":
		writeConfigEntry("IGNORE_RELOAD", configVal)
	default:
		system.ErrorExit("wrong value '%s' for config variable 'IGNORE_RELOAD'. Only 'yes' or 'no' supported. Please check.", configVal)
	}
}

// ConfigureActionSetDebug sets the variable DEBUG
func ConfigureActionSetDebug(configVal string) {
	switch configVal {
	case "on", "off":
		writeConfigEntry("DEBUG", configVal)
	default:
		system.ErrorExit("wrong value '%s' for config variable 'DEBUG'. Only 'on' or 'off' supported. Please check.", configVal)
	}
}

// ConfigureActionSetTrentoASDP sets the saptune-discovery-period of the
// Trento Agent
func ConfigureActionSetTrentoASDP(configVal string) {
	switch configVal {
	case "300", "600", "900", "1800", "3600", "off":
		err := system.CheckAndSetTrento("TrentoASDP", configVal, true)
		if err != nil {
			system.ErrorExit("", 1)
		}
		writeConfigEntry("TrentoASDP", configVal)
	default:
		system.ErrorExit("wrong value '%s' for the Trento Agent saptune-discovery-period. Supported values are '300', '600', '900', '1800', '3600'. Please check.", configVal)
	}
}

// ConfigureActionSetSkipSysctlFiles sets the exclude list for the sysctl
// config warnings
func ConfigureActionSetSkipSysctlFiles(configVals []string) {
	if configVals[0] == "" {
		writeConfigEntry("SKIP_SYSCTL_FILES", configVals[0])
		system.ErrorExit("", 0)
	}
	confVals := configVals
	if len(configVals) == 1 && strings.Contains(configVals[0], ",") {
		confVals = strings.Split(configVals[0], ",")
	}
	confVal := ""
	//for _, file := range configVals {
	for _, file := range confVals {
		file = strings.TrimSuffix(file, ",")
		if system.IsValidSysctlLocations(file) {
			if confVal == "" {
				confVal = file
			} else {
				confVal = confVal + ", " + file
			}
		} else {
			system.ErrorLog("wrong value '%s' for config variable 'SKIP_SYSCTL_FILES'. sysctl command will not search in this locaction. Skipping.", file)
		}
	}
	if confVal != "" {
		writeConfigEntry("SKIP_SYSCTL_FILES", confVal)
	} else {
		system.ErrorExit("wrong value(s) '%+v' for config variable 'SKIP_SYSCTL_FILES' provided. sysctl command will not search in this locaction(s). Exiting without changing saptune configuration. Please check.", configVals)
	}
}

// writeConfigEntry writes the changed config entry setting to the saptune
// config file
func writeConfigEntry(entry, val string) {
	sconf, err := txtparser.ParseSysconfigFile(saptuneSysconfig, true)
	if err != nil {
		system.ErrorExit("Unable to read file '%s': '%v'\n", saptuneSysconfig, err, 128)
	}
	if val == "" {
		system.InfoLog("Reset '%s' to empty value", entry)
	} else {
		system.InfoLog("Set '%s' to '%s'", entry, val)
	}
	sconf.Set(entry, val)
	if err := os.WriteFile(saptuneSysconfig, []byte(sconf.ToText()), 0644); err != nil {
		system.ErrorExit("'%s' could not be set to '%s'. - '%v'\n", entry, val, err)
	}
}

// ConfigureActionShow shows the content of the saptune configuration file
func ConfigureActionShow(writer io.Writer) {
	cont, err := os.ReadFile(saptuneSysconfig)
	if err != nil {
		system.ErrorExit("Unable to read file '%s': '%v'\n", saptuneSysconfig, err, 128)
	}
	fmt.Fprintf(writer, "\nContent of saptune configuration file %s:\n\n%s\n", saptuneSysconfig, string(cont))
}

// ConfigureActionReset resets the main saptune configuration to the delivery
// state
func ConfigureActionReset(reader io.Reader, writer io.Writer, tuneApp *app.App) {
	errcnt := 0
	fmt.Fprintf(writer, "\nATTENTION: resetting the main saptune configuration.\nThis will reset the tuning of the system and remove/reset all saptune related configuration and runtime files.\n")
	txtConfirm := fmt.Sprintf("Do you really want to reset the main saptune configuration?")
	if readYesNo(txtConfirm, reader, writer) {
		system.InfoLog("ATTENTION: Resetting main saptune configuration")
		// revert all
		if err := tuneApp.RevertAll(true); err != nil {
			system.ErrorLog("Failed to revert notes: %v", err)
			errcnt = errcnt + 1
		}
		// remove saved_state files, if some left over
		os.RemoveAll(system.SaptuneSectionDir)
		os.RemoveAll(system.SaptuneParameterStateDir)
		os.RemoveAll(system.SaptuneSavedStateDir)

		// set configuration file back to default/delivery
		saptuneTemplate := system.SaptuneConfigTemplate()
		if err := system.CopyFile(saptuneTemplate, saptuneSysconfig); err != nil {
			system.ErrorLog("Failed to set saptune configuration file '%s' back to delivery state by copying the template file '%s'", saptuneSysconfig, saptuneTemplate)
			errcnt = errcnt + 1
		}
	}
	if errcnt != 0 {
		system.ErrorExit("", 1)
	}
}
