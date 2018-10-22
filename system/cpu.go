// wrapper to cpupower command
package system

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const	notSupported = "System does not support Intel's performance bias setting"

func GetPerfBias() string {

	isCpu := regexp.MustCompile(`analyzing CPU \d+`)
	isPBias := regexp.MustCompile(`perf-bias: \d+`)
	setAll := true
	str := ""
	oldpb := "99"
	cmdName := "/usr/bin/cpupower"
	cmdArgs := []string{"-c", "all", "info", "-b"}

	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("There was an error running external command 'cpupower -c all info -b': %v, output: %s", err, cmdOut)
		return ""
	}

	for k, line := range strings.Split(strings.TrimSpace(string(cmdOut)), "\n") {
		switch {
		case line == notSupported:
			//log.Print(notSupported)
			return "all:none"
		case isCpu.MatchString(line):
			str = str + fmt.Sprintf("cpu%d", k/2)
		case isPBias.MatchString(line):
			pb := strings.Split(line, ":")
			if oldpb == "99" {
				oldpb = strings.TrimSpace(pb[1])
			}
			if oldpb != strings.TrimSpace(pb[1]) {
				setAll = false
			}
			str = str + fmt.Sprintf(":%s ", strings.TrimSpace(pb[1]))
		}
	}
	if setAll {
		str = "all:" + oldpb
	}
	return strings.TrimSpace(str)
}

func SetPerfBias(value string) error {
	//cmd := exec.Command("cpupower", "-c", "all", "set", "-b", value)
	cpu := ""
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if fields[1] == "none" {
			log.Print(notSupported)
			return nil
		}
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
		} else {
			cpu = fields[0]
		}
		cmd := exec.Command("cpupower", "-c", cpu, "set", "-b", fields[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to invoke external command 'cpupower -c all set -b %s': %v, output: %s", value, err, out)
		}
	}
	return nil
}
