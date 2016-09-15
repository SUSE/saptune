// Manipulate /sys/ switches.
package system

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

const (
	SYS_THP = "kernel/mm/transparent_hugepage/enabled"
	SYS_KSM = "kernel/mm/ksm/run"
)

// Read a /sys/ key and return the string value. Panic on error.
func GetSysString(parameter string) string {
	val, err := ioutil.ReadFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		panic(fmt.Errorf("failed to read sys string key '%s': %v", parameter, err))
	}
	return strings.TrimSpace(string(val))
}

// Read a /sys/ key that comes with current value and alternative choices, return the current choice or empty string. Panic on error
func GetSysChoice(parameter string) string {
	val, err := ioutil.ReadFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		panic(fmt.Errorf("failed to read sys key of choices '%s': %v", parameter, err))
	}
	// Split up the choices
	allChoices := consecutiveSpaces.Split(string(val), -1)
	for _, choice := range allChoices {
		if len(choice) > 2 && choice[0] == '[' && choice[len(choice)-1] == ']' {
			return choice[1 : len(choice)-1]
		}
	}
	return ""
}

// Read an integer /sys/ key. Panic on error.
func GetSysInt(parameter string) int {
	value, err := strconv.Atoi(GetSysString(parameter))
	if err != nil {
		panic(fmt.Errorf("failed to read integer sys key '%s': %v", parameter, err))
	}
	return value
}

// Write a string /sys/ value. Panic on error.
func SetSysString(parameter, value string) {
	if err := ioutil.WriteFile(path.Join("/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 644); err != nil {
		panic(fmt.Errorf("failed to set sys key '%s' to string '%s': %v", parameter, value, err))
	}
}

// Write an integer /sys/ value. Panic on error.
func SetSysInt(parameter string, value int) {
	SetSysString(parameter, strconv.Itoa(value))
}
