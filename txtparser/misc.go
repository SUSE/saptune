package txtparser

import (
	"github.com/SUSE/saptune/system"
	"strconv"
)

// currently not used functions, for future use

// CalculateOptimumValue calculates optimum parameter value given the current
// value, comparison operator, and expected value. Return optimised value.
func CalculateOptimumValue(operator Operator, currentValue string, expectedValue string) (string, error) {
	if operator == OperatorEqual {
		return expectedValue, nil
	}
	// Numeric comparisons
	var iCurrentValue int64
	iExpectedValue, err := strconv.ParseInt(expectedValue, 10, 64)
	if err != nil {
		return "", system.ErrorLog("%+v - Expected value \"%s\" should be but is not an integer", err, expectedValue)
	}
	if currentValue == "" {
		switch operator {
		case OperatorLessThan:
			iCurrentValue = iExpectedValue - 1
		case OperatorLessThanEqual:
			iCurrentValue = iExpectedValue
		case OperatorMoreThan:
			iCurrentValue = iExpectedValue + 1
		case OperatorMoreThanEqual:
			iCurrentValue = iExpectedValue
		}
	} else {
		iCurrentValue, err = strconv.ParseInt(currentValue, 10, 64)
		if err != nil {
			return "", system.ErrorLog("%+v - Current value \"%s\" should be but is not an integer", err, currentValue)
		}
		switch operator {
		case OperatorLessThan:
			if iCurrentValue >= iExpectedValue {
				iCurrentValue = iExpectedValue - 1
			}
		case OperatorMoreThan:
			if iCurrentValue <= iExpectedValue {
				iCurrentValue = iExpectedValue + 1
			}
		case OperatorLessThanEqual:
			if iCurrentValue >= iExpectedValue {
				iCurrentValue = iExpectedValue
			}
		case OperatorMoreThanEqual:
			if iCurrentValue <= iExpectedValue {
				iCurrentValue = iExpectedValue
			}
		}
	}
	return strconv.FormatInt(iCurrentValue, 10), nil
}
