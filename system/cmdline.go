package system

// Gather information about kernel cmdline

import (
	"io/ioutil"
	"log"
	"strings"
)

// ParseCmdline parse /proc/cmdline into key(string) - value(string) pairs.
// return value for given boot option or 'NA', if not available
func ParseCmdline(fileName, option string) string {
	opt := "NA"
	cmdLine, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("ParseCmdline: failed to read  %s: %v", fileName, err)
		return opt
	}
	for _, param := range strings.Fields(string(cmdLine)) {
		fields := strings.Split(param, "=")
		if fields[0] == option {
			if len(fields) > 1 {
				opt = fields[1]
			} else {
				opt = option
			}
		}
	}
	return opt
}
