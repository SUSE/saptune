package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetAvailServices returns a map of the available services of the system
func GetAvailServices() map[string]string {
	allServices := make(map[string]string)
	cmdArgs := []string{"--no-pager", "list-unit-files"}
	cmdOut, err := exec.Command(systemctlCmd, cmdArgs...).CombinedOutput()
	if err != nil {
		WarningLog("There was an error running external command %s %s: %v, output: %s", systemctlCmd, cmdArgs, err, cmdOut)
		return allServices
	}
	for _, line := range strings.Split(string(cmdOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		serv := strings.TrimSpace(fields[0])
		allServices[serv] = serv
	}
	return allServices
}

// GetServiceName returns the systemd service name for supported services
func GetServiceName(service string) string {
	serviceName := ""
	if len(services) == 0 {
		services = GetAvailServices()
	}
	if _, ok := services[service]; ok {
		serviceName = service
	} else {
		serv := fmt.Sprintf("%s.service", service)
		if _, ok := services[serv]; ok {
			serviceName = serv
		}
	}
	if serviceName == "" {
		WarningLog("skipping unknown service '%s'", service)
	}
	return serviceName
}

// CmpServiceStates compares the expected service states with the current
// active service states
func CmpServiceStates(actStates, expStates string) bool {
	ret := false
	if expStates == "" {
		return true
	}
	retStart, retEnable := checkStates(actStates, expStates)
	if (retStart == "" || retStart == "true") && (retEnable == "" || retEnable == "true") {
		ret = true
	}
	if retStart == "" && retEnable == "" {
		ret = false
	}
	return ret
}

// checkStates checks, if the expcted state matches the active state
func checkStates(actStates, expStates string) (string, string) {
	start := ""
	enable := ""
	for _, state := range strings.Split(expStates, ",") {
		// expected state
		sval := strings.ToLower(strings.TrimSpace(state))
		// check for valid states. Supported for now:
		// 'start', 'stop', 'enable' and 'disable'
		if sval != "start" && sval != "stop" && sval != "enable" && sval != "disable" {
			continue
		}
		// check, if the expected state is already availabel in the active states
		start, enable = chkActStates(actStates, sval, start, enable)
	}
	return start, enable
}

// chkActStates checks, if the expected state is already availabel in the active states
func chkActStates(actStates, sval, start, enable string) (string, string) {
	match := ""
	for _, aState := range strings.Split(actStates, ",") {
		aval := strings.ToLower(strings.TrimSpace(aState))
		if sval == aval {
			match = "true"
			break
		} else {
			match = "false"
		}
	}
	// evaluate start and enable result per expected state
	start, enable = evalStartEnable(sval, match, start, enable)
	return start, enable
}

// evaluate start and enable result per expected state
func evalStartEnable(sval, match, start, enable string) (string, string) {
	if sval == "start" || sval == "stop" {
		if start != "true" {
			start = match
		}
	} else {
		if enable != "true" {
			enable = match
		}
	}
	return start, enable
}

// IsSapconfActive checks, if sapconf is active
func IsSapconfActive(sapconf string) bool {
	active, _ := SystemctlIsRunning(sapconf)
	enabled, _ := SystemctlIsEnabled(sapconf)
	actFile1 := CmdIsAvailable("/var/lib/sapconf/act_profile")
	actEmpty1 := FileIsEmpty("/var/lib/sapconf/act_profile")
	actFile2 := CmdIsAvailable("/run/sapconf_act_profile")
	actFile3 := CmdIsAvailable("/run/sapconf/active")
	DebugLog("IsSapconfActive - sapconf is active:%+v, enabled:%+v, /var/lib/sapconf/act_profile is available:%+v, /var/lib/sapconf/act_profile is empty:%+v, /run/sapconf_act_profile is available:%+v, /run/sapconf/active is available:%+v", active, enabled, actFile1, actEmpty1, actFile2, actFile3)
	if enabled || active || !actEmpty1 || actFile2 || actFile3 {
		return true
	}
	return false
}
