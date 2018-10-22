package note

import (
	"fmt"
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"log"
	"path"
	"strconv"
)

const OverrideTunigSheets  = "/etc/saptune/override/"
var pc = LinuxPagingImprovements{}

// Tuning options composed by a third party vendor.

// Calculate optimum parameter value given the current value, comparison operator, and expected value. Return optimised value.
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

// Tuning options composed by a third party vendor.
type INISettings struct {
	ConfFilePath    string            // Full path to the 3rd party vendor's tuning configuration file
	ID              string            // ID portion of the tuning configuration
	DescriptiveName string            // Descriptive name portion of the tuning configuration
	SysctlParams    map[string]string // Sysctl parameter values from the computer system
	ValuesToApply   map[string]string // values to apply
	OverrideParams  map[string]string // parameter values from the override file
}

func (vend INISettings) Name() string {
	if len(vend.DescriptiveName) == 0 {
		vend.DescriptiveName = txtparser.GetINIFileDescriptiveName(vend.ConfFilePath)
	}
	return vend.DescriptiveName
}

func (vend INISettings) Initialise() (Note, error) {
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}

	// looking for override file
	override := false
	ow, err := txtparser.ParseINIFile(path.Join(OverrideTunigSheets, vend.ID), false)
	if err == nil {
		override = true
	}
	// Read current parameter values
	vend.SysctlParams = make(map[string]string)
	vend.OverrideParams = make(map[string]string)
	for _, param := range ini.AllValues {
		if override {
			if len(ow.KeyValue[param.Section]) != 0 && ow.KeyValue[param.Section][param.Key].Value != "" {
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
			vend.SysctlParams[param.Key] = GetVmVal(param.Key)
		case INISectionBlock:
			vend.SysctlParams[param.Key], _ = GetBlkVal(param.Key)
		case INISectionLimits:
			vend.SysctlParams[param.Key], _ = GetLimitsVal(param.Key, ini.KeyValue["limits"]["LIMIT_ITEM"].Value, ini.KeyValue["limits"]["LIMIT_DOMAIN"].Value)
		case INISectionUuidd:
			vend.SysctlParams[param.Key] = GetUuiddVal()
		case INISectionLogin:
			vend.SysctlParams[param.Key], _ = GetLoginVal(param.Key)
		case INISectionMEM:
			vend.SysctlParams[param.Key] = GetMemVal(param.Key)
		case INISectionCPU:
			vend.SysctlParams[param.Key] = GetCpuVal(param.Key)
		case INISectionReminder:
			vend.SysctlParams[param.Key] = param.Value
		case INISectionPagecache:
			vend.SysctlParams[param.Key] = GetPagecacheVal(param.Key, &pc)
		default:
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	return vend, nil
}

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
			param.Value = vend.OverrideParams[param.Key]
		}
		switch param.Section {
		case INISectionSysctl:
			//optimisedValue, err := CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			//vend.SysctlParams[param.Key] = optimisedValue
			vend.SysctlParams[param.Key] = OptSysctlVal(param.Operator, param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionVM:
			vend.SysctlParams[param.Key] = OptVmVal(param.Key, param.Value)
		case INISectionBlock:
			vend.SysctlParams[param.Key] = OptBlkVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionLimits:
			vend.SysctlParams[param.Key] = OptLimitsVal(param.Key, vend.SysctlParams[param.Key], param.Value, ini.KeyValue["limits"]["LIMIT_ITEM"].Value, ini.KeyValue["limits"]["LIMIT_DOMAIN"].Value)
		case INISectionUuidd:
			vend.SysctlParams[param.Key] = OptUuiddVal(param.Value)
		case INISectionLogin:
			vend.SysctlParams[param.Key] = OptLoginVal(param.Value)
		case INISectionMEM:
			vend.SysctlParams[param.Key] = OptMemVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionCPU:
			vend.SysctlParams[param.Key] = OptCpuVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionReminder:
			vend.SysctlParams[param.Key] = param.Value
		case INISectionPagecache:
			vend.SysctlParams[param.Key] = OptPagecacheVal(param.Key, param.Value, &pc, ini.KeyValue)
		default:
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	return vend, nil
}

func (vend INISettings) Apply() error {
	errs := make([]error, 0, 0)
	revertValues := false
	if len(vend.ValuesToApply) == 0 {
		revertValues = true
	}
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return err
	}
	//for key, value := range vend.SysctlParams {
	for _, param := range ini.AllValues {
		if _, ok := vend.ValuesToApply[param.Key]; !ok && !revertValues {
			continue
		}
		switch param.Section {
		case INISectionSysctl:
			// Apply sysctl parameters
			errs = append(errs, system.SetSysctlString(param.Key, vend.SysctlParams[param.Key]))
		case INISectionVM:
			errs = append(errs, SetVmVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionBlock:
			errs = append(errs, SetBlkVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionLimits:
			errs = append(errs, SetLimitsVal(param.Key, vend.SysctlParams[param.Key], ini.KeyValue["limits"]["LIMIT_ITEM"].Value))
		case INISectionUuidd:
			errs = append(errs, SetUuiddVal(vend.SysctlParams[param.Key]))
		case INISectionLogin:
			errs = append(errs, SetLoginVal(param.Key, vend.SysctlParams[param.Key], revertValues))
		case INISectionMEM:
			errs = append(errs, SetMemVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionCPU:
			errs = append(errs, SetCpuVal(param.Key, vend.SysctlParams[param.Key], revertValues))
		case INISectionReminder:
			//nothing to do here
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

func (vend INISettings) SetValuesToApply(values []string) (Note) {
	vend.ValuesToApply = make(map[string]string)
	for _, v := range values {
		vend.ValuesToApply[v] = v
	}
	return vend
}
