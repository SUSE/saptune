// wrapper to cpupower command
package system

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
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
			return fmt.Errorf("failed to invoke external command 'cpupower -c %s set -b %s': %v, output: %s", cpu, fields[1], err, out)
		}
	}
	return nil
}

func GetGovernor() map[string]string {
	setAll := true
	oldgov := "99"
	isCpu := regexp.MustCompile(`^cpu\d+$`)
	gov := ""
	gGov := make(map[string]string)

	dirCont, err := ioutil.ReadDir("/sys/devices/system/cpu")
        if err != nil {
                return gGov
        }
        for _, entry := range dirCont {
		if isCpu.MatchString(entry.Name()) {
			gov, _ = GetSysString(path.Join("devices", "system", "cpu", entry.Name(), "cpufreq", "scaling_governor"))
			if gov == "" {
				gov = "none"
			}
			if oldgov == "99" {
				// starting point
				oldgov = gov
			}
			if oldgov != gov {
				setAll = false
			}
			gGov[entry.Name()] = gov
		}
	}
	if setAll {
		gGov = make(map[string]string)
		gGov["all"] = oldgov
	}
	return gGov
}

func SetGovernor(value string) error {
	//cmd := exec.Command("cpupower", "-c", "all", "frequency-set", "-g", value)
	cpu := ""
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if fields[1] == "none" {
			return nil
		}
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
		} else {
			cpu = fields[0]
		}
		cmd := exec.Command("cpupower", "-c", cpu, "frequency-set", "-g", fields[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to invoke external command 'cpupower -c %s frequency-set -g %s': %v, output: %s", cpu, fields[1], err, out)
		}
	}
	return nil
}

func IsValidGovernor(cpu, gov string) bool {
	val, err := ioutil.ReadFile(path.Join("/sys/devices/system/cpu/", cpu, "/cpufreq/scaling_available_governors"))
	if err == nil && strings.Contains(string(val), gov) {
		return true
	}
	return false
}
