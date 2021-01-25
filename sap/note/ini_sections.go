package note

import (
	"fmt"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io/ioutil"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// section name definition
const (
	INISectionSysctl    = "sysctl"
	INISectionSys       = "sys"
	INISectionVM        = "vm"
	INISectionCPU       = "cpu"
	INISectionMEM       = "mem"
	INISectionBlock     = "block"
	INISectionService   = "service"
	INISectionLimits    = "limits"
	INISectionLogin     = "login"
	INISectionVersion   = "version"
	INISectionPagecache = "pagecache"
	INISectionRpm       = "rpm"
	INISectionGrub      = "grub"
	INISectionReminder  = "reminder"
	SysKernelTHPEnabled = "kernel/mm/transparent_hugepage/enabled"
	SysKSMRun           = "kernel/mm/ksm/run"

	// LoginConfDir is the path to systemd's logind configuration directory under /etc.
	LogindConfDir = "/etc/systemd/logind.conf.d"
	// LogindSAPConfFile is a configuration file full of SAP-specific settings for logind.
	LogindSAPConfFile = "saptune-UserTasksMax.conf"
)

// section handling
// section [sysctl]

// OptSysctlVal optimises a sysctl parameter value
// use exactly the value from the config file. No calculation any more
func OptSysctlVal(operator txtparser.Operator, key, actval, cfgval string) string {
	if actval == "" {
		// sysctl parameter not available in system
		return ""
	}
	if cfgval == "" {
		// sysctl parameter should be leave untouched
		return ""
	}
	allFieldsC := strings.Fields(actval)
	allFieldsE := strings.Fields(cfgval)
	allFieldsS := ""

	if len(allFieldsC) != len(allFieldsE) && (operator == txtparser.OperatorEqual || len(allFieldsE) > 1) {
		system.WarningLog("wrong number of fields given in the config file for parameter '%s'\n", key)
		return ""
	}

	for k, fieldC := range allFieldsC {
		fieldE := ""
		if len(allFieldsC) != len(allFieldsE) {
			fieldE = fieldC

			if (operator == txtparser.OperatorLessThan || operator == txtparser.OperatorLessThanEqual) && k == 0 {
				fieldE = allFieldsE[0]
			}
			if (operator == txtparser.OperatorMoreThan || operator == txtparser.OperatorMoreThanEqual) && k == len(allFieldsC)-1 {
				fieldE = allFieldsE[0]
			}
		} else {
			fieldE = allFieldsE[k]
		}

		// use exactly the value from the config file. No calculation any more
		/*
			optimisedValue, err := CalculateOptimumValue(operator, fieldC, fieldE)
			//optimisedValue, err := CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			if err != nil {
				return ""
			}
			allFieldsS = allFieldsS + optimisedValue + "\t"
		*/
		allFieldsS = allFieldsS + fieldE + "\t"
	}

	return strings.TrimSpace(allFieldsS)
}

// section [block]

var isSched = regexp.MustCompile(`^IO_SCHEDULER_\w+$`)
var isNrreq = regexp.MustCompile(`^NRREQ_\w+$`)
var isRahead = regexp.MustCompile(`^READ_AHEAD_KB_\w+$`)

// GetBlkVal initialise the block device structure with the current
// system settings
func GetBlkVal(key string, cur *param.BlockDeviceQueue) (string, string, error) {
	newQueue := make(map[string]string)
	newReq := make(map[string]int)
	retVal := ""
	info := ""

	switch {
	case isSched.MatchString(key):
		newIOQ, err := cur.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return "", info, err
		}
		newQueue = newIOQ.(param.BlockDeviceSchedulers).SchedulerChoice
		retVal = newQueue[strings.TrimPrefix(key, "IO_SCHEDULER_")]
		cur.BlockDeviceSchedulers = newIOQ.(param.BlockDeviceSchedulers)
	case isNrreq.MatchString(key):
		newNrR, err := cur.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return "", info, err
		}
		newReq = newNrR.(param.BlockDeviceNrRequests).NrRequests
		retVal = strconv.Itoa(newReq[strings.TrimPrefix(key, "NRREQ_")])
		cur.BlockDeviceNrRequests = newNrR.(param.BlockDeviceNrRequests)
	case isRahead.MatchString(key):
		newRah, err := cur.BlockDeviceReadAheadKB.Inspect()
		if err != nil {
			return "", info, err
		}
		newReq = newRah.(param.BlockDeviceReadAheadKB).ReadAheadKB
		retVal = strconv.Itoa(newReq[strings.TrimPrefix(key, "READ_AHEAD_KB_")])
		cur.BlockDeviceReadAheadKB = newRah.(param.BlockDeviceReadAheadKB)
	}
	return retVal, info, nil
}

// OptBlkVal optimises the block device structure with the settings
// from the configuration file
func OptBlkVal(key, cfgval string, cur *param.BlockDeviceQueue, bOK map[string][]string) (string, string) {
	info := ""
	if cfgval == "" {
		return cfgval, info
	}
	sval := cfgval
	switch {
	case isSched.MatchString(key):
		// ANGI TODO - support different scheduler per device or
		// all devices with same scheduler (oval="all none")
		oval := ""
		sfound := false
		dname := regexp.MustCompile(`^IO_SCHEDULER_(\w+)$`)
		bdev := dname.FindStringSubmatch(key)
		for _, sched := range strings.Split(cfgval, ",") {
			sval = strings.ToLower(strings.TrimSpace(sched))
			if !param.IsValidScheduler(bdev[1], sval) {
				continue
			} else {
				sfound = true
				oval = bdev[1] + " " + sval
				bOK[sval] = append(bOK[sval], bdev[1])
				break
			}
		}
		if !sfound {
			sval = cfgval
			info = "NA"
		} else {
			opt, _ := cur.BlockDeviceSchedulers.Optimise(oval)
			cur.BlockDeviceSchedulers = opt.(param.BlockDeviceSchedulers)
		}
	case isNrreq.MatchString(key):
		if sval == "0" {
			sval = "1024"
		}
		ival, _ := strconv.Atoi(sval)
		opt, _ := cur.BlockDeviceNrRequests.Optimise(ival)
		cur.BlockDeviceNrRequests = opt.(param.BlockDeviceNrRequests)
	case isRahead.MatchString(key):
		if sval == "0" {
			sval = "512"
		}
		ival, _ := strconv.Atoi(sval)
		opt, _ := cur.BlockDeviceReadAheadKB.Optimise(ival)
		cur.BlockDeviceReadAheadKB = opt.(param.BlockDeviceReadAheadKB)
	}
	return sval, info
}

// SetBlkVal applies the settings to the system
func SetBlkVal(key, value string, cur *param.BlockDeviceQueue, revert bool) error {
	var err error

	switch {
	case isSched.MatchString(key):
		if revert {
			cur.BlockDeviceSchedulers.SchedulerChoice[strings.TrimPrefix(key, "IO_SCHEDULER_")] = value
		}
		err = cur.BlockDeviceSchedulers.Apply(strings.TrimPrefix(key, "IO_SCHEDULER_"))
		if err != nil {
			return err
		}
	case isNrreq.MatchString(key):
		if revert {
			ival, _ := strconv.Atoi(value)
			cur.BlockDeviceNrRequests.NrRequests[strings.TrimPrefix(key, "NRREQ_")] = ival
		}
		err = cur.BlockDeviceNrRequests.Apply(strings.TrimPrefix(key, "NRREQ_"))
		if err != nil {
			return err
		}
	case isRahead.MatchString(key):
		if revert {
			ival, _ := strconv.Atoi(value)
			cur.BlockDeviceReadAheadKB.ReadAheadKB[strings.TrimPrefix(key, "READ_AHEAD_KB_")] = ival
		}
		err = cur.BlockDeviceReadAheadKB.Apply(strings.TrimPrefix(key, "READ_AHEAD_KB_"))
		if err != nil {
			return err
		}
	}
	return err
}

// section [limits]

// GetLimitsVal initialise the security limit structure with the current
// system settings
func GetLimitsVal(value string) (string, error) {
	// Find out current limits
	limit := value
	if limit != "" && limit != "NA" {
		lim := strings.Fields(limit)
		// dom=[0], type=[1], item=[2], value=[3]
		// no check, that the syntax/order of the entry in the config file is
		// a valid limits entry

		// /etc/security/limits.d/saptune-<domain>-<item>-<type>.conf
		dropInFile := fmt.Sprintf("/etc/security/limits.d/saptune-%s-%s-%s.conf", lim[0], lim[2], lim[1])
		secLimits, err := system.ParseSecLimitsFile(dropInFile)
		if err != nil {
			//ANGI TODO - check, if other files in /etc/security/limits.d contain a value for the touple "<domain>-<item>-<type>"
			return "", err
		}
		lim[3], _ = secLimits.Get(lim[0], lim[1], lim[2])
		if lim[3] == "" {
			lim[3] = "NA"
		}
		// current limit found
		limit = strings.Join(lim, " ")
	}
	return limit, nil
}

// OptLimitsVal optimises the security limit structure with the settings
// from the configuration file or with a calculation
func OptLimitsVal(actval, cfgval string) string {
	cfgval = strings.Join(strings.Fields(strings.TrimSpace(cfgval)), " ")
	return cfgval
}

// SetLimitsVal applies the settings to the system
func SetLimitsVal(key, noteID, value string, revert bool) error {
	var err error
	limit := value
	if limit != "" && limit != "NA" {
		lim := strings.Fields(limit)
		// dom=[0], type=[1], item=[2], value=[3]

		// /etc/security/limits.d/saptune-<domain>-<item>-<type>.conf
		dropInFile := fmt.Sprintf("/etc/security/limits.d/saptune-%s-%s-%s.conf", lim[0], lim[2], lim[1])

		if revert && IsLastNoteOfParameter(key) {
			// revert - remove limits drop-in file
			os.Remove(dropInFile)
			return nil
		}

		secLimits, err := system.ParseSecLimitsFile(dropInFile)
		if err != nil {
			return err
		}

		if lim[3] != "" && lim[3] != "NA" {
			// revert with value from another former applied note
			// or
			// apply - Prepare limits drop-in file
			secLimits.Set(lim[0], lim[1], lim[2], lim[3])

			//err = secLimits.Apply()
			err = secLimits.ApplyDropIn(lim, noteID)
		}
	}
	return err
}

// section [vm]
// Manipulate /sys/kernel/mm switches.

// GetVMVal initialise the memory management structure with the current
// system settings
func GetVMVal(key string) string {
	var val string
	switch key {
	case "THP":
		val, _ = system.GetSysChoice(SysKernelTHPEnabled)
	case "KSM":
		ksmval, _ := system.GetSysInt(SysKSMRun)
		val = strconv.Itoa(ksmval)
	}
	return val
}

// OptVMVal optimises the memory management structure with the settings
// from the configuration file
func OptVMVal(key, cfgval string) string {
	val := strings.ToLower(cfgval)
	switch key {
	case "THP":
		if val != "always" && val != "madvise" && val != "never" {
			system.WarningLog("wrong selection for THP. Now set to 'never' to disable transarent huge pages")
			val = "never"
		}
	case "KSM":
		if val != "1" && val != "0" {
			system.WarningLog("wrong selection for KSM. Now set to default value '0'")
			val = "0"
		}
	}
	return val
}

// SetVMVal applies the settings to the system
func SetVMVal(key, value string) error {
	var err error
	switch key {
	case "THP":
		err = system.SetSysString(SysKernelTHPEnabled, value)
	case "KSM":
		ksmval, _ := strconv.Atoi(value)
		err = system.SetSysInt(SysKSMRun, ksmval)
	}
	return err
}

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
		// - does not work now - check lenght of both?
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
func SetCPUVal(key, value, noteID, savedStates, oval, info string, revert bool) error {
	var err error
	switch key {
	case "force_latency":
		if oval != "untouched" {
			err = system.SetForceLatency(value, savedStates, info, revert)
			if !revert {
				// the cpu state values of the note need to be stored
				// after they are set. Special for 'force_latency'
				// as we set and handle 2 different sort of values
				// the 'force_latency' value and the related
				// cpu state values
				_, flstates, _ = system.GetFLInfo()
				AddParameterNoteValues("fl_states", flstates, noteID)
			}
		}
	case "energy_perf_bias":
		err = system.SetPerfBias(value)
	case "governor":
		err = system.SetGovernor(value, info)
	}

	return err
}

// section [mem]

// GetMemVal initialise the shared memory structure with the current
// system settings
func GetMemVal(key string) string {
	var val string
	switch key {
	case "VSZ_TMPFS_PERCENT", "ShmFileSystemSizeMB":
		// Find out size of SHM
		mount, found := system.ParseProcMounts().GetByMountPoint("/dev/shm")
		if found {
			val = strconv.FormatUint(mount.GetFileSystemSizeMB(), 10)
			if key == "VSZ_TMPFS_PERCENT" {
				// rounded value
				percent := math.Floor(float64(mount.GetFileSystemSizeMB())*100/float64(system.GetTotalMemSizeMB()) + 0.5)
				val = strconv.FormatFloat(percent, 'f', -1, 64)
			}
		} else {
			system.WarningLog("GetMemVal: failed to find /dev/shm mount point")
			val = "-1"
		}
	}
	return val
}

// OptMemVal optimises the shared memory structure with the settings
// from the configuration file or with a calculation
func OptMemVal(key, actval, cfgval, tmpfspercent string) string {
	// tmpfspercent value of VSZ_TMPFS_PERCENT from config or override file
	var size uint64
	var ret string

	switch key {
	case "VSZ_TMPFS_PERCENT":
		ret = cfgval
	case "ShmFileSystemSizeMB":
		if actval == "-1" {
			system.WarningLog("OptMemVal: /dev/shm is not a valid mount point, will not calculate its optimal size.")
			ret = "-1"
		} else if cfgval != "0" {
			ret = cfgval
		} else {
			if tmpfspercent == "0" {
				// Calculate optimal SHM size (TotalMemSizeMB*75/100) (SAP-Note 941735)
				size = uint64(system.GetTotalMemSizeMB()) * 75 / 100
			} else {
				// Calculate optimal SHM size (TotalMemSizeMB*VSZ_TMPFS_PERCENT/100)
				val, _ := strconv.ParseUint(tmpfspercent, 10, 64)
				size = uint64(system.GetTotalMemSizeMB()) * val / 100
			}
			ret = strconv.FormatUint(size, 10)
		}
	}
	return ret
}

// SetMemVal applies the settings to the system
func SetMemVal(key, value string) error {
	var err error
	switch key {
	case "ShmFileSystemSizeMB":
		val, err := strconv.ParseUint(value, 10, 64)
		if val > 0 {
			if err = system.RemountSHM(uint64(val)); err != nil {
				return err
			}
		} else {
			system.WarningLog("SetMemVal: /dev/shm is not a valid mount point, will not adjust its size.")
		}
	}
	return err
}

// section [rpm]

// GetRpmVal initialise the rpm structure with the current system settings
func GetRpmVal(key string) string {
	keyFields := strings.Split(key, ":")
	instvers := system.GetRpmVers(keyFields[1])
	return instvers
}

// OptRpmVal returns the value from the configuration file
func OptRpmVal(key, cfgval string) string {
	// nothing to do, only checking for 'verify'
	return cfgval
}

// SetRpmVal nothing to do, only checking for 'verify'
func SetRpmVal(value string) error {
	// nothing to do, only checking for 'verify'
	return nil
}

// section [grub]

// GetGrubVal initialise the grub structure with the current system settings
func GetGrubVal(key string) string {
	keyFields := strings.Split(key, ":")
	val := system.ParseCmdline("/proc/cmdline", keyFields[1])
	return val
}

// OptGrubVal returns the value from the configuration file
func OptGrubVal(key, cfgval string) string {
	// nothing to do, only checking for 'verify'
	return cfgval
}

// SetGrubVal nothing to do, only checking for 'verify'
func SetGrubVal(value string) error {
	// nothing to do, only checking for 'verify'
	return nil
}

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

// section [login]

// GetLoginVal initialise the systemd login structure with the current
// system settings
func GetLoginVal(key string) (string, error) {
	var val string
	var utmPat = regexp.MustCompile(`UserTasksMax=(.*)`)
	switch key {
	case "UserTasksMax":
		logindContent, err := ioutil.ReadFile(path.Join(LogindConfDir, LogindSAPConfFile))
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		matches := utmPat.FindStringSubmatch(string(logindContent))
		if len(matches) != 0 {
			val = matches[1]
		} else {
			val = "NA"
		}
	}
	return val, nil
}

// OptLoginVal returns the value from the configuration file
func OptLoginVal(cfgval string) string {
	return strings.ToLower(cfgval)
}

// SetLoginVal applies the settings to the system
func SetLoginVal(key, value string, revert bool) error {
	switch key {
	case "UserTasksMax":
		// set limit per active user (for both - revert and apply)
		if value != "" && value != "NA" {
			for _, userID := range system.GetCurrentLogins() {
				if err := system.SetTasksMax(userID, value); err != nil {
					return err
				}
			}
		}
		// handle drop-in file
		if revert && IsLastNoteOfParameter(key) {
			// revert - remove logind drop-in file
			os.Remove(path.Join(LogindConfDir, LogindSAPConfFile))
			// reload-or-try-restart systemd-logind.service
			err := system.SystemctlReloadTryRestart("systemd-logind.service")
			return err
		}
		if value != "" && value != "NA" {
			// revert with value from another former applied note
			// or
			// apply - Prepare logind drop-in file
			// LogindSAPConfContent is the verbatim content of
			// SAP-specific logind settings file.
			LogindSAPConfContent := fmt.Sprintf("[Login]\nUserTasksMax=%s\n", value)
			if err := os.MkdirAll(LogindConfDir, 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path.Join(LogindConfDir, LogindSAPConfFile), []byte(LogindSAPConfContent), 0644); err != nil {
				return err
			}
			// reload-or-try-restart systemd-logind.service
			if err := system.SystemctlReloadTryRestart("systemd-logind.service"); err != nil {
				return err
			}
			if value == "infinity" {
				system.WarningLog("Be aware: system-wide UserTasksMax is now set to infinity according to SAP recommendations.\n" +
					"This opens up entire system to fork-bomb style attacks.")
			}
		}
	}
	return nil
}

// section [pagecache]

// GetPagecacheVal initialise the pagecache structure with the current
// system settings
func GetPagecacheVal(key string, cur *LinuxPagingImprovements) string {
	val := ""
	currentPagecache, err := LinuxPagingImprovements{PagingConfig: cur.PagingConfig}.Initialise()
	if err != nil {
		return ""
	}
	current := currentPagecache.(LinuxPagingImprovements)

	switch key {
	case "ENABLE_PAGECACHE_LIMIT":
		if current.VMPagecacheLimitMB == 0 {
			val = "no"
		} else {
			val = "yes"
		}
	case system.SysctlPagecacheLimitIgnoreDirty:
		val = strconv.Itoa(current.VMPagecacheLimitIgnoreDirty)
	case "OVERRIDE_PAGECACHE_LIMIT_MB":
		if current.VMPagecacheLimitMB == 0 {
			val = ""
		} else {
			val = strconv.FormatUint(current.VMPagecacheLimitMB, 10)
		}
	}
	*cur = current
	return val
}

// OptPagecacheVal optimises the pagecache structure with the settings
// from the configuration file or with a calculation
//func OptPagecacheVal(key, cfgval string, cur *LinuxPagingImprovements, keyvalue map[string]map[string]txtparser.INIEntry) string {
func OptPagecacheVal(key, cfgval string, cur *LinuxPagingImprovements) string {
	val := strings.ToLower(cfgval)

	switch key {
	case "ENABLE_PAGECACHE_LIMIT":
		if val != "yes" && val != "no" {
			system.WarningLog("wrong selection for ENABLE_PAGECACHE_LIMIT. Now set to default 'no'")
			val = "no"
		}
	case system.SysctlPagecacheLimitIgnoreDirty:
		if val != "2" && val != "1" && val != "0" {
			system.WarningLog("wrong selection for %s. Now set to default '1'", system.SysctlPagecacheLimitIgnoreDirty)
			val = "1"
		}
		cur.VMPagecacheLimitIgnoreDirty, _ = strconv.Atoi(val)
	case "OVERRIDE_PAGECACHE_LIMIT_MB":
		opt, _ := cur.Optimise()
		if opt == nil {
			_ = system.ErrorLog("page cache optimise had problems reading the Note definition file '%s'. Please check", cur.PagingConfig)
			return ""
		}
		optval := opt.(LinuxPagingImprovements).VMPagecacheLimitMB
		if optval != 0 {
			cur.VMPagecacheLimitMB = optval
			val = strconv.FormatUint(optval, 10)
		} else {
			cur.VMPagecacheLimitMB = 0
			val = ""
		}
	}

	return val
}

// SetPagecacheVal applies the settings to the system
func SetPagecacheVal(key string, cur *LinuxPagingImprovements) error {
	var err error
	if key == "OVERRIDE_PAGECACHE_LIMIT_MB" {
		err = cur.Apply()
	}
	return err
}
