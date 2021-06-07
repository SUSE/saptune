package note

import (
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"strings"
)

// section [sys]

// GetSysVal reads the sys system value
func GetSysVal(key string) (string, string) {
	info := ""
	keyFields := strings.Split(key, ":")
	if len(keyFields) > 1 {
		key = keyFields[1]
	}
	val, _ := system.GetSysString(key)
	if strings.ContainsAny(val, "[]") {
		val, _ = system.GetSysChoice(key)
	}
	if val == "" {
		val = "NA"
	}
	return val, info
}

// OptSysVal optimises a sys parameter value
func OptSysVal(operator txtparser.Operator, key, actval, cfgval string) string {
	syskey := key
	keyFields := strings.Split(key, ":")
	if len(keyFields) > 1 {
		syskey = keyFields[1]
	}
	val := OptSysctlVal(operator, syskey, actval, cfgval)
	return val
}

// SetSysVal applies the settings to the system
func SetSysVal(key, value string) error {
	syskey := key
	keyFields := strings.Split(key, ":")
	if len(keyFields) > 1 {
		syskey = keyFields[1]
	}
	err := system.SetSysString(syskey, value)
	return err
}
