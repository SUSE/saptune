package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"strings"
)

// section [service]

// GetServiceVal initialise the systemd service structure with the current
// system settings
func GetServiceVal(key string) string {
	var val string
	serviceKey := key
	keyFields := strings.Split(key, ":")
	if len(keyFields) == 2 {
		// keyFields[0] = systemd - for further use
		serviceKey = keyFields[1]
	}
	service := system.GetServiceName(serviceKey)
	if service == "" {
		return "NA"
	}
	if system.SystemctlIsRunning(service) {
		val = "start"
	} else {
		val = "stop"
	}
	if system.SystemctlIsEnabled(service) {
		val = fmt.Sprintf("%s, enable", val)
	} else {
		val = fmt.Sprintf("%s, disable", val)
	}
	return val
}

// OptServiceVal optimises the systemd service structure with the settings
// from the configuration file
func OptServiceVal(key, cfgval string) string {
	ssState := false
	edState := false
	retVal := ""
	serviceKey := key
	keyFields := strings.Split(key, ":")
	if len(keyFields) == 2 {
		// keyFields[0] = systemd - for further use
		serviceKey = keyFields[1]
	}
	service := system.GetServiceName(serviceKey)
	if service == "" {
		return "NA"
	}

	for _, state := range strings.Split(cfgval, ",") {
		sval := strings.ToLower(strings.TrimSpace(state))
		if sval != "" && sval != "start" && sval != "stop" && sval != "enable" && sval != "disable" {
			system.WarningLog("wrong service state '%s' for '%s'. Skipping...\n", sval, service)
		}
		setVal := ""
		if sval == "start" || sval == "stop" {
			if ssState {
				system.WarningLog("multiple start/stop entries found, using the first one and skipping '%s'\n", sval)
			} else {
				// only the first 'start/stop' value is used
				ssState = true
				setVal = sval
			}
			// for uuidd.socket we only support 'start' (bsc#1100107)
			if service == "uuidd.socket" && setVal != "start" {
				system.WarningLog("wrong selection '%s' for '%s'. Now set to 'start' to start the service\n", sval, service)
				setVal = "start"
			}
		}
		if sval == "enable" || sval == "disable" {
			if edState {
				system.WarningLog("multiple enable/disable entries found, using the first one and skipping '%s'\n", sval)
			} else {
				// only the first 'enable/disable' value is used
				edState = true
				setVal = sval
			}
		}
		if setVal == "" {
			continue
		}
		if retVal == "" {
			retVal = setVal
		} else {
			retVal = fmt.Sprintf("%s, %s", retVal, setVal)
		}
	}
	if service == "uuidd.socket" {
		if retVal == "" {
			system.WarningLog("Set missing selection 'start' for '%s' to start the service\n", service)
			retVal = "start"
		} else if !ssState {
			system.WarningLog("Set missing selection 'start' for '%s' to start the service\n", service)
			retVal = fmt.Sprintf("%s, start", retVal)
		}
	}
	return retVal
}

// SetServiceVal applies the settings to the system
func SetServiceVal(key, value string) error {
	var err error
	// for compatibility to saptune v2 (revert!)
	// v2 - servicename, v3 - systemd:servicename
	serviceKey := key
	keyFields := strings.Split(key, ":")
	if len(keyFields) == 2 {
		// keyFields[0] = systemd - for further use
		serviceKey = keyFields[1]
	}
	service := system.GetServiceName(serviceKey)
	if service == "" {
		return nil
	}
	for _, state := range strings.Split(value, ",") {
		sval := strings.ToLower(strings.TrimSpace(state))

		if sval == "start" && !system.SystemctlIsRunning(service) {
			err = system.SystemctlStart(service)
		}
		if sval == "stop" && system.SystemctlIsRunning(service) {
			err = system.SystemctlStop(service)
		}
		if sval == "enable" && !system.SystemctlIsEnabled(service) {
			err = system.SystemctlEnable(service)
		}
		if sval == "disable" && system.SystemctlIsEnabled(service) {
			err = system.SystemctlDisable(service)
		}
	}
	return err
}
