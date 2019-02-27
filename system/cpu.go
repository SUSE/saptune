// wrapper to cpupower command
package system

import (
	"encoding/binary"
	"encoding/json"
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

//constant definition
const (
	notSupported = "System does not support Intel's performance bias setting"
	forceLatDir  = "/var/lib/saptune/force_latency"
	cpu_dir      = "/sys/devices/system/cpu"
	cpu_dir_sys  = "devices/system/cpu"
	cpupower_cmd = "/usr/bin/cpupower"
)

var isCpu = regexp.MustCompile(`^cpu\d+$`)
var isState = regexp.MustCompile(`^state\d+$`)
var forceLatFile = path.Join(forceLatDir, "/sap_force_latency")

// GetPerfBias retrieve CPU performance configuration from the system
func GetPerfBias() string {
	isPBCpu := regexp.MustCompile(`analyzing CPU \d+`)
	isPBias := regexp.MustCompile(`perf-bias: \d+`)
	setAll := true
	str := ""
	oldpb := "99"
	cmdName := cpupower_cmd
	cmdArgs := []string{"-c", "all", "info", "-b"}

	if !CmdIsAvailable(cmdName) {
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

// SetPerfBias set CPU performance configuration to the system using 'cpupower' command
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

// SupportsPerfBias check, if the system will support CPU performance settings
func SupportsPerfBias() bool {
	cmdName := cpupower_cmd
	cmdArgs := []string{"info", "-b"}

	if !CmdIsAvailable(cmdName) {
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

// GetGovernor retrieve performance configuration regarding to cpu frequency
// from the system
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

// SetGovernor set performance configuration regarding to cpu frequency
// to the system using 'cpupower' command
func SetGovernor(value string) error {
	//cmd := exec.Command("cpupower", "-c", "all", "frequency-set", "-g", value)
	cpu := ""
	tst := ""
	cmdName := cpupower_cmd

	if !CmdIsAvailable(cmdName) {
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

// IsValidGovernor check, if the system will support CPU frequency settings
func IsValidGovernor(cpu, gov string) bool {
	val, err := ioutil.ReadFile(path.Join(cpu_dir, cpu, "/cpufreq/scaling_available_governors"))
	if err == nil && strings.Contains(string(val), gov) {
		return true
	}
	return false
}

// GetForceLatency retrieve CPU latency configuration from the system
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

	val, err := ioutil.ReadFile(forceLatFile)
	if err != nil {
		// file 'sap_force_latency' not found, so no sap_force_latency set before
		// use /dev/cpu_dma_latency instead
		flval = GetdmaLatency()
	} else {
		flval = string(val)
	}
	return flval
}

// SetForceLatency set CPU latency configuration to the system
func SetForceLatency(noteId, value string, revert bool) error {
	oldLat := ""
	oldLatStates := make(map[string]string)
	savedStateFile := noteId + "_saved_fl_states"

	if value == "all:none" {
		return fmt.Errorf("latency settings not supported by the system")
	}

	supported := false
	flval, _ := strconv.Atoi(value) // decimal value for force latency
	if revert {
		// read saved latency states from /var/lib/saptune/force_latency/<NoteId>_saved_fl_states
		content, err := ioutil.ReadFile(path.Join(forceLatDir, savedStateFile))
		if err == nil {
			json.Unmarshal(content, &oldLatStates)
		}
	}

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
						if oldLatStates[entry.Name()] != "" {
							oldLatFields := strings.Split(oldLatStates[entry.Name()], ":")
							if oldLatFields[0] == centry.Name() {
								oldLat = oldLatFields[1]
							}
						}
						if oldLat != "" {
							err = SetSysString(path.Join(cpu_dir_sys, entry.Name(), "cpuidle", centry.Name(), "disable"), oldLat)
							// clear old latency value for next cpu/state cycle
							oldLat = ""
						}
					} else {
						// apply
						if lat > flval {
							// save old latency states for 'revert'
							// oldLatStates["cpu1"] = "state0:0"
							oldLat, _ = GetSysString(path.Join(cpu_dir_sys, entry.Name(), "cpuidle", centry.Name(), "disable"))
							oldLatStates[entry.Name()] = centry.Name() + ":" + oldLat
							// set new latency states
							err = SetSysString(path.Join(cpu_dir_sys, entry.Name(), "cpuidle", centry.Name(), "disable"), "1")
						}
					}
				}
			}
		}
	}

	if revert {
		// revert
		dmaLat := GetdmaLatency()
		if value != dmaLat {
			log.Printf("SetForceLatency: reverted value for force latency (%s) differs from /dev/cpu_dma_latency setting (%s).", value, dmaLat)
		}
		ioutil.WriteFile(forceLatFile, []byte(value), 0644)

		// remove saved latency states
		_, err := os.Stat(path.Join(forceLatDir, savedStateFile))
		if os.IsNotExist(err) {
			err = nil
		} else if err == nil {
			os.Remove(path.Join(forceLatDir, savedStateFile))
		}
	} else {
		// apply
		if supported {
			if err = os.MkdirAll(forceLatDir, 0755); err != nil {
				return err
			}
			ioutil.WriteFile(forceLatFile, []byte(value), 0644)
			content, err := json.Marshal(oldLatStates)
			if err != nil {
				return err
			}
			ioutil.WriteFile(path.Join(forceLatDir, savedStateFile), content, 0644)
		}
	}

	return err
}

// GetdmaLatency retrieve DMA latency configuration from the system
func GetdmaLatency() string {
	latency := make([]byte, 4)
	dmaLatency, err := os.OpenFile("/dev/cpu_dma_latency", os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("GetForceLatency: failed to open cpu_dma_latency - %v", err)
	}
	_, err = dmaLatency.Read(latency)
	if err != nil {
		log.Printf("GetForceLatency: reading from '/dev/cpu_dma_latency' failed: %v", err)
	}
	// Close the file handle after the latency value is no longer maintained
	defer dmaLatency.Close()

	ret := fmt.Sprintf("%v", binary.LittleEndian.Uint32(latency))
	return ret
}
