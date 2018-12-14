// wrapper to cpupower command
package system

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	notSupported = "System does not support Intel's performance bias setting"
	sap_fl_file  = "/var/lib/saptune/saved_state/sap_force_latency"
	cpu_dir      = "/sys/devices/system/cpu"
	cpu_dir_sys  = "devices/system/cpu"
	cpupower_cmd = "/usr/bin/cpupower"
)

var isCpu = regexp.MustCompile(`^cpu\d+$`)
var isState = regexp.MustCompile(`^state\d+$`)

func GetPerfBias() string {
	isPBCpu := regexp.MustCompile(`analyzing CPU \d+`)
	isPBias := regexp.MustCompile(`perf-bias: \d+`)
	setAll := true
	str := ""
	oldpb := "99"
	cmdName := cpupower_cmd
	cmdArgs := []string{"-c", "all", "info", "-b"}

	if _, err := os.Stat(cmdName); os.IsNotExist(err) {
		log.Printf("command '%s' not found", cmdName)
		return "all:none"
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		log.Printf("There was an error running external command 'cpupower -c all info -b': %v, output: %s", err, cmdOut)
		return "all:none"
	}

	for k, line := range strings.Split(strings.TrimSpace(string(cmdOut)), "\n") {
		switch {
		case line == notSupported:
			//log.Print(notSupported)
			return "all:none"
		case isPBCpu.MatchString(line):
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
	if !SupportsPerfBias() {
		log.Print(notSupported)
		return nil
	}
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
		} else {
			cpu = fields[0]
		}
		cmd := exec.Command(cpupower_cmd, "-c", cpu, "set", "-b", fields[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to invoke external command 'cpupower -c %s set -b %s': %v, output: %s", cpu, fields[1], err, out)
		}
	}
	return nil
}

func SupportsPerfBias() bool {
	cmdName := cpupower_cmd
	cmdArgs := []string{"info", "-b"}

	if _, err := os.Stat(cmdName); os.IsNotExist(err) {
		log.Printf("command '%s' not found", cmdName)
		return false
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil || (err == nil && strings.Contains(string(cmdOut), notSupported)) {
		// does not support perf bias
		return false
	}
	return true
}

func GetGovernor() map[string]string {
	setAll := true
	oldgov := "99"
	gov := ""
	gGov := make(map[string]string)

	dirCont, err := ioutil.ReadDir(cpu_dir)
        if err != nil {
                return gGov
        }
        for _, entry := range dirCont {
		if isCpu.MatchString(entry.Name()) {
			gov, _ = GetSysString(path.Join(cpu_dir_sys, entry.Name(), "cpufreq", "scaling_governor"))
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
	tst := ""
	cmdName := cpupower_cmd

	if _, err := os.Stat(cmdName); os.IsNotExist(err) {
		log.Printf("command '%s' not found", cmdName)
		return nil
	}
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
			tst = cpu
		} else {
			cpu = fields[0]
			tst = "cpu0"
		}
		if !IsValidGovernor(tst, fields[1]) {
			log.Printf("'%s' is not a valid governor, skipping.", fields[1])
			continue
		}
		cmd := exec.Command(cpupower_cmd, "-c", cpu, "frequency-set", "-g", fields[1])
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to invoke external command 'cpupower -c %s frequency-set -g %s': %v, output: %s", cpu, fields[1], err, out)
		}
	}
	return nil
}

func IsValidGovernor(cpu, gov string) bool {
	val, err := ioutil.ReadFile(path.Join(cpu_dir, cpu, "/cpufreq/scaling_available_governors"))
	if err == nil && strings.Contains(string(val), gov) {
		return true
	}
	return false
}

func GetForceLatency() string {
	flval := ""
	supported := false

	// first check, if idle settings are supported by the system
	// on VMs it's not supported

	dirCont, err := ioutil.ReadDir(cpu_dir)
        if err != nil {
		return "all:none"
        }
        for _, entry := range dirCont {
		// cpu0 ... cpuXY
		if isCpu.MatchString(entry.Name()) {
			_, err := ioutil.ReadDir(path.Join(cpu_dir, entry.Name(), "cpuidle"))
			if err != nil {
				//log.Printf("SetForceLatency: idle settings not supported for '%s'", entry.Name())
				continue
			} else {
				supported = true
			}
		}
	}
	if !supported {
		return "all:none"
	}

	val, err := ioutil.ReadFile(sap_fl_file)
	if err != nil {
		// file 'sap_fl_file' not found, so no sap_force_latency set before
		// use /dev/cpu_dma_latency instead
		flval = GetdmaLatency()
	} else {
		flval = string(val)
	}
        return flval
}

func SetForceLatency(value string, revert bool) error {
	if value == "all:none" {
		return fmt.Errorf("latency settings not supported by the system")
	}

	supported := false
	flval, _ := strconv.Atoi(value) // decimal value for force latency

	dirCont, err := ioutil.ReadDir(cpu_dir)
        if err != nil {
		return fmt.Errorf("latency settings not supported by the system")
        }
        for _, entry := range dirCont {
		// cpu0 ... cpuXY
		if isCpu.MatchString(entry.Name()) {
			cpudirCont, err := ioutil.ReadDir(path.Join(cpu_dir, entry.Name(), "cpuidle"))
			if err != nil {
				log.Printf("SetForceLatency: idle settings not supported for '%s'", entry.Name())
				continue
			}
			supported = true
			for _, centry := range cpudirCont {
				// state0 ... stateXY
				if isState.MatchString(centry.Name()) {
					// read /sys/devices/system/cpu/cpu*/cpuidle/state*/latency
					lat, _ := GetSysInt(path.Join(cpu_dir_sys, entry.Name(), "cpuidle", centry.Name(), "latency"))
					// write /sys/devices/system/cpu/cpu*/cpuidle/state*/disable
					if revert {
						// revert
						if lat <= flval {
							err = SetSysString(path.Join(cpu_dir_sys, entry.Name(), "cpuidle", centry.Name(), "disable"), "0")
						}
					} else {
						// apply
						if lat > flval {
							err = SetSysString(path.Join(cpu_dir_sys, entry.Name(), "cpuidle", centry.Name(), "disable"), "1")
						}
					}
				}
			}
		}
	}

	// ANGI TODO - was, wenn force_latency in mehr als einer Note gesetzt wird, mit unterschiedlichen Werten?
	if revert {
		dmaLat := GetdmaLatency()
		// ANGI TODO - was, wenn der Wert von 'revert' nicht mit dem aktuellen Wert von /dev/cpu_dma_latency übereinstimmt
		if value != dmaLat {
			log.Printf("SetForceLatency: reverted value for force latency (%s) differs from /dev/cpu_dma_latency setting (%s).", value, dmaLat)
		}

		// revert
		if _, err := os.Stat(sap_fl_file); err == nil {
			os.Remove(sap_fl_file)
		}
	} else {
		// apply
		if supported {
			ioutil.WriteFile(sap_fl_file, []byte(value), 0644)
		}
	}

	return err
}

func GetdmaLatency() string {
	latency := make([]byte, 4)
	dmaLatency, err := os.OpenFile("/dev/cpu_dma_latency", os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("GetForceLatency: failed to open cpu_dma_latency - %v", err)
	}
	_, err = dmaLatency.Read(latency)
	if err != nil {
		log.Printf("GetForceLatency: reading from '/dev/cpu_dma_latency' failed:", err)
	}
	// Close the file handle after the latency value is no longer maintained
	defer dmaLatency.Close()

	ret := fmt.Sprintf("%v", binary.LittleEndian.Uint32(latency))
	return ret
}
