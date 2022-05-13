package system

// Manipulate sysctl switches.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// mapping of system parameter names to configuration names
const (
	SysctlPagecacheLimitMB          = "vm.pagecache_limit_mb"
	SysctlPagecacheLimitIgnoreDirty = "vm.pagecache_limit_ignore_dirty"
	SysctlDirtyRatio                = "vm.dirty_ratio"
	SysctlDirtyBackgroundRatio      = "vm.dirty_background_ratio"
	SysKernelTHPEnabled             = "kernel.mm.transparent_hugepage.enabled"
	SysKSMRun                       = "kernel.mm.ksm.run"
)

// sysctlDirs contains all locations sysctl is searching for parameter settings.
// see comment in /etc/sysctl.conf and man page sysctl.conf(5)
var sysctlDirs = []string{"/etc/sysctl.conf", "/run/sysctl.d/", "/etc/sysctl.d/", "/usr/local/lib/sysctl.d/", "/usr/lib/sysctl.d/", "/lib/sysctl.d/", "/boot/"}

var sysctlParms = sysctlDefined{}
var sysctlWarn = map[string]string{}

// sysctlEntry contains the 'sysctl config filename - value' pair
type sysctlEntry struct {
	File  string
	Value string
}

// sysctlConf contains a list of all files-value pairs for the related
// sysctl parameter
type sysctlConf []sysctlEntry

// sysctlDefined contains all sysctl parameter, which are defined in the
// sysctl config files of the system
type sysctlDefined map[string]sysctlConf

// ChkForSysctlDoubles checks if the given sysctl parameter is additional set
// in a sysctl system configuration file
func ChkForSysctlDoubles(param string) string {
	info := ""
	if len(sysctlParms[param]) > 0 {
		// found double
		for _, entries := range sysctlParms[param] {
			txt := entries.File + "(" + entries.Value + ")"
			if info == "" {
				info = "sysctl config file " + txt
			} else {
				info = info + ", " + txt
			}
		}
		printDoubleWarning(param, info)
	}
	return info
}

// printDoubleWarning checks, if we need to print a sysctl double warning
func printDoubleWarning(param, info string) {
	warn := false
	if _, ok := sysctlWarn[param]; !ok {
		warn = true
	}
	if warn {
		// print warning
		WarningLog("Parameter '%s' additional defined in the following %s.", param, info)
		sysctlWarn[param] = info
	}
}

// CollectGlobalSysctls collects all sysctl parameters defined in all
// of the sysctl.conf related files
func CollectGlobalSysctls() {
	fileList := make(map[string]string)

	for _, file := range getAllSysctlFiles() {
		// check all config files mentioned in /etc/sysctl.conf and
		// the sysctl.conf(5) man page
		info, err := os.Lstat(file)
		if err != nil {
			// file or directory does not exist
			continue
		}
		switch mode := info.Mode(); {
		case mode.IsRegular():
			fileList[file] = file
		case mode&os.ModeSymlink != 0:
			// symlink
			origFile, err := filepath.EvalSymlinks(file)
			if err != nil {
				continue
			}
			fileList[origFile] = file
		}
	}

	for _, sfile := range fileList {
		sconf, err := parseSysctlConfFile(sfile)
		if err != nil {
			// skip file
			continue
		}
		for param := range sconf {
			sysctlcnf := append(sysctlParms[param], sconf[param])
			sysctlParms[param] = sysctlcnf
		}
	}
}

// getAllSysctlFiles retrieves all sysctl config files from all
// locations/directories
func getAllSysctlFiles() []string {
	files := []string{}
	for _, file := range sysctlDirs {
		info, err := os.Stat(file)
		if err != nil {
			// file or directory does not exist
			continue
		}
		if info.IsDir() {
			for f := range GetFiles(file) {
				if (file == "/boot/" && !strings.HasPrefix(f, "sysctl.conf-")) && !strings.HasSuffix(f, ".conf") {
					// wrong file name format, skip file
					continue
				}
				files = append(files, file+f)
			}
		} else {
			files = append(files, file)
		}
	}
	return files
}

// parseSysctlConfFile parses a special sysctl config file and returns
// the key-value pairs of the contained sysctl parameters
func parseSysctlConfFile(file string) (map[string]sysctlEntry, error) {
	entries := make(map[string]sysctlEntry)
	content, err := ReadConfigFile(file, false)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Line is a comment
			continue
		}
		if eqChar := strings.IndexRune(line, '='); eqChar != -1 {
			// Line is a key-value pair
			key := strings.TrimSpace(line[0:eqChar])
			entries[key] = sysctlEntry{
				File:  file,
				Value: strings.Trim(strings.TrimSpace(line[eqChar+1:]), `"`),
			}
		}
	}
	return entries, nil
}

// GetSysctlString read a sysctl key and return the string value.
func GetSysctlString(parameter string) (string, error) {
	val, err := ioutil.ReadFile(path.Join("/proc/sys", strings.Replace(parameter, ".", "/", -1)))
	if err != nil {
		WarningLog("Failed to read sysctl key '%s': %v", parameter, err)
		return "PNA", err
	}
	return strings.TrimSpace(string(val)), nil
}

// GetSysctlInt read an integer sysctl key.
func GetSysctlInt(parameter string) (int, error) {
	value, err := GetSysctlString(parameter)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetSysctlUint64 read an uint64 sysctl key.
func GetSysctlUint64(parameter string) (uint64, error) {
	value, err := GetSysctlString(parameter)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(value, 10, 64)
}

// GetSysctlUint64Field extracts an uint64 value from a sysctl key of many fields.
func GetSysctlUint64Field(param string, field int) (uint64, error) {
	fields, err := GetSysctlString(param)
	if err == nil {
		allFields := consecutiveSpaces.Split(fields, -1)
		if field < len(allFields) {
			value, err := strconv.ParseUint(allFields[field], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("Failed to read sysctl key field '%s' %d: %v", param, field, err)
			}
			return value, nil
		}
	}
	return 0, err
}

// SetSysctlString write a string sysctl value.
func SetSysctlString(parameter, value string) error {
	if value == "PNA" {
		WarningLog("value is '%s', so sysctl key '%s' is/was not supported by os, skipping.", value, parameter)
		return nil
	}
	err := ioutil.WriteFile(path.Join("/proc/sys", strings.Replace(parameter, ".", "/", -1)), []byte(value), 0644)
	if os.IsNotExist(err) {
		WarningLog("sysctl key '%s' is not supported by os, skipping.", parameter)
	} else if err != nil {
		WarningLog("Failed to write sysctl key '%s': %v", parameter, err)
		return err
	}
	return nil
}

// SetSysctlInt write an integer sysctl value.
func SetSysctlInt(parameter string, value int) error {
	err := SetSysctlString(parameter, strconv.Itoa(value))
	return err
}

// SetSysctlUint64 write an integer sysctl value.
func SetSysctlUint64(parameter string, value uint64) error {
	err := SetSysctlString(parameter, strconv.FormatUint(value, 10))
	return err
}

// SetSysctlUint64Field write an integer sysctl value into the specified field pf the key.
func SetSysctlUint64Field(param string, field int, value uint64) error {
	fields, err := GetSysctlString(param)
	if err != nil {
		return err
	}
	allFields := consecutiveSpaces.Split(fields, -1)
	if field < len(allFields) {
		allFields[field] = strconv.FormatUint(value, 10)
		err = SetSysctlString(param, strings.Join(allFields, " "))
	} else {
		err = fmt.Errorf("Failed to write sysctl key field '%s' %d: %v", param, field, err)
	}
	return err
}

// IsPagecacheAvailable check, if system supports pagecache limit
func IsPagecacheAvailable() bool {
	_, err := ioutil.ReadFile(path.Join("/proc/sys", strings.Replace(SysctlPagecacheLimitMB, ".", "/", -1)))
	if err == nil {
		return true
	}
	return false
}
