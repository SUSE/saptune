package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"strconv"
)

const (
	INISectionSysctl = "sysctl"
)

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
		case txtparser.OperatorMoreThan:
			iCurrentValue = iExpectedValue + 1
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
}

func (vend INISettings) Name() string {
	return vend.DescriptiveName
}

func (vend INISettings) Initialise() (Note, error) {
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}
	// Read current parameter values
	vend.SysctlParams = make(map[string]string)
	for _, param := range ini.AllValues {
		switch param.Section {
		case INISectionSysctl:
			currValue := system.GetSysctlString(param.Key, "")
			if currValue == "" {
				return vend, fmt.Errorf("INISettings %s: cannot find parameter \"%s\" in system", vend.ID, param)
			}
			vend.SysctlParams[param.Key] = currValue
		}
	}
	return vend, nil
}

func (vend INISettings) Optimise() (Note, error) {
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}
	for _, param := range ini.AllValues {
		// Compare current values against INI's definition
		switch param.Section {
		case INISectionSysctl:
			optimisedValue, err := CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			if err != nil {
				return vend, err
			}
			vend.SysctlParams[param.Key] = optimisedValue
		}
	}
	return vend, nil
}

func (vend INISettings) Apply() error {
	// Apply sysctl parameters
	for key, value := range vend.SysctlParams {
		system.SetSysctlString(key, value)
	}
	return nil
}
