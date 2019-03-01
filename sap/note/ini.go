package note

import (
	"fmt"
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"log"
	"path"
	"strconv"
	"strings"
)

// OverrideTuningSheets defines saptunes override directory
const OverrideTuningSheets = "/etc/saptune/override/"

var pc = LinuxPagingImprovements{}

// Tuning options composed by a third party vendor.

// CalculateOptimumValue calculates optimum parameter value given the current
// value, comparison operator, and expected value. Return optimised value.
func CalculateOptimumValue(operator txtparser.Operator, currentValue string, expectedValue string) (string, error) {
	if operator == txtparser.OperatorEqual {
		return expectedValue, nil
	}
	// Numeric comparisons
	var iCurrentValue int64
	iExpectedValue, err := strconv.ParseInt(expectedValue, 10, 64)
	if err != nil {
		return "", fmt.Errorf("Expected value \"%s\" should be but is not an integer", expectedValue)
	}
	if currentValue == "" {
		switch operator {
		case txtparser.OperatorLessThan:
			iCurrentValue = iExpectedValue - 1
		case txtparser.OperatorLessThanEqual:
			iCurrentValue = iExpectedValue
		case txtparser.OperatorMoreThan:
			iCurrentValue = iExpectedValue + 1
		case txtparser.OperatorMoreThanEqual:
			iCurrentValue = iExpectedValue
		}
	} else {
		iCurrentValue, err = strconv.ParseInt(currentValue, 10, 64)
		if err != nil {
			return "", fmt.Errorf("Current value \"%s\" should be but is not an integer", currentValue)
		}
		switch operator {
		case txtparser.OperatorLessThan:
			if iCurrentValue >= iExpectedValue {
				iCurrentValue = iExpectedValue - 1
			}
		case txtparser.OperatorMoreThan:
			if iCurrentValue <= iExpectedValue {
				iCurrentValue = iExpectedValue + 1
			}
		case txtparser.OperatorLessThanEqual:
			if iCurrentValue >= iExpectedValue {
				iCurrentValue = iExpectedValue
			}
		case txtparser.OperatorMoreThanEqual:
			if iCurrentValue <= iExpectedValue {
				iCurrentValue = iExpectedValue
			}
		}
	}
	return strconv.FormatInt(iCurrentValue, 10), nil
}

// INISettings defines tuning options composed by a third party vendor.
type INISettings struct {
	ConfFilePath    string            // Full path to the 3rd party vendor's tuning configuration file
	ID              string            // ID portion of the tuning configuration
	DescriptiveName string            // Descriptive name portion of the tuning configuration
	SysctlParams    map[string]string // Sysctl parameter values from the computer system
	ValuesToApply   map[string]string // values to apply
	OverrideParams  map[string]string // parameter values from the override file
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
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}

	// looking for override file
	override := false
	ow, err := txtparser.ParseINIFile(path.Join(OverrideTuningSheets, vend.ID), false)
	if err == nil {
		override = true
	}
	// Read current parameter values
	vend.SysctlParams = make(map[string]string)
	vend.OverrideParams = make(map[string]string)
	for _, param := range ini.AllValues {
		if override && len(ow.KeyValue[param.Section]) != 0 {
			if ow.KeyValue[param.Section][param.Key].Value == "" && ow.KeyValue[param.Section][param.Key].Key != "" {
				// disable parameter setting in override file
				vend.OverrideParams[param.Key] = "untouched"
			}
			if ow.KeyValue[param.Section][param.Key].Value != "" {
				vend.OverrideParams[param.Key] = ow.KeyValue[param.Section][param.Key].Value
				if ow.KeyValue[param.Section][param.Key].Operator != param.Operator {
					// operator from override file will
					// replace the operator from our note file
					param.Operator = ow.KeyValue[param.Section][param.Key].Operator
				}
			}
		}

		switch param.Section {
		case INISectionSysctl:
			vend.SysctlParams[param.Key], _ = system.GetSysctlString(param.Key)
		case INISectionVM:
			vend.SysctlParams[param.Key] = GetVMVal(param.Key)
		case INISectionBlock:
			vend.SysctlParams[param.Key], _ = GetBlkVal(param.Key)
		case INISectionLimits:
			vend.SysctlParams[param.Key], _ = GetLimitsVal(param.Key, ini.KeyValue["limits"]["LIMIT_ITEM"].Value, ini.KeyValue["limits"]["LIMIT_DOMAIN"].Value)
		case INISectionUuidd:
			vend.SysctlParams[param.Key] = GetUuiddVal()
		case INISectionService:
			vend.SysctlParams[param.Key] = GetServiceVal(param.Key)
		case INISectionLogin:
			vend.SysctlParams[param.Key], _ = GetLoginVal(param.Key)
		case INISectionMEM:
			vend.SysctlParams[param.Key] = GetMemVal(param.Key)
		case INISectionCPU:
			vend.SysctlParams[param.Key] = GetCPUVal(param.Key)
		case INISectionRpm:
			vend.SysctlParams[param.Key] = GetRpmVal(param.Key)
			continue
		case INISectionGrub:
			vend.SysctlParams[param.Key] = GetGrubVal(param.Key)
			continue
		case INISectionReminder:
			vend.SysctlParams[param.Key] = param.Value
			continue
		case INISectionPagecache:
			// page cache is special, has it's own config file
			// so adjust path to pagecache config file, if needed
			pcPrefix := strings.Split(vend.ConfFilePath, "/usr/share/saptune/notes/1557506")
			if len(pcPrefix) != 0 {
				pc = LinuxPagingImprovements{SysconfigPrefix: strings.Join(pcPrefix, "")}
			}
			vend.SysctlParams[param.Key] = GetPagecacheVal(param.Key, &pc)
		default:
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
		// do not write parameter values to the saved state file during
		// a pure 'verify' action
		if _, ok := vend.ValuesToApply["verify"]; !ok && vend.SysctlParams[param.Key] != "" {
			CreateParameterStartValues(param.Key, vend.SysctlParams[param.Key])
		}
	}
	return vend, nil
}

// Optimise gets the expected parameter values from the configuration
func (vend INISettings) Optimise() (Note, error) {
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}

	for _, param := range ini.AllValues {
		// Compare current values against INI's definition
		if len(vend.OverrideParams[param.Key]) != 0 {
			// use value from override file instead of the value
			// from the sap note (ConfFile)
			if vend.OverrideParams[param.Key] == "untouched" {
				continue
			}
			param.Value = vend.OverrideParams[param.Key]
		}
		switch param.Section {
		case INISectionSysctl:
			//optimisedValue, err := CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			//vend.SysctlParams[param.Key] = optimisedValue
			vend.SysctlParams[param.Key] = OptSysctlVal(param.Operator, param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionVM:
			vend.SysctlParams[param.Key] = OptVMVal(param.Key, param.Value)
		case INISectionBlock:
			vend.SysctlParams[param.Key] = OptBlkVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionLimits:
			vend.SysctlParams[param.Key] = OptLimitsVal(param.Key, vend.SysctlParams[param.Key], param.Value, ini.KeyValue["limits"]["LIMIT_ITEM"].Value, ini.KeyValue["limits"]["LIMIT_DOMAIN"].Value)
		case INISectionUuidd:
			vend.SysctlParams[param.Key] = OptUuiddVal(param.Value)
		case INISectionService:
			vend.SysctlParams[param.Key] = OptServiceVal(param.Key, param.Value)
		case INISectionLogin:
			vend.SysctlParams[param.Key] = OptLoginVal(param.Value)
		case INISectionMEM:
			vend.SysctlParams[param.Key] = OptMemVal(param.Key, vend.SysctlParams[param.Key], param.Value, ini.KeyValue["mem"]["ShmFileSystemSizeMB"].Value, ini.KeyValue["mem"]["VSZ_TMPFS_PERCENT"].Value)
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
		case INISectionPagecache:
			//vend.SysctlParams[param.Key] = OptPagecacheVal(param.Key, param.Value, &pc, ini.KeyValue)
			vend.SysctlParams[param.Key] = OptPagecacheVal(param.Key, param.Value, &pc)
		default:
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
		// do not write parameter values to the saved state file during
		// a pure 'verify' action
		if _, ok := vend.ValuesToApply["verify"]; !ok && vend.SysctlParams[param.Key] != "" {
			AddParameterNoteValues(param.Key, vend.SysctlParams[param.Key], vend.ID)
		}
	}
	return vend, nil
}

// Apply sets the new parameter values in the system or
// revert the system to the former parameter values
func (vend INISettings) Apply() error {
	errs := make([]error, 0, 0)
	revertValues := false

	if len(vend.ValuesToApply) == 0 {
		// nothing to apply
		return nil
	}
	if _, ok := vend.ValuesToApply["revert"]; ok {
		revertValues = true
	}
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return err
	}

	//for key, value := range vend.SysctlParams {
	for _, param := range ini.AllValues {
		switch param.Section {
		case INISectionRpm, INISectionGrub, INISectionReminder:
			// These parameters are only checked, but not applied.
			// So nothing to do during apply and no need for revert
			continue
		}

		if _, ok := vend.ValuesToApply[param.Key]; !ok && !revertValues {
			continue
		}

		if revertValues && vend.SysctlParams[param.Key] != "" {
			// revert parameter value
			pvalue := RevertParameter(param.Key, vend.ID)
			if pvalue != "" {
				vend.SysctlParams[param.Key] = pvalue
			}
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
			cpart := "" //counterpart parameter
			switch param.Key {
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
			if revertValues && cpart != "" && vend.SysctlParams[cpart] != "0" {
				errs = append(errs, system.SetSysctlString(cpart, vend.SysctlParams[cpart]))
			} else {
				errs = append(errs, system.SetSysctlString(param.Key, vend.SysctlParams[param.Key]))
			}
		case INISectionVM:
			errs = append(errs, SetVMVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionBlock:
			errs = append(errs, SetBlkVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionLimits:
			errs = append(errs, SetLimitsVal(param.Key, vend.SysctlParams[param.Key], ini.KeyValue["limits"]["LIMIT_ITEM"].Value))
		case INISectionUuidd:
			errs = append(errs, SetUuiddVal(vend.SysctlParams[param.Key]))
		case INISectionService:
			errs = append(errs, SetServiceVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionLogin:
			errs = append(errs, SetLoginVal(param.Key, vend.SysctlParams[param.Key], revertValues))
		case INISectionMEM:
			errs = append(errs, SetMemVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionCPU:
			errs = append(errs, SetCPUVal(param.Key, vend.SysctlParams[param.Key], revertValues, vend.ID))
		case INISectionPagecache:
			errs = append(errs, SetPagecacheVal(param.Key, &pc))
		default:
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
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
