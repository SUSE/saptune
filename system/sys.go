package system

// Manipulate /sys/ switches.

import (
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

// GetSysString read a /sys/ key and return the string value.
func GetSysString(parameter string) (string, error) {
	val, err := ioutil.ReadFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		WarningLog("failed to read sys string key '%s': %v", parameter, err)
		return "", err
	}
	return strings.TrimSpace(string(val)), nil
}

// GetSysChoice read a /sys/ key that comes with current value and alternative
// choices, return the current choice or empty string.
func GetSysChoice(parameter string) (string, error) {
	val, err := ioutil.ReadFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		WarningLog("failed to read sys key of choices '%s': %v", parameter, err)
		return "", err
	}
	// Split up the choices
	allChoices := consecutiveSpaces.Split(string(val), -1)
	for _, choice := range allChoices {
		if len(choice) > 2 && choice[0] == '[' && choice[len(choice)-1] == ']' {
			return choice[1 : len(choice)-1], nil
		}
	}
	return "", nil
}

// GetSysInt read an integer /sys/ key.
func GetSysInt(parameter string) (int, error) {
	value, err := GetSysString(parameter)
	if err != nil {
		WarningLog("failed to read integer sys key '%s': %v", parameter, err)
		return 0, err
	}
	return strconv.Atoi(value)
}

// SetSysString write a string /sys/ value.
func SetSysString(parameter, value string) error {
	if err := ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 0644); err != nil {
		WarningLog("failed to set sys key '%s' to string '%s': %v", parameter, value, err)
		return err
	}
	return nil
}

// SetSysInt write an integer /sys/ value.
func SetSysInt(parameter string, value int) error {
	return SetSysString(parameter, strconv.Itoa(value))
}

// TestSysString Test writing a string /sys/ value.
func TestSysString(parameter, value string) error {
	save, err := GetSysString(parameter)
	if err != nil {
		WarningLog("failed to get sys key '%s': %v", parameter, err)
		return err
	}
	if err = ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 0644); err == nil {
		// set key back to previous value, because this was only a test
		err = ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(save), 0644)
	}
	return err
}
