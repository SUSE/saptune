package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"strings"
)

// section [cpu]

// GetCPUVal initialise the cpu performance structure with the current
// system settings
func GetCPUVal(key string) (string, string, string) {
	var val string
	cpuStateDiffer := false
	flsVal := ""
	info := ""
	switch key {
	case "force_latency":
		val, flsVal, cpuStateDiffer = system.GetFLInfo()
		if cpuStateDiffer {
			info = "hasDiffs"
		}
	case "energy_perf_bias":
		// cpupower -c all info  -b
		val = system.GetPerfBias()
	case "governor":
		// cpupower -c all frequency-info -p
		//or better
		// cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor
		newGov := system.GetGovernor()
		for k, v := range newGov {
			val = val + fmt.Sprintf("%s:%s ", k, v)
		}
	}
	val = strings.TrimSpace(val)
	if val == "all:none" {
		info = "notSupported"
	}
	return val, flsVal, info
}

// OptCPUVal optimises the cpu performance structure with the settings
// from the configuration file
func OptCPUVal(key, actval, cfgval string) string {
	//ANGI TODO - check cfgval is not a single value like 'performance' but
	// cpu0:performance cpu2:powersave
	sval := strings.ToLower(cfgval)
	rval := ""
	val := "0"
	switch key {
	case "force_latency":
		rval = sval
	case "energy_perf_bias":
		//performance - 0, normal - 6, powersave - 15
		switch sval {
		case "performance":
			val = "0"
		case "normal":
			val = "6"
		case "powersave":
			val = "15"
		default:
			system.WarningLog("wrong selection for energy_perf_bias. Now set to 'performance'")
			val = "0"
		}
		//ANGI TODO - if actval 'all:6', but cfgval 'cpu0:performance cpu1:normal cpu2:powersave'
		// - does not work now - check length of both?
		// same for governor
		for _, entry := range strings.Fields(actval) {
			fields := strings.Split(entry, ":")
			rval = rval + fmt.Sprintf("%s:%s ", fields[0], val)
		}
	case "governor":
		val = sval
		for _, entry := range strings.Fields(actval) {
			fields := strings.Split(entry, ":")
			rval = rval + fmt.Sprintf("%s:%s ", fields[0], val)
		}
	}
	return strings.TrimSpace(rval)
}

// SetCPUVal applies the settings to the system
func SetCPUVal(key, value, noteID, savedStates, oval string, revert bool) error {
	var err error
	switch key {
	case "force_latency":
		if oval != "untouched" {
			err = system.SetForceLatency(value, savedStates, revert)
			if !revert {
				// the cpu state values of the note need to be stored
				// after they are set. Special for 'force_latency'
				// as we set and handle 2 different sort of values
				// the 'force_latency' value and the related
				// cpu state values
				_, flstates, _ = system.GetFLInfo()
				AddParameterNoteValues("fl_states", flstates, noteID, "add")
			}
		}
	case "energy_perf_bias":
		err = system.SetPerfBias(value)
	case "governor":
		err = system.SetGovernor(value)
	}

	return err
}
