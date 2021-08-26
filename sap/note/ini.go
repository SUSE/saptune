package note

import (
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// and section name definition
const (
	INISectionSysctl    = "sysctl"
	INISectionSys       = "sys"
	INISectionVM        = "vm"
	INISectionFS        = "filesystem"
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

	// LoginConfDir is the path to systemd's logind configuration directory under /etc.
	LogindConfDir = "/etc/systemd/logind.conf.d"
	// LogindSAPConfFile is a configuration file full of SAP-specific settings for logind.
	LogindSAPConfFile = "saptune-UserTasksMax.conf"
)

var ini *txtparser.INIFile
var pc = LinuxPagingImprovements{}

var blck = param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}, BlockDeviceReadAheadKB: param.BlockDeviceReadAheadKB{ReadAheadKB: make(map[string]int)}, BlockDeviceMaxSectorsKB: param.BlockDeviceMaxSectorsKB{MaxSectorsKB: make(map[string]int)}}
var isLimitSoft = regexp.MustCompile(`LIMIT_.*_soft_memlock`)
var isLimitHard = regexp.MustCompile(`LIMIT_.*_hard_memlock`)
var flstates = ""

// Tuning options composed by a third party vendor.

// INISettings defines tuning options composed by a third party vendor.
type INISettings struct {
	ConfFilePath    string            // Full path to the 3rd party vendor's tuning configuration file
	ID              string            // ID portion of the tuning configuration
	DescriptiveName string            // Descriptive name portion of the tuning configuration
	SysctlParams    map[string]string // Sysctl parameter values from the computer system
	ValuesToApply   map[string]string // values to apply
	OverrideParams  map[string]string // parameter values from the override file
	Inform          map[string]string // special information for parameter values
}

// Name returns the name of the related SAP Note or en empty string
func (vend INISettings) Name() string {
	if len(vend.DescriptiveName) == 0 {
		vend.DescriptiveName = txtparser.GetINIFileDescriptiveName(vend.ConfFilePath)
	}
	return vend.DescriptiveName
}

// Initialise retrieves the current parameter values from the system
func (vend INISettings) Initialise() (Note, error) {
	ini, err := txtparser.GetSectionInfo("sns", vend.ID, false)
	if err != nil {
		// Parse the configuration file
		ini, err = txtparser.ParseINIFile(vend.ConfFilePath, false)
		if err != nil {
			return vend, err
		}
		// write section data to section runtime file
		err = txtparser.StoreSectionInfo(ini, "run", vend.ID, true)
		if err != nil {
			system.ErrorLog("Problems during storing of section information")
			return vend, err
		}
	}

	// looking for override file
	override, ow := txtparser.GetOverrides("ovw", vend.ID)

	// Read current parameter values
	vend.SysctlParams = make(map[string]string)
	vend.OverrideParams = make(map[string]string)
	vend.Inform = make(map[string]string)
	pc = LinuxPagingImprovements{}
	blck = param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}, BlockDeviceReadAheadKB: param.BlockDeviceReadAheadKB{ReadAheadKB: make(map[string]int)}, BlockDeviceMaxSectorsKB: param.BlockDeviceMaxSectorsKB{MaxSectorsKB: make(map[string]int)}}

	for _, param := range ini.AllValues {
		if override && len(ow.KeyValue[param.Section]) != 0 {
			param.Key, param.Value, param.Operator = vend.handleInitOverride(param.Key, param.Value, param.Section, param.Operator, ow)
		}

		switch param.Section {
		case INISectionSysctl:
			vend.Inform[param.Key] = ""
			vend.SysctlParams[param.Key], _ = system.GetSysctlString(param.Key)
		case INISectionSys:
			vend.SysctlParams[param.Key], vend.Inform[param.Key] = GetSysVal(param.Key)
		case INISectionVM:
			vend.SysctlParams[param.Key], vend.Inform[param.Key] = GetVMVal(param.Key)
		case INISectionFS:
			vend.SysctlParams[param.Key], vend.Inform[param.Key] = GetFSVal(param.Key, param.Value)
			continue
		case INISectionBlock:
			vend.SysctlParams[param.Key], vend.Inform[param.Key], _ = GetBlkVal(param.Key, &blck)
		case INISectionLimits:
			vend.SysctlParams[param.Key], vend.Inform[param.Key], _ = GetLimitsVal(param.Value)
		case INISectionService:
			vend.SysctlParams[param.Key] = GetServiceVal(param.Key)
		case INISectionLogin:
			vend.SysctlParams[param.Key], _ = GetLoginVal(param.Key)
		case INISectionMEM:
			vend.SysctlParams[param.Key] = GetMemVal(param.Key)
		case INISectionCPU:
			vend.SysctlParams[param.Key], flstates, vend.Inform[param.Key] = GetCPUVal(param.Key)
		case INISectionRpm:
			vend.SysctlParams[param.Key] = GetRpmVal(param.Key)
			continue
		case INISectionGrub:
			vend.SysctlParams[param.Key] = GetGrubVal(param.Key)
			continue
		case INISectionReminder:
			vend.SysctlParams[param.Key] = param.Value
			continue
		case INISectionVersion:
			continue
		case INISectionPagecache:
			// page cache is special, has it's own config file
			// so adjust path to pagecache config file, if needed
			if override {
				pc.PagingConfig = path.Join(txtparser.OverrideTuningSheets, vend.ID)
			} else {
				pc.PagingConfig = vend.ConfFilePath
			}
			vend.SysctlParams[param.Key] = GetPagecacheVal(param.Key, &pc)
		default:
			system.WarningLog("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
		// create parameter saved state file, if NOT in 'verify'
		vend.createParamSavedStates(param.Key, flstates)
	}
	return vend, nil
}

// Optimise gets the expected parameter values from the configuration
func (vend INISettings) Optimise() (Note, error) {
	blckOK := make(map[string][]string)
	scheds := ""
	next := false

	// read saved section data == config data from configuration file
	ini, err := txtparser.GetSectionInfo("sns", vend.ID, false)
	if err != nil {
		// fallback, parse the configuration file
		ini, err = txtparser.ParseINIFile(vend.ConfFilePath, false)
		if err != nil {
			return vend, err
		}
		// write section data to section runtime file
		err = txtparser.StoreSectionInfo(ini, "run", vend.ID, true)
		if err != nil {
			system.ErrorLog("Problems during storing of section information")
			return vend, err
		}
	}

	for _, param := range ini.AllValues {
		// Compare current values against INI's definition
		// handle note 1805750
		param.Key, param.Value = vend.handleID1805750(param.Key, param.Value)
		// check, if we should use the value from override file
		next, scheds, param.Value = vend.useOverrides(param.Key, scheds, param.Value)
		if next {
			continue
		}

		switch param.Section {
		case INISectionSysctl:
			//optimisedValue, err := txtparser.CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			//vend.SysctlParams[param.Key] = optimisedValue
			vend.Inform[param.Key] = system.ChkForSysctlDoubles(param.Key)
			vend.SysctlParams[param.Key] = OptSysctlVal(param.Operator, param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionSys:
			vend.Inform[param.Key] = vend.chkDoubles(param.Key, vend.Inform[param.Key])
			vend.SysctlParams[param.Key] = OptSysVal(param.Operator, param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionVM:
			vend.Inform[param.Key] = vend.chkDoubles(param.Key, vend.Inform[param.Key])
			vend.SysctlParams[param.Key] = OptVMVal(param.Key, param.Value)
		case INISectionFS:
			vend.SysctlParams[param.Key] = OptFSVal(param.Key, param.Value)
			continue
		case INISectionBlock:
			vend.SysctlParams[param.Key], vend.Inform[param.Key] = OptBlkVal(param.Key, param.Value, &blck, blckOK)
			vend.Inform[param.Key] = vend.chkDoubles(param.Key, vend.Inform[param.Key])
			if system.IsSched.MatchString(param.Key) {
				scheds = param.Value
			}
		case INISectionLimits:
			vend.SysctlParams[param.Key] = OptLimitsVal(vend.SysctlParams[param.Key], param.Value)
		case INISectionService:
			vend.SysctlParams[param.Key] = OptServiceVal(param.Key, param.Value)
		case INISectionLogin:
			vend.SysctlParams[param.Key] = OptLoginVal(param.Value)
		case INISectionMEM:
			if vend.OverrideParams["VSZ_TMPFS_PERCENT"] == "untouched" || vend.OverrideParams["VSZ_TMPFS_PERCENT"] == "" {
				vend.SysctlParams[param.Key] = OptMemVal(param.Key, vend.SysctlParams[param.Key], param.Value, ini.KeyValue["mem"]["VSZ_TMPFS_PERCENT"].Value)
			} else {
				vend.SysctlParams[param.Key] = OptMemVal(param.Key, vend.SysctlParams[param.Key], param.Value, vend.OverrideParams["VSZ_TMPFS_PERCENT"])
			}
		case INISectionCPU:
			vend.SysctlParams[param.Key] = OptCPUVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionRpm:
			vend.SysctlParams[param.Key] = OptRpmVal(param.Key, param.Value)
			continue
		case INISectionGrub:
			vend.SysctlParams[param.Key] = OptGrubVal(param.Key, param.Value)
			continue
		case INISectionReminder:
			vend.SysctlParams[param.Key] = param.Value
			continue
		case INISectionVersion:
			continue
		case INISectionPagecache:
			vend.SysctlParams[param.Key] = OptPagecacheVal(param.Key, param.Value, &pc)
		default:
			system.WarningLog("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
		// add values to parameter saved state file, if NOT in 'verify'
		vend.addParamSavedStates(param.Key)
	}

	// print info about used block scheduler
	vend.printSchedInfo(scheds, blckOK)

	// write section data to section store file, if NOT in 'verify'
	// will cover the situation where a note fully conforms with the
	// system, so that there is NO apply operation, but later a
	// revert may happen
	if _, ok := vend.ValuesToApply["verify"]; !ok {
		// this code section was moved from function 'Apply'
		err = txtparser.StoreSectionInfo(ini, "section", vend.ID, true)
		if err != nil {
			system.ErrorLog("Problems during storing of section information")
			return vend, err
		}
	}
	return vend, nil
}

// Apply sets the new parameter values in the system or
// revert the system to the former parameter values
func (vend INISettings) Apply() error {
	var err error
	errs := make([]error, 0, 0)
	revertValues := false
	pvendID := vend.ID

	if len(vend.ValuesToApply) == 0 {
		// nothing to apply
		return nil
	}
	if _, ok := vend.ValuesToApply["revert"]; ok {
		revertValues = true
	}

	ini, err = txtparser.GetSectionInfo("sns", vend.ID, revertValues)
	if err != nil {
		// fallback, reading info from config file
		ini, err = txtparser.ParseINIFile(vend.ConfFilePath, false)
		if err != nil {
			return err
		}
	}

	for _, param := range ini.AllValues {
		// handle note 1805750
		param.Key, param.Value = vend.handleID1805750(param.Key, param.Value)
		switch param.Section {
		case INISectionVersion, INISectionRpm, INISectionGrub, INISectionFS, INISectionReminder:
			// These parameters are only checked, but not applied.
			// So nothing to do during apply and no need for revert
			continue
		}

		if _, ok := vend.ValuesToApply[param.Key]; !ok && !revertValues {
			continue
		}

		if revertValues && vend.SysctlParams[param.Key] != "" {
			// revert parameter value
			pvendID, flstates = vend.setRevertParamValues(param.Key)
		}

		switch param.Section {
		case INISectionSysctl:
			// Apply sysctl parameters
			// for the vm.dirty parameters take the counterpart
			// parameters into account (only during revert)
			// if vm.dirty_background_bytes is set to a value != 0,
			// vm.dirty_background_ratio is set to 0 and vice versa
			// if vm.dirty_bytes is set to a value != 0,
			// vm.dirty_ratio is set to 0 and vice versa
			key, val := vend.getCounterPart(param.Key, revertValues)
			errs = append(errs, system.SetSysctlString(key, val))
		case INISectionSys:
			errs = append(errs, SetSysVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionVM:
			errs = append(errs, SetVMVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionBlock:
			errs = append(errs, SetBlkVal(param.Key, vend.SysctlParams[param.Key], &blck, revertValues))
		case INISectionLimits:
			errs = append(errs, SetLimitsVal(param.Key, pvendID, vend.SysctlParams[param.Key], revertValues))
		case INISectionService:
			errs = append(errs, SetServiceVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionLogin:
			errs = append(errs, SetLoginVal(param.Key, vend.SysctlParams[param.Key], revertValues))
		case INISectionMEM:
			errs = append(errs, SetMemVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionCPU:
			errs = append(errs, SetCPUVal(param.Key, vend.SysctlParams[param.Key], vend.ID, flstates, vend.OverrideParams[param.Key], vend.Inform[param.Key], revertValues))
		case INISectionPagecache:
			if revertValues {
				switch param.Key {
				case system.SysctlPagecacheLimitIgnoreDirty:
					pc.VMPagecacheLimitIgnoreDirty, _ = strconv.Atoi(vend.SysctlParams[param.Key])
				case "OVERRIDE_PAGECACHE_LIMIT_MB":
					pc.VMPagecacheLimitMB, _ = strconv.ParseUint(vend.SysctlParams[param.Key], 10, 64)
				}
			}
			errs = append(errs, SetPagecacheVal(param.Key, &pc))
		default:
			system.WarningLog("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	err = sap.PrintErrors(errs)
	return err
}

// SetValuesToApply fills the data structure for applying the changes
func (vend INISettings) SetValuesToApply(values []string) Note {
	vend.ValuesToApply = make(map[string]string)
	for _, v := range values {
		vend.ValuesToApply[v] = v
	}
	return vend
}

// getCounterPart gets the counterpart parameters of the vm.dirty parameters
func (vend INISettings) getCounterPart(key string, revert bool) (string, string) {
	// for the vm.dirty parameters take the counterpart
	// parameters into account (only during revert)
	// if vm.dirty_background_bytes is set to a value != 0,
	// vm.dirty_background_ratio is set to 0 and vice versa
	// if vm.dirty_bytes is set to a value != 0,
	// vm.dirty_ratio is set to 0 and vice versa
	rkey := key
	rval := vend.SysctlParams[key]
	cpart := "" //counterpart parameter
	switch key {
	case "vm.dirty_background_bytes":
		cpart = "vm.dirty_background_ratio"
	case "vm.dirty_bytes":
		cpart = "vm.dirty_ratio"
	case "vm.dirty_background_ratio":
		cpart = "vm.dirty_background_bytes"
	case "vm.dirty_ratio":
		cpart = "vm.dirty_bytes"
	}
	// in case of revert of a vm.dirty parameter
	// check, if the saved counterpart value is != 0
	// then revert this value
	if revert && cpart != "" && vend.SysctlParams[cpart] != "0" {
		rkey = cpart
		rval = vend.SysctlParams[cpart]
	}
	return rkey, rval
}

// setRevertParamValues sets the parameter values for revert
func (vend INISettings) setRevertParamValues(key string) (string, string) {
	// revert parameter value
	flstates := ""
	pvalue, pvendID := RevertParameter(key, vend.ID)
	if pvendID == "" {
		pvendID = vend.ID
	}
	if pvalue != "" {
		vend.SysctlParams[key] = pvalue
	}
	if key == "force_latency" {
		flstates, _ = RevertParameter("fl_states", vend.ID)
	}
	return pvendID, flstates
}

// createParamSavedStates creates the parameter saved state file
func (vend INISettings) createParamSavedStates(key, flstates string) {
	// do not write parameter values to the saved state file during
	// a pure 'verify' action
	if _, ok := vend.ValuesToApply["verify"]; !ok && vend.SysctlParams[key] != "" {
		start := vend.SysctlParams[key]
		if key == "UserTasksMax" {
			if system.SystemctlIsStarting() {
				start = system.GetBackupValue("/var/lib/saptune/working/.tmbackup")
			} else {
				start = system.GetTasksMax("0")
				system.WriteBackupValue(start, "/var/lib/saptune/working/.tmbackup")
			}
		}
		CreateParameterStartValues(key, start)
		if key == "force_latency" {
			CreateParameterStartValues("fl_states", flstates)
		}
	}
}

// addParamSavedStates adds values to the parameter saved state file
func (vend INISettings) addParamSavedStates(key string) {
	// do not write parameter values to the saved state file during
	// a pure 'verify' action
	if _, ok := vend.ValuesToApply["verify"]; !ok && vend.SysctlParams[key] != "" {
		AddParameterNoteValues(key, vend.SysctlParams[key], vend.ID)
	}
}

// handleID1805750 handles the special case of SAP Note 1805750
func (vend INISettings) handleID1805750(key, val string) (string, string) {
	// as note 1805750 does not set a limits domain, but
	// the customer should be able to set the correct
	// domain using an override file we need to rewrite
	// param.Key and param.Value to get a correct behaviour
	if len(vend.OverrideParams) != 0 && vend.ID == "1805750" {
		for owkey, owval := range vend.OverrideParams {
			if (isLimitSoft.MatchString(key) && isLimitSoft.MatchString(owkey)) || (isLimitHard.MatchString(key) && isLimitHard.MatchString(owkey)) {
				key = owkey
				val = owval
			}
		}
	}
	return key, val
}

// handleInitOverride handles the override parameter settings
func (vend INISettings) handleInitOverride(key, val, section string, op txtparser.Operator, over *txtparser.INIFile) (string, string, txtparser.Operator) {
	chkKey := key
	if section == "service" {
		cKey := strings.TrimSuffix(chkKey, ".service")
		if _, ok := over.KeyValue[section][cKey]; ok {
			chkKey = cKey
		}
	}
	if vend.ID == "1805750" {
		// as note 1805750 does not set a limits
		// domain, but the customer should be able to
		// set the correct domain using an override
		// file we need to rewrite param.Key and
		// param.Value to get a correct behaviour
		for owkey, owparam := range over.KeyValue[section] {
			if (isLimitSoft.MatchString(key) && isLimitSoft.MatchString(owkey)) || (isLimitHard.MatchString(key) && isLimitHard.MatchString(owkey)) {
				chkKey = owkey
				key = owkey
				val = owparam.Value
			}
		}
	}
	if over.KeyValue[section][chkKey].Value == "" && section != INISectionPagecache && (over.KeyValue[section][chkKey].Key != "" || (section == INISectionLimits && over.KeyValue[section][chkKey].Key == "")) {
		// disable parameter setting in override file
		vend.OverrideParams[chkKey] = "untouched"
	}
	if over.KeyValue[section][chkKey].Value != "" {
		vend.OverrideParams[chkKey] = over.KeyValue[section][chkKey].Value
		if over.KeyValue[section][chkKey].Operator != op {
			// operator from override file will
			// replace the operator from our note file
			op = over.KeyValue[section][chkKey].Operator
		}
	}
	return key, val, op
}

// printSchedInfo prints info about used block scheduler only during 'verify' to
// suppress double prints in case of 'apply'
func (vend INISettings) printSchedInfo(scheds string, blckOK map[string][]string) {
	if _, ok := vend.ValuesToApply["verify"]; ok && scheds != "" {
		if scheds == "untouched" {
			system.NoticeLog("Schedulers will be remain untouched!")
		} else {
			system.NoticeLog("Trying scheduler in this order: %s.", scheds)
			for b, s := range blckOK {
				system.NoticeLog("'%s' will be used as new scheduler for device '%s'.", b, strings.Join(s, " "))
			}
		}
	}
}

// useOverrides checks, if we should use the value from override file
func (vend INISettings) useOverrides(key, scheds, val string) (bool, string, string) {
	nxt := false
	if len(vend.OverrideParams[key]) != 0 {
		// use value from override file instead of the value
		// from the sap note (ConfFile)
		if vend.OverrideParams[key] == "untouched" {
			if system.IsSched.MatchString(key) {
				scheds = "untouched"
			}
			nxt = true
		} else {
			val = vend.OverrideParams[key]
		}
	}
	return nxt, scheds, val
}

// chkDoubles checks for double defined parameters
// till now for /sys parameter settings
// like KSM, THP and /sys/block/*/queue
func (vend INISettings) chkDoubles(key, info string) string {
	paramFiles := system.GetFiles(SaptuneParameterStateDir)

	syskey := key
	inf := ""
	searchParam, sect := system.GetSysSearchParam(syskey)
	sParam := strings.TrimPrefix(searchParam, "sys:")
	matchTxt := "[" + sect + "] '" + sParam

	if vend.SysctlParams[searchParam] != "" {
		// defined in same note definition file
		inf = matchTxt + "' of note " + vend.ID
	} else if _, exists := paramFiles[sParam]; exists {
		// defined in another note definition file
		inf = matchTxt + "' from the other applied notes"
	}
	if inf != "" {
		system.WarningLog("'%s' is defined twice, see section %s", syskey, inf)
		if info != "" {
			info = info + "§" + inf
		} else {
			info = inf
		}
	}
	return info
}
