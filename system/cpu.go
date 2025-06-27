package system

// wrapper to cpupower command

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// constant definition
const (
	notSupportedX86 = "System does not support Intel's performance bias setting"
	notSupportedIBM = "Subcommand not supported on POWER."
	cpuDirSys       = "devices/system/cpu"
)

var efiVarsDir = "/sys/firmware/efi/efivars"
var cpuDir = "/sys/devices/system/cpu"
var cpupowerCmd = "/usr/bin/cpupower"
var isCPU = regexp.MustCompile(`^cpu\d+$`)
var isState = regexp.MustCompile(`^state\d+$`)
var perfCnt = 0
var govCnt = 0
var latCnt = 0

// GetPerfBias retrieve CPU performance configuration from the system
func GetPerfBias() string {
	isPBCpu := regexp.MustCompile(`analyzing CPU \d+`)
	isPBias := regexp.MustCompile(`perf-bias: \d+`)
	setAll := true
	str := ""
	oldpb := "99"
	cmdName := cpupowerCmd
	cmdArgs := []string{"-c", "all", "info", "-b"}

	if !supportsPerfBiasSettings() {
		return "all:none"
	}

	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		WarningLog("There was an error running external command 'cpupower -c all info -b': %v, output: %s", err, cmdOut)
		return "all:none"
	}

	for k, line := range strings.Split(strings.TrimSpace(string(cmdOut)), "\n") {
		switch {
		case line == notSupportedX86 || line == notSupportedIBM:
			// safety net - check already done in supportsPerfBiasSettings()
			return "all:none"
		case isPBCpu.MatchString(line):
			str = str + fmt.Sprintf("cpu%d", k/2)
		case isPBias.MatchString(line):
			pb := strings.Split(line, ":")
			if len(pb) < 2 {
				continue
			}
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
	if !supportsPerfBiasSettings() {
		return nil
	}

	cpu := ""
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if len(fields) < 2 {
			continue
		}
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
		} else {
			cpu = fields[0]
		}
		out, err := exec.Command(cpupowerCmd, "-c", cpu, "set", "-b", fields[1]).CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command 'cpupower -c %s set -b %s': %v, output: %s", cpu, fields[1], err, out)
			return err
		}
	}
	return nil
}

// supportsPerfBiasSettings checks, if Perf Bias is supported for the system
func supportsPerfBiasSettings() bool {
	setPerf := true
	if GetCSP() == "azure" {
		PrintLog(perfCnt, "warn", "Perf Bias settings not supported on '%s'\n", CSPAzureLong)
		setPerf = false
	} else if !supportsPerfBias() {
		setPerf = false
	}
	if perfCnt == 0 {
		perfCnt += 1
	}
	return setPerf
}

// SecureBootEnabled checks, if the system is in lock-down mode
func SecureBootEnabled() bool {
	var isSecBootFileName = regexp.MustCompile(`^SecureBoot-\w[\w-]+`)
	if _, err := os.Stat(efiVarsDir); os.IsNotExist(err) {
		InfoLog("no EFI directory '%+s' found, assuming legacy boot", efiVarsDir)
		return false
	}
	secureBootFile := ""
	_, efiFiles := ListDir(efiVarsDir, "the available efi variables")
	for _, eFile := range efiFiles {
		if isSecBootFileName.MatchString(eFile) {
			// work with the first file matching 'SecureBoot-*'
			secureBootFile = path.Join(efiVarsDir, eFile)
			break
		}
	}
	if secureBootFile == "" {
		InfoLog("no EFI SecureBoot file (SecureBoot-*) found in '%s', assuming legacy boot", efiVarsDir)
		return false
	}

	content, err := os.ReadFile(secureBootFile)
	if err != nil {
		InfoLog("failed to read EFI SecureBoot file '%s': %v", secureBootFile, err)
		return false
	}
	lastElement := content[len(content)-1]
	if lastElement == 1 {
		DebugLog("secure boot enabled - '%v'", content)
		return true
	}
	DebugLog("secure boot disabled - '%v'", content)
	return false
}

// supportsPerfBias check, if the system will support CPU performance settings
func supportsPerfBias() bool {
	cmdName := cpupowerCmd
	cmdArgs := []string{"info", "-b"}

	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return false
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	PrintLog(perfCnt, "info", "supportsPerfBias - command 'cpupower info -b' returned: %s - %v", string(cmdOut), err)
	if err != nil || (err == nil && (strings.Contains(string(cmdOut), notSupportedX86) || strings.Contains(string(cmdOut), notSupportedIBM))) {
		// does not support perf bias
		if SecureBootEnabled() {
			PrintLog(perfCnt, "warn", "Cannot set Perf Bias when SecureBoot is enabled, skipping")
		} else {
			PrintLog(perfCnt, "warn", "Perf Bias settings not supported by the system")
		}
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

	if !supportsGovernorSettings("") {
		gGov["all"] = "none"
		return gGov
	}

	dirCont, err := os.ReadDir(cpuDir)
	if err != nil {
		PrintLog(govCnt, "warn", "Governor settings not supported by the system - %v", err)
		gGov["all"] = "none"
		return gGov
	}
	for _, entry := range dirCont {
		cpuName := entry.Name()
		if isCPU.MatchString(cpuName) {
			if !isCPUonline(cpuName) {
				DebugLog("GetGovernor - Skipping CPU '%s' because it's offline", cpuName)
				continue
			}
			if _, err = os.Stat(path.Join(cpuDir, cpuName, "cpufreq", "scaling_governor")); os.IsNotExist(err) {
				// os/sys/devices/system/cpu/cpu*/online.Stat needs cpuDir as path - including /sys
				tmpfile := path.Join(cpuDir, cpuName, "cpufreq", "scaling_governor")
				InfoLog("Unable to identify the current scaling governor for CPU '%s', missing file '%s'.", cpuName, tmpfile)
				gov = ""
			} else {
				// GetSysString needs cpuDirSys as path - without /sys
				gov, _ = GetSysString(path.Join(cpuDirSys, cpuName, "cpufreq", "scaling_governor"))
			}
			if gov == "" || gov == "NA" || gov == "PNA" {
				gov = "none"
			}
			if oldgov == "99" {
				// starting point
				oldgov = gov
			}
			if oldgov != gov {
				setAll = false
			}
			gGov[cpuName] = gov
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
	if !supportsGovernorSettings(value) {
		return nil
	}

	cpu := ""
	tst := ""
	for k, entry := range strings.Fields(value) {
		fields := strings.Split(entry, ":")
		if len(fields) < 2 {
			continue
		}
		if fields[0] != "all" {
			cpu = strconv.Itoa(k)
			tst = cpu
		} else {
			cpu = fields[0]
			tst = "cpu0"
		}
		if !isValidGovernor(tst, fields[1]) {
			WarningLog("'%s' is not a valid governor for cpu '%s', skipping.", fields[1], tst)
			continue
		}
		out, err := exec.Command(cpupowerCmd, "-c", cpu, "frequency-set", "-g", fields[1]).CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command 'cpupower -c %s frequency-set -g %s': %v, output: %s", cpu, fields[1], err, out)
			return err
		}
	}
	return nil
}

// supportsGovernorSettings checks, if governor settings supported by the system
func supportsGovernorSettings(value string) bool {
	setGov := true
	if GetCSP() == "azure" {
		PrintLog(govCnt, "warn", "Governor settings not supported on '%s'\n", CSPAzureLong)
		setGov = false
	} else if value == "all:none" {
		PrintLog(govCnt, "warn", "Governor settings not supported by the system")
		setGov = false
	} else if !CmdIsAvailable(cpupowerCmd) {
		PrintLog(govCnt, "warn", "command '%s' not found", cpupowerCmd)
		setGov = false
	} else if !chkKernelCmdline() || !chkCPUDriver() {
		PrintLog(govCnt, "warn", "Governor settings not supported by the system - see log for details")
		setGov = false
	} else if _, err := os.Stat(path.Join(cpuDir, "cpu0/cpufreq/scaling_governor")); os.IsNotExist(err) {
		// check only first cpu - cpu0, not all
		PrintLog(govCnt, "warn", "Governor settings not supported by the system. Missing scaling governor")
		setGov = false
	}
	if govCnt == 0 {
		govCnt += 1
	}
	return setGov
}

// chkKernelCmdline
func chkKernelCmdline() bool {
	validCmdline := true
	val := ParseCmdline("/proc/cmdline", "idle")
	if val == "poll" || val == "halt" {
		PrintLog(govCnt, "info", "The entire CPUIdle subsystem is disabled, acpi_idle and intel_idle drivers are disabled altogether (idle=%s).", val)
		validCmdline = false
	}
	if val == "nomwait" {
		PrintLog(govCnt, "info", "intel_idle driver is disabled, C-states handled by acpi_idle driver and the BIOS ACPI tables (idle=%s).", val)
		validCmdline = false
	}
	if val := ParseCmdline("/proc/cmdline", "cpuidle.off"); val == "1" {
		PrintLog(govCnt, "info", "CPU idle time management entirely disabled (cpuidle.off=%s).", val)
		validCmdline = false
	}
	if val := ParseCmdline("/proc/cmdline", "intel_idle.max_cstate"); val == "0" {
		PrintLog(govCnt, "info", "intel_idle driver is disabled, C-states handled by acpi_idle driver and the BIOS ACPI tables (intel_idle.max_cstate=%s).", val)
		validCmdline = false
	}
	if val := ParseCmdline("/proc/cmdline", "intel_pstate"); val == "disable" {
		PrintLog(govCnt, "info", "intel_pstate driver is disabled, C-states handled by acpi_idle driver and the BIOS ACPI tables (intel_pstate=%s).", val)
		validCmdline = false
	}
	return validCmdline
}

// chkCPUDriver checks, which CPUIdle driver is used
// currently only intel_idle and intel_pstate are supported
func chkCPUDriver() bool {
	validDriver := true
	val, err := os.ReadFile(path.Join(cpuDir, "cpuidle/current_driver"))
	if err != nil || !strings.Contains(string(val), "intel_idle") {
		PrintLog(govCnt, "info", "Unsupported CPUIdle driver '%s' used (%v)", strings.TrimSpace(string(val)), err)
		validDriver = false
	}
	val, err = os.ReadFile(path.Join(cpuDir, "cpu0/cpufreq/scaling_driver"))
	if err != nil || !strings.Contains(string(val), "intel_pstate") {
		PrintLog(govCnt, "info", "Unsupported CPU driver '%s' used (%v)", strings.TrimSpace(string(val)), err)
		validDriver = false
	}
	_, err = os.Stat(path.Join(cpuDir, "intel_pstate"))
	if err != nil {
		PrintLog(govCnt, "info", "Missing directory '%s'", path.Join(cpuDir, "intel_pstate"))
		validDriver = false
	} else {
		val, err := os.ReadFile(path.Join(cpuDir, "intel_pstate/status"))
		if err != nil || !strings.Contains(string(val), "active") {
			PrintLog(govCnt, "info", "Status of intel_pstate driver is '%s' (%v)", strings.TrimSpace(string(val)), err)
			validDriver = false
		}
	}
	return validDriver
}

// isCPUonline checks, if the cpu is currently online
// check cotent of file /sys/devices/system/cpu/cpu*/online
// '0' means offline, '1' means online
func isCPUonline(cpu string) bool {
	// cpu0 is special case. As there is always a cpu0 and as it can not
	// be brought offline, there is no file 'online' available in sysfs
	if cpu == "cpu0" {
		return true
	}
	val, err := os.ReadFile(path.Join(cpuDir, cpu, "/online"))
	if err == nil && strings.TrimSpace(string(val)) == "1" {
		return true
	}
	return false
}

// isValidGovernor check, if the system will support CPU frequency settings
func isValidGovernor(cpu, gov string) bool {
	val, err := os.ReadFile(path.Join(cpuDir, cpu, "/cpufreq/scaling_available_governors"))
	if err == nil && strings.Contains(string(val), gov) {
		return true
	}
	return false
}

// GetFLInfo retrieve CPU latency configuration from the system and returns
// the current latency,
// the latency states of all CPUs to save Latency states for 'revert',
// if cpu states differ
// return lat, savedStates, cpuStateDiffer
func GetFLInfo() (string, string, bool) {
	lat := 0
	maxlat := 0
	supported := false
	savedStates := ""
	stateDisabled := false
	cpuStateDiffer := false
	cpuStateMap := make(map[string]string)

	if !supportsForceLatencySettings("") {
		return "all:none", "all:none", cpuStateDiffer
	}

	// read /sys/devices/system/cpu
	dirCont, err := os.ReadDir(cpuDir)
	if err != nil {
		WarningLog("Latency settings not supported by the system")
		return "all:none", "all:none", cpuStateDiffer
	}
	for _, entry := range dirCont {
		// cpu0 ... cpuXY
		cpuName := entry.Name()
		if isCPU.MatchString(cpuName) {
			if !isCPUonline(cpuName) {
				DebugLog("GetFLInfo - Skipping CPU '%s' because it's offline", cpuName)
				continue
			}
			// read /sys/devices/system/cpu/cpu*/cpuidle
			cpudirCont, err := os.ReadDir(path.Join(cpuDir, cpuName, "cpuidle"))
			if err != nil {
				// idle settings not supported for cpuName
				continue
			}
			supported = true
			for _, centry := range cpudirCont {
				stateName := centry.Name()
				// state0 ... stateXY
				if isState.MatchString(stateName) {
					// read /sys/devices/system/cpu/cpu*/cpuidle/state*/disable
					state, _ := GetSysString(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "disable"))
					// save latency states for 'revert'
					// savedStates = "cpu1:state0:0 cpu1:state1:0"
					savedStates = savedStates + " " + cpuName + ":" + stateName + ":" + state
					cpuStateMap[cpuName] = cpuStateMap[cpuName] + " " + state
					// read /sys/devices/system/cpu/cpu*/cpuidle/state*/latency
					lattmp, _ := GetSysInt(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "latency"))
					if lattmp > maxlat {
						maxlat = lattmp
					}
					if state == "1" {
						stateDisabled = true
					} else {
						lat = lattmp
					}
				}
			}
		}
	}
	// check, if all cpus have the same state settings
	cpuStateDiffer = checkCPUState(cpuStateMap)

	if !stateDisabled {
		// start value for force latency, if no states are disabled
		lat = maxlat
	}

	rval := strconv.Itoa(lat)
	if !supported {
		savedStates = "all:none"
		rval = "all:none"
	}
	return rval, savedStates, cpuStateDiffer
}

// SetForceLatency set CPU latency configuration to the system
func SetForceLatency(value, savedStates string, revert bool) error {
	oldState := ""

	if !supportsForceLatencySettings(value) {
		return nil
	}

	flval, _ := strconv.Atoi(value) // decimal value for force latency

	dirCont, err := os.ReadDir(cpuDir)
	if err != nil {
		WarningLog("Latency settings not supported by the system")
		return err
	}
	for _, entry := range dirCont {
		// cpu0 ... cpuXY
		cpuName := entry.Name()
		if isCPU.MatchString(cpuName) {
			if !isCPUonline(cpuName) {
				DebugLog("SetForceLatency - Skipping CPU '%s' because it's offline", cpuName)
				continue
			}
			cpudirCont, errns := os.ReadDir(path.Join(cpuDir, cpuName, "cpuidle"))
			if errns != nil {
				WarningLog("idle settings not supported for '%s'", cpuName)
				continue
			}
			for _, centry := range cpudirCont {
				stateName := centry.Name()
				// state0 ... stateXY
				if isState.MatchString(stateName) {
					// read /sys/devices/system/cpu/cpu*/cpuidle/state*/latency
					lat, _ := GetSysInt(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "latency"))
					// write /sys/devices/system/cpu/cpu*/cpuidle/state*/disable
					if revert {
						// revert
						for _, ole := range strings.Fields(savedStates) {
							FLFields := strings.Split(ole, ":")
							if len(FLFields) > 2 {
								if FLFields[0] == cpuName && FLFields[1] == stateName {
									oldState = FLFields[2]
								}
							}
						}
						if oldState != "" {
							err = SetSysString(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "disable"), oldState)
							// clear latency value for next cpu/state cycle
							oldState = ""
						}
					} else {
						// apply
						oldState, _ = GetSysString(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "disable"))
						// save old latency states for 'revert'
						if lat > flval {
							// set new latency states
							err = SetSysString(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "disable"), "1")
						}
						if lat <= flval && oldState == "1" {
							// reset previous set latency state
							err = SetSysString(path.Join(cpuDirSys, cpuName, "cpuidle", stateName, "disable"), "0")
						}
					}
				}
			}
		}
	}

	return err
}

// supportsForceLatencySettings checks, if Force Latency can be set
func supportsForceLatencySettings(value string) bool {
	setLatency := true
	if GetCSP() == "azure" {
		PrintLog(latCnt, "warn", "Latency settings are not supported on '%s'\n", CSPAzureLong)
		setLatency = false
	} else if runtime.GOARCH == "ppc64le" {
		// latency settings are only relevant for Intel-based systems
		PrintLog(latCnt, "warn", "Latency settings not relevant for '%s' systems", runtime.GOARCH)
		setLatency = false
	} else if value == "all:none" {
		PrintLog(latCnt, "warn", "Latency settings not supported by the system")
		setLatency = false
	} else if _, err := os.Stat(path.Join(cpuDir, "cpu0")); os.IsNotExist(err) {
		// check only first cpu - cpu0, not all
		PrintLog(latCnt, "warn", "Latency settings not supported by the system")
		setLatency = false
	} else if currentCPUDriver() == "none" {
		PrintLog(latCnt, "warn", "Latency settings not supported by the system, no active cpuidle driver")
		setLatency = false
	}
	if latCnt == 0 {
		latCnt += 1
	}
	return setLatency
}

// currentCPUDriver returns the current active cpuidle driver from
// /sys/devices/system/cpu/cpuidle/current_driver
func currentCPUDriver() string {
	cpuDriver := "none"
	cpuDriverFile := path.Join(cpuDir, "/cpuidle/current_driver")
	if _, err := os.Stat(cpuDriverFile); os.IsNotExist(err) {
		PrintLog(latCnt, "info", "File '%s' not found - %v", cpuDriverFile, err)
		return cpuDriver
	}
	if val, err := os.ReadFile(cpuDriverFile); err != nil {
		PrintLog(latCnt, "info", "Problems reading file '%s' - %+v\n", cpuDriverFile, err)
	} else {
		cpuDriver = strings.TrimSpace(string(val))
	}
	return cpuDriver
}

// CPUPlatform returns the CPU platform
// read /sys/devices/cpu/caps/pmu_name
func CPUPlatform() string {
	cpuPlatformFile := "/sys/devices/cpu/caps/pmu_name"
	cpuPlatform, err := os.ReadFile(cpuPlatformFile)
	if err != nil {
		InfoLog("Problems reading file '%s' - %+v\n", cpuPlatformFile, err)
	}
	return strings.TrimSpace(string(cpuPlatform))
}

// checkCPUState checks, if all cpus have the same state settings
// returns true, if the cpu states differ
func checkCPUState(csMap map[string]string) bool {
	ret := false
	oldcpuState := ""
	for _, cpuState := range csMap {
		if oldcpuState == "" {
			oldcpuState = cpuState
		}
		if oldcpuState != cpuState {
			ret = true
			break
		}
	}
	return ret
}

// GetdmaLatency retrieve DMA latency configuration from the system
func GetdmaLatency() string {
	latency := make([]byte, 4)
	dmaLatency, err := os.OpenFile("/dev/cpu_dma_latency", os.O_RDONLY, 0600)
	if err != nil {
		WarningLog("GetForceLatency: failed to open cpu_dma_latency - %v", err)
	}
	_, err = dmaLatency.Read(latency)
	if err != nil {
		WarningLog("GetForceLatency: reading from '/dev/cpu_dma_latency' failed: %v", err)
	}
	// Close the file handle after the latency value is no longer maintained
	defer dmaLatency.Close()

	ret := fmt.Sprintf("%v", binary.LittleEndian.Uint32(latency))
	return ret
}
