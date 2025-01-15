package system

// functions related with config files

import (
	"fmt"
	"os"
	"strings"
)

// CheckAndSetTrento checks parameter and value of Trento Agent config
// and sets a new value in the Trento Agent config if requested
func CheckAndSetTrento(entry, val string, change bool) error {
	var err error
	comment := "# modified by saptune"
	if val == "off" || val == "" {
		DebugLog("Check of Trento config file is turned off (val is '%s')", val)
		return nil
	}
	logSwitch := true
	if strings.HasPrefix(CalledFrom(), "main.go:") {
		// logging not initialised
		logSwitch = false
	}
	if err := checkForValidValue(val, logSwitch); err != nil {
		return err
	}
	param := ""
	if entry == "TrentoASDP" {
		// Trento Agent
		// # saptune-discovery-period: 900s
		param = "saptune-discovery-period:"
	}
	trentoFile, err := detectTrentoConfig(logSwitch)
	if err != nil {
		return err
	}
	value := readConfValue(trentoFile, param)
	if value != val {
		val = val + "s"
		if value == val {
			return nil
		}
		// value in Trento config file differs from the new selected one.
		if !change {
			// value = "" --> entry not yet available in Trento config
			if logSwitch {
				WarningLog("Value '%s' of entry '%s' in Trento config file '%s' differs from the value configured with saptune ('%s'). Please check.", value, param, trentoFile, val)
			} else {
				// logging not initialised
				fmt.Fprintf(os.Stderr, "WARNING: Value '%s' of entry '%s' in Trento config file '%s' differs from the value configured with saptune ('%s'). Please check.\n", value, param, trentoFile, val)
			}
		}
		if value != "" && change {
			// entry available in config file but value differs
			// change with comment
			InfoLog("Value '%s' of entry '%s' in Trento config file '%s' differs from the value configured with saptune ('%s'). Changing entry...", value, param, trentoFile, val)
			err = changeEntry(trentoFile, param, val, comment)
		}
		if value == "" && change {
			InfoLog("Entry '%s' not yet available in Trento config file '%s'. Adding entry with value '%s'....", param, trentoFile, val)
			// entry not found in config file, append with comment
			err = appendEntry(trentoFile, param+" "+val, comment)
		}
	}
	return err
}

// checkForValidValue checks the value from saptune configuration file is a valid value for Trento Agent configuration
func checkForValidValue(value string, logSwitch bool) error {
	switch value {
	case "300", "600", "900", "1800", "3600":
		// valid value
		return nil
	}
	errorTXT := "Wrong value '%s' for the Trento Agent saptune-discovery-period found in saptune configuration. Supported values are '300', '600', '900', '1800', '3600'. Please check."
	if logSwitch {
		WarningLog(errorTXT, value)
	} else {
		// logging not initialised
		fmt.Fprintf(os.Stderr, "WARNING: "+errorTXT+"\n", value)
	}
	return fmt.Errorf(errorTXT, value)
}

// detectTrentoConfig returns the name of the Trento Agent config file
func detectTrentoConfig(logSwitch bool) (string, error) {
	trentoFile := "/etc/trento/agent.yaml"
	_, err := os.Stat(trentoFile)
	if err != nil {
		if logSwitch {
			ErrorLog("Trento config file '%s' not found, exiting.", trentoFile)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Trento config file '%s' not found.\n", trentoFile)
		}
	}
	return trentoFile, err
}
