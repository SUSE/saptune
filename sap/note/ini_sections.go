package note

import (
	"fmt"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	INISectionSysctl    = "sysctl"
	INISectionVM        = "vm"
	INISectionCPU       = "cpu"
	INISectionMEM       = "mem"
	INISectionBlock     = "block"
	INISectionUuidd     = "uuidd"
	INISectionLimits    = "limits"
	INISectionLogin     = "login"
	INISectionVersion   = "version"
	INISectionPagecache = "pagecache"
	INISectionReminder  = "reminder"
	SysKernelTHPEnabled = "kernel/mm/transparent_hugepage/enabled"
	SysKSMRun           = "kernel/mm/ksm/run"

	// LoginConfDir is the path to systemd's logind configuration directory under /etc.
	LogindConfDir = "/etc/systemd/logind.conf.d"
	// LogindSAPConfFile is a configuration file full of SAP-specific settings for logind.
	LogindSAPConfFile = "sap.conf"
	// LogindSAAPConfContent is the verbatim content of SAP-specific logind settings file.
	LogindSAPConfContent = `[Login]
UserTasksMax=infinity
`
)

// section handling
// section [sysctl]
func OptSysctlVal(operator txtparser.Operator, key, actval, cfgval string) string {
	allFieldsC := strings.Fields(actval)
	allFieldsE := strings.Fields(cfgval)
	allFieldsS := ""

	if len(allFieldsC) != len(allFieldsE) && (operator == txtparser.OperatorEqual || len(allFieldsE) > 1) {
		log.Printf("wrong number of fields given in the config file for parameter '%s'\n", key)
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

		optimisedValue, err := CalculateOptimumValue(operator, fieldC, fieldE)
		//optimisedValue, err := CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
		if err != nil {
			return ""
		}
		allFieldsS = allFieldsS + optimisedValue + "\t"
	}

	return strings.TrimSpace(allFieldsS)
}

// section [block]
type BlockDeviceQueue struct {
	BlockDeviceSchedulers param.BlockDeviceSchedulers
	BlockDeviceNrRequests param.BlockDeviceNrRequests
}

func GetBlkVal(key string) (string, error) {
	newQueue := make(map[string]string)
	newReq := make(map[string]int)
	ret_val := ""
	switch key {
	case "IO_SCHEDULER":
		newIOQ, err := BlockDeviceQueue{}.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return "", err
		}
		newQueue = newIOQ.(param.BlockDeviceSchedulers).SchedulerChoice
		for k, v := range newQueue {
			ret_val = ret_val + fmt.Sprintf("%s@%s ", k, v)
		}
	case "NRREQ":
		newNrR, err := BlockDeviceQueue{}.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return "", err
		}
		newReq = newNrR.(param.BlockDeviceNrRequests).NrRequests
		for k, v := range newReq {
			ret_val = ret_val + fmt.Sprintf("%s@%s ", k, strconv.Itoa(v))
		}
	}
	fields := strings.Fields(ret_val)
	sort.Strings(fields)
	ret_val = strings.Join(fields, " ")
	return ret_val, nil
}

func OptBlkVal(key, actval, cfgval string) string {
	sval := cfgval
	val := actval
	ret_val := ""
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
		if ret_val == "" {
			ret_val = ret_val + fmt.Sprintf("%s@%s", fields[0], sval)
		} else {
			ret_val = ret_val + " " + fmt.Sprintf("%s@%s", fields[0], sval)
		}
	}
	return ret_val
}

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
func GetLimitsVal(key, item, domain string) (string, error) {
	// Find out current limits
	limit := ""
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return "", err
	}

	for _, dom := range strings.Fields(domain) {
		lim := ""

		switch key {
		case "LIMIT_HARD":
			//lim, _ := secLimits.Get(domain, "hard", item)
			lim, _ := secLimits.Get(dom, "hard", item)
			limit = limit + fmt.Sprintf("%s:%s ", dom, lim)
		case "LIMIT_SOFT":
			//lim, _ := secLimits.Get(domain, "soft", item)
			lim, _ := secLimits.Get(dom, "soft", item)
			limit = limit + fmt.Sprintf("%s:%s ", dom, lim)
		case "LIMIT_ITEM":
			//_, isset := secLimits.Get(domain, "soft", item)
			_, isset := secLimits.Get(dom, "hard", item)
			if isset {
				lim = item
			} else {
				lim = ""
			}
			limit = limit + fmt.Sprintf("%s:%s ", dom, lim)
		case "LIMIT_DOMAIN":
			//_, isset := secLimits.Get(domain, "soft", item)
			_, isset := secLimits.Get(dom, "hard", item)
			if isset {
				lim = dom
			} else {
				lim = ""
			}
			limit = limit + fmt.Sprintf("%s ", lim)
		}
	}
	return limit, nil
}

func OptLimitsVal(key, actval, cfgval, item, domain string) string {
	limit := cfgval
	lim := ""

	for _, dom := range strings.Fields(domain) {
		if key == "LIMIT_HARD" || key == "LIMIT_SOFT" {
			for _, entry := range strings.Fields(actval) {
				fields := strings.Split(entry, ":")
				if fields[0] != dom {
					continue
				}
				limact, _ := strconv.Atoi(fields[1])
				if limit == "0" && item == "memlock" {
					//calculate limit (RAM in KB - 10%)
					memlock := system.GetMainMemSizeMB()*1024 - (system.GetMainMemSizeMB() * 1024 * 10 / 100)
					limit = strconv.Itoa(param.MaxI(limact, int(memlock)))
				} else {
					limcfg, _ := strconv.Atoi(cfgval)
					limit = strconv.Itoa(param.MaxI(limact, limcfg))
				}
			}
			lim = lim + fmt.Sprintf("%s:%s ", dom, limit)
		}
		if key == "LIMIT_ITEM" {
			lim = lim + fmt.Sprintf("%s:%s ", dom, item)
		}
		if key == "LIMIT_DOMAIN" {
			lim = lim + fmt.Sprintf("%s ", dom)
		}
	}
	return lim
}

func SetLimitsVal(key, value, item string) error {
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return err
	}

	for _, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		switch key {
		case "LIMIT_HARD":
			secLimits.Set(fields[0], "hard", item, fields[1])
		case "LIMIT_SOFT":
			secLimits.Set(fields[0], "soft", item, fields[1])
		default:
			return nil
		}
	}

	err = secLimits.Apply()
	return err
}

// section [vm]
// Manipulate /sys/kernel/mm switches.
func GetVmVal(key string) string {
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

func OptVmVal(key, cfgval string) string {
	val := strings.ToLower(cfgval)
	switch key {
	case "THP":
		if val != "always" && val != "madvise" && val != "never" {
			log.Print("wrong selection for THP. Now set to 'never' to disable transarent huge pages")
			val = "never"
		}
	case "KSM":
		if val != "1" && val != "0" {
			log.Print("wrong selection for KSM. Now set to default value '0'")
			val = "0"
		}
	}
	return val
}

func SetVmVal(key, value string) error {
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
func GetCpuVal(key string) string {
	var val string
	switch key {
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

func OptCpuVal(key, actval, cfgval string) string {
	sval := strings.ToLower(cfgval)
	rval := ""
	cpu := ""
	val := "0"
	switch key {
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
			log.Print("wrong selection for energy_perf_bias. Now set to 'performance'")
			val = "0"
		}
		for _, entry := range strings.Fields(actval) {
			fields := strings.Split(entry, ":")
			if fields[1] == "none" {
				//System does not support Intel's performance bias setting
				val = fields[1]
			}
			rval = rval + fmt.Sprintf("%s:%s ", fields[0], val)
		}
		sval = strings.TrimSpace(rval)
	case "governor":
		val = sval
		for _, entry := range strings.Fields(actval) {
			fields := strings.Split(entry, ":")
			if fields[1] == "none" {
				//System does not support a scaling governor
				val = fields[1]
			} else {
				if fields[0] == "all" {
					cpu = "cpu0"
				} else {
					cpu = fields[0]
				}
				if !system.IsValidGovernor(cpu, sval) {
					log.Printf("'%s' is not a valid governor. Set value to 'none' for ignoring", sval)
					val = "none"
				}
			}
			rval = rval + fmt.Sprintf("%s:%s ", fields[0], val)
		}
		sval = strings.TrimSpace(rval)
	}
	return sval
}

func SetCpuVal(key, value string, revert bool) error {
	var err error
	switch key {
	case "energy_perf_bias":
		err = system.SetPerfBias(value)
	case "governor":
		err = system.SetGovernor(value)
	}

	return err
}

// section [mem]
func GetMemVal(key string) string {
	var val string
	switch key {
	case "KernelShmMax":
		val, _ = system.GetSysctlString(system.SysctlShmax)
	case "KernelShmAll":
		val, _ = system.GetSysctlString(system.SysctlShmall)
	case "KernelShmMni":
		val, _ = system.GetSysctlString(system.SysctlShmni)
	case "VMMaxMapCount":
		val, _ = system.GetSysctlString(system.SysctlMaxMapCount)
	case "ShmFileSystemSizeMB":
		// Find out size of SHM
		mount, found := system.ParseProcMounts().GetByMountPoint("/dev/shm")
		if found {
			val = strconv.FormatUint(mount.GetFileSystemSizeMB(), 10)
		} else {
			log.Print("GetMemVal: failed to find /dev/shm mount point")
			val = "-1"
		}
	}
	return val
}

func OptMemVal(key, actval, cfgval string) string {
	var curval, val uint64

	val, _ = strconv.ParseUint(cfgval, 10, 64)
	curval, _ = strconv.ParseUint(actval, 10, 64)
	switch key {
	case "KernelShmMax":
		//KernelShmMax=0 - max of (system, TotalMemSizeMB*1048576, 20*1024*1024*1024)
		if val == 0 {
			// Calculate - max of (system, TotalMemSizeMB*1048576, 20*1024*1024*1024)
			val = param.MaxU64(curval, system.GetTotalMemSizeMB()*1048576 /* MB to Bytes */, 20*1024*1024*1024)
		} else {
			// max of (system, value from sap note file)
			val = param.MaxU64(curval, val)
		}
	case "KernelShmAll":
		//KernelShmAll=0 - max of (system, TotalMemSizePages)
		if val == 0 {
			// Calculate - max of (system, TotalMemSizePages)
			val = param.MaxU64(curval, system.GetTotalMemSizePages())
		} else {
			// max of (system, value from sap note file)
			val = param.MaxU64(curval, val)
		}
	case "KernelShmMni":
		//KernelShmMni >= 2048 - max of (system, KernelShmMni)
		// max of (system, value from sap note file)
		val = param.MaxU64(curval, val)
	case "VMMaxMapCount":
		//VMMaxMapCount >= 2147483647
		val = param.MaxU64(curval, val)
	case "ShmFileSystemSizeMB":
		// ShmFileSystemSizeMB=0 - max of (system, TotalMemSizeMB*75/100) (SAP-Note 941735)
		// ShmFileSystemSizeMB>0 - max of (system, ShmFileSystemSizeMB)
		if curval < 0 {
			log.Print("OptMemVal: /dev/shm is not a valid mount point, will not calculate its optimal size.")
		} else if val == 0 {
			// Calculate optimal SHM size (TotalMemSizeMB*75/100)
			val = param.MaxU64(curval, uint64(system.GetTotalMemSizeMB())*75/100)
		} else {
			// max of (system, value from sap note file)
			val = param.MaxU64(curval, val)
		}
	}
	return strconv.FormatUint(val, 10)
}

func SetMemVal(key, value string) error {
	var err error
	switch key {
	case "KernelShmMax":
		err = system.SetSysctlString(system.SysctlShmax, value)
	case "KernelShmAll":
		err = system.SetSysctlString(system.SysctlShmall, value)
	case "KernelShmMni":
		err = system.SetSysctlString(system.SysctlShmni, value)
	case "VMMaxMapCount":
		err = system.SetSysctlString(system.SysctlMaxMapCount, value)
	case "ShmFileSystemSizeMB":
		// ShmFileSystemSizeMB=0 - max of (system, TotalMemSizeMB*75/100) (SAP-Note 941735)
		// ShmFileSystemSizeMB>0 - max of (system, ShmFileSystemSizeMB)
		val, err := strconv.ParseUint(value, 10, 64)
		if val > 0 {
			if err = system.RemountSHM(uint64(val)); err != nil {
				return err
			}
		} else {
			log.Print("SetMemVal: /dev/shm is not a valid mount point, will not adjust its size.")
		}
	}
	return err
}

// section [uuidd]
func GetUuiddVal() string {
	var val string
	if system.SystemctlIsRunning("uuidd.socket") {
		val = "start"
	} else {
		val = "stop"
	}
	return val
}

func OptUuiddVal(cfgval string) string {
	sval := strings.ToLower(cfgval)
	if sval != "start" && sval != "stop" {
		fmt.Println("wrong selection for UuiddSocket. Now set to 'start' to start the uuid daemon")
		sval = "start"
	}
	return sval
}

func SetUuiddVal(value string) error {
	var err error
	if value == "start" {
		err = system.SystemctlEnableStart("uuidd.socket")
	} else {
		err = system.SystemctlDisableStop("uuidd.socket")
	}
	return err
}

// section [login]
func GetLoginVal(key string) (string, error) {
	var val string
	switch key {
	case "UserTasksMax":
		logindContent, err := ioutil.ReadFile(path.Join(LogindConfDir, LogindSAPConfFile))
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if string(logindContent) == LogindSAPConfContent {
			val = "infinity"
		}
	}
	return val, nil
}

func OptLoginVal(cfgval string) string {
	return strings.ToLower(cfgval)
}

func SetLoginVal(key, value string, revert bool) error {
	switch key {
	case "UserTasksMax":
		// Prepare logind config file
		if err := os.MkdirAll(LogindConfDir, 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(LogindConfDir, LogindSAPConfFile), []byte(LogindSAPConfContent), 0644); err != nil {
			return err
		}
		if !revert {
			log.Print("Be aware: system-wide UserTasksMax is now set to infinity according to SAP recommendations.\n" +
				"This opens up entire system to fork-bomb style attacks. Please reboot the system for the changes to take effect.")
		}
	}
	return nil
}

// section [pagecache]
func GetPagecacheVal(key string, cur *LinuxPagingImprovements) string {
	val := ""
	currentPagecache, err := LinuxPagingImprovements{}.Initialise()
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
	case "TUNE_FOR_HANA":
		if current.UseAlgorithmForHANA {
			val = "yes"
		} else {
			val = "no"
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

func OptPagecacheVal(key, cfgval string, cur *LinuxPagingImprovements, keyvalue map[string]map[string]txtparser.INIEntry) string {
	val := strings.ToLower(cfgval)

	switch key {
	case "ENABLE_PAGECACHE_LIMIT":
		if val != "yes" && val != "no" {
			log.Print("wrong selection for ENABLE_PAGECACHE_LIMIT. Now set to default 'no'")
			val = "no"
		}
	case "TUNE_FOR_HANA":
		if val != "yes" && val != "no" {
			log.Print("wrong selection for TUNE_FOR_HANA. Now set to default 'no'")
			val = "no"
		}
	case "PAGECACHE_LIMIT_IGNORE_DIRTY":
		if val != "2" && val != "1" && val != "0" {
			log.Print("wrong selection for PAGECACHE_LIMIT_IGNORE_DIRTY. Now set to default '1'")
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

func SetPagecacheVal(key string, cur *LinuxPagingImprovements) error {
	var err error
	if key == "OVERRIDE_PAGECACHE_LIMIT_MB" {
		err = cur.Apply()
	}
	return err
}
