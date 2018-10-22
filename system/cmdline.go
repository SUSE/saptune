// Gather information about kernel cmdline
package system

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Parse /proc/cmdline into key(string) - value(string) pairs.
// return value for given boot option or 'NA', if not available
func ParseCmdline(option string) string {
	opt := "NA"
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		panic(fmt.Errorf("failed to read /proc/cmdline: %v", err))
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
