package note

import (
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"strings"
)

// section [sysctl]

// OptSysctlVal optimises a sysctl parameter value
// use exactly the value from the config file. No calculation any more
func OptSysctlVal(operator txtparser.Operator, key, actval, cfgval string) string {
	if actval == "PNA" || actval == "" {
		// sysctl parameter not available in system
		// or system value is 'empty'
		return cfgval
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
			optimisedValue, err := txtparser.CalculateOptimumValue(operator, fieldC, fieldE)
			//optimisedValue, err := txtparser.CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			if err != nil {
				return ""
			}
			allFieldsS = allFieldsS + optimisedValue + "\t"
		*/
		allFieldsS = allFieldsS + fieldE + "\t"
	}

	return strings.TrimSpace(allFieldsS)
}
