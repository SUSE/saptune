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
	"sort"
	"strconv"
	"strings"
)

// section name definition
const (
	INISectionSysctl    = "sysctl"
	INISectionVM        = "vm"
	INISectionCPU       = "cpu"
	INISectionMEM       = "mem"
	INISectionBlock     = "block"
	INISectionUuidd     = "uuidd"
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

// GetServiceName returns the systemd service name for supported services
func GetServiceName(service string) string {
	switch service {
	case "UuiddSocket":
		service = "uuidd.socket"
	case "Sysstat":
		service = "sysstat"
	default:
		system.WarningLog("skipping unkown service '%s'", service)
		service = ""
	}
	return service
}

// section handling
// section [sysctl]

// OptSysctlVal optimises a sysctl parameter value
// use exactly the value from the config file. No calculation any more
func OptSysctlVal(operator txtparser.Operator, key, actval, cfgval string) string {
	if actval == "" {
		// sysctl parameter not available in system
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

// BlockDeviceQueue defines block device structures
type BlockDeviceQueue struct {
	BlockDeviceSchedulers param.BlockDeviceSchedulers
	BlockDeviceNrRequests param.BlockDeviceNrRequests
}

// GetBlkVal initialise the block device structure with the current
// system settings
func GetBlkVal(key string) (string, error) {
	newQueue := make(map[string]string)
	newReq := make(map[string]int)
	retVal := ""
	switch key {
	case "IO_SCHEDULER":
		newIOQ, err := BlockDeviceQueue{}.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return "", err
		}
		newQueue = newIOQ.(param.BlockDeviceSchedulers).SchedulerChoice
		for k, v := range newQueue {
			retVal = retVal + fmt.Sprintf("%s@%s ", k, v)
		}
	case "NRREQ":
		newNrR, err := BlockDeviceQueue{}.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return "", err
		}
		newReq = newNrR.(param.BlockDeviceNrRequests).NrRequests
		for k, v := range newReq {
			retVal = retVal + fmt.Sprintf("%s@%s ", k, strconv.Itoa(v))
		}
	}
	fields := strings.Fields(retVal)
	sort.Strings(fields)
	retVal = strings.Join(fields, " ")
	return retVal, nil
}

// OptBlkVal optimises the block device structure with the settings
// from the configuration file
func OptBlkVal(key, actval, cfgval string) string {
	sval := cfgval
	val := actval
	retVal := ""
	switch key {
	case "IO_SCHEDULER":
		sval = strings.ToLower(cfgval)
	case "NRREQ":
		if sval == "0" {
			sval = "1024"
		}
	}
	for _, entry := range strings.Fields(val) {
		fields := strings.Split(entry, "@")
		if retVal == "" {
			retVal = retVal + fmt.Sprintf("%s@%s", fields[0], sval)
		} else {
			retVal = retVal + " " + fmt.Sprintf("%s@%s", fields[0], sval)
		}
	}
	return retVal
}

// SetBlkVal applies the settings to the system
func SetBlkVal(key, value string) error {
	var err error

	switch key {
	case "IO_SCHEDULER":
		setIOQ, err := BlockDeviceQueue{}.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return err
		}

		for _, entry := range strings.Fields(value) {
			fields := strings.Split(entry, "@")
			setIOQ.(param.BlockDeviceSchedulers).SchedulerChoice[fields[0]] = fields[1]
		}
		err = setIOQ.(param.BlockDeviceSchedulers).Apply()
		if err != nil {
			return err
		}
	case "NRREQ":
		setNrR, err := BlockDeviceQueue{}.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return err
		}

		for _, entry := range strings.Fields(value) {
			fields := strings.Split(entry, "@")
			NrR, _ := strconv.Atoi(fields[1])
			setNrR.(param.BlockDeviceNrRequests).NrRequests[fields[0]] = NrR
		}
		err = setNrR.(param.BlockDeviceNrRequests).Apply()
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
	lim := strings.Fields(value)
	// dom=[0], type=[1], item=[2], value=[3]
	// no check, that the syntax/order of the entry in the config file is
	// a valid limits entry

	// Find out current limits
	limit := ""
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
	return limit, nil
}

// OptLimitsVal optimises the security limit structure with the settings
// from the configuration file or with a calculation
func OptLimitsVal(actval, cfgval string) string {
	lim := strings.Fields(cfgval)

	//ANGI - check, if we will preserve 'unlimited' or if we set value from config
	actlim := strings.Fields(actval)
	if actlim[3] == "unlimited" || actlim[3] == "infinity" || actlim[3] == "-1" {
		lim[3] = actlim[3]
	}

	return strings.Join(lim, " ")
}

// SetLimitsVal applies the settings to the system
func SetLimitsVal(key, noteID, value string, revert bool) error {
	lim := strings.Fields(value)
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
func GetCPUVal(key string) string {
	var val string
	switch key {
	case "force_latency":
		val = system.GetForceLatency()
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
	return strings.TrimSpace(val)
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
	sval = strings.TrimSpace(rval)
	return sval
}

// SetCPUVal applies the settings to the system
func SetCPUVal(key, value string, revert bool, noteID string) error {
	var err error
	switch key {
	case "force_latency":
		//iVal, _ := strconv.Atoi(value)
		err = system.SetForceLatency(noteID, value, revert)
	case "energy_perf_bias":
		err = system.SetPerfBias(value)
	case "governor":
		err = system.SetGovernor(value)
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
func OptMemVal(key, actval, cfgval, shmsize, tmpfspercent string) string {
	// shmsize       value of ShmFileSystemSizeMB from config file
	// tmpfspercent  value of VSZ_TMPFS_PERCENT from config file
	var size uint64
	var ret string

	if actval == "-1" {
		system.WarningLog("OptMemVal: /dev/shm is not a valid mount point, will not calculate its optimal size.")
		size = 0
	} else if shmsize == "0" {
		if tmpfspercent == "0" {
			// Calculate optimal SHM size (TotalMemSizeMB*75/100) (SAP-Note 941735)
			size = uint64(system.GetTotalMemSizeMB()) * 75 / 100
		} else {
			// Calculate optimal SHM size (TotalMemSizeMB*VSZ_TMPFS_PERCENT/100)
			val, _ := strconv.ParseUint(tmpfspercent, 10, 64)
			size = uint64(system.GetTotalMemSizeMB()) * val / 100
		}
	} else {
		size, _ = strconv.ParseUint(shmsize, 10, 64)
	}
	switch key {
	case "VSZ_TMPFS_PERCENT":
		ret = cfgval
	case "ShmFileSystemSizeMB":
		if size == 0 {
			ret = "-1"
		} else {
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

// section [uuidd]

// GetUuiddVal initialise the uuidd socket service structure with the current
// system settings
func GetUuiddVal() string {
	var val string
	if system.SystemctlIsRunning("uuidd.socket") {
		val = "start"
	} else {
		val = "stop"
	}
	return val
}

// OptUuiddVal optimises the uuidd socket service structure with the settings
// from the configuration file
func OptUuiddVal(cfgval string) string {
	sval := strings.ToLower(cfgval)
	if sval != "start" {
		system.WarningLog("wrong selection for UuiddSocket. Now set to 'start' to start the uuid daemon")
		sval = "start"
	}
	return sval
}

// SetUuiddVal applies the settings to the system
func SetUuiddVal(value string) error {
	var err error
	if !system.SystemctlIsRunning("uuidd.socket") {
		err = system.SystemctlStart("uuidd.socket")
	}
	return err
}

// section [service]

// GetServiceVal initialise the systemd service structure with the current
// system settings
func GetServiceVal(key string) string {
	var val string
	service := GetServiceName(key)
	if service == "" {
		return ""
	}
	if system.SystemctlIsRunning(service) {
		val = "start"
	} else {
		val = "stop"
	}
	return val
}

// OptServiceVal optimises the systemd service structure with the settings
// from the configuration file
func OptServiceVal(key, cfgval string) string {
	sval := strings.ToLower(cfgval)
	switch key {
	case "UuiddSocket":
		if sval != "start" {
			system.WarningLog("wrong selection for '%s'. Now set to 'start' to start the service\n", key)
			sval = "start"
		}
	case "Sysstat":
		if sval != "start" && sval != "stop" {
			system.WarningLog("wrong selection for '%s'. Now set to 'start' to start the service\n", key)
			sval = "start"
		}
	default:
		system.WarningLog("skipping unkown service '%s'", key)
		return ""
	}
	return sval
}

// SetServiceVal applies the settings to the system
func SetServiceVal(key, value string) error {
	var err error
	service := GetServiceName(key)
	if service == "" {
		return nil
	}
	if value == "start" && !system.SystemctlIsRunning(service) {
		err = system.SystemctlStart(service)
	}
	if value == "stop" {
		if service == "uuidd.socket" {
			if !system.SystemctlIsRunning(service) {
				err = system.SystemctlStart(service)
			}
		} else {
			if system.SystemctlIsRunning(service) {
				err = system.SystemctlStop(service)
			}
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
		if revert && IsLastNoteOfParameter("UserTasksMax") {
			// revert - remove logind drop-in file
			os.Remove(path.Join(LogindConfDir, LogindSAPConfFile))
			// restart systemd-logind.service
			if err := system.SystemctlRestart("systemd-logind.service"); err != nil {
				return err
			}
			return nil
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
			// restart systemd-logind.service
			if err := system.SystemctlRestart("systemd-logind.service"); err != nil {
				return err
			}
			if value == "infinity" {
				system.WarningLog("Be aware: system-wide UserTasksMax is now set to infinity according to SAP recommendations.\n" +
					"This opens up entire system to fork-bomb style attacks.")
			}
			// set per user
			for _, userID := range system.GetCurrentLogins() {
				//oldLimit := system.GetTasksMax(userID)
				if err := system.SetTasksMax(userID, value); err != nil {
					return err
				}
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
	currentPagecache, err := LinuxPagingImprovements{SysconfigPrefix: cur.SysconfigPrefix}.Initialise()
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
	case "PAGECACHE_LIMIT_IGNORE_DIRTY":
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
	case "PAGECACHE_LIMIT_IGNORE_DIRTY":
		if val != "2" && val != "1" && val != "0" {
			system.WarningLog("wrong selection for PAGECACHE_LIMIT_IGNORE_DIRTY. Now set to default '1'")
			val = "1"
		}
		cur.VMPagecacheLimitIgnoreDirty, _ = strconv.Atoi(val)
	case "OVERRIDE_PAGECACHE_LIMIT_MB":
		opt, _ := cur.Optimise()
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
