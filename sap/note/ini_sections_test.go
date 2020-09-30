package note

import (
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
)

var PCTestBaseConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/note/1557506")
var blockDev = system.CollectBlockDeviceInfo()
var services = system.GetAvailServices()

func TestGetServiceName(t *testing.T) {
	val := system.GetServiceName("uuidd.socket")
	if val != "uuidd.socket" && val != "" {
		t.Error(val)
	}
	val = system.GetServiceName("sysstat")
	if val != "sysstat.service" && val != "" {
		t.Error(val)
	}
	val = system.GetServiceName("sysstat.service")
	if val != "sysstat.service" && val != "" {
		t.Error(val)
	}
	val = system.GetServiceName("UnkownService")
	if val != "" {
		t.Error(val)
	}
}

func TestOptSysctlVal(t *testing.T) {
	// remember the change in saptune 2.0 (SAP and Alliance decision)
	// use exactly the value from the config file. No calculation any more
	op := txtparser.Operator("=")
	val := OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300 200", "100 330 180")
	if val != "100	330	180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300", "100 330 180")
	if val != "" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "", "100 330 180")
	if val != "" {
		t.Error(val)
	}
	op = txtparser.Operator("<")
	val = OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "180")
	if val != "180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "120")
	if val != "120" {
		t.Error(val)
	}
	op = txtparser.Operator(">")
	val = OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "180")
	if val != "180" {
		t.Error(val)
	}
	val = OptSysctlVal(op, "TestParam", "120", "120")
	if val != "120" {
		t.Error(val)
	}
}

func TestGetBlkVal(t *testing.T) {
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	_, _, err := GetBlkVal("IO_SCHEDULER_sda", &tblck)
	if err != nil {
		t.Error(err)
	}
}

func TestOptBlkVal(t *testing.T) {
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info := OptBlkVal("IO_SCHEDULER_sda", "noop", &tblck, blckOK)
	if val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info := OptBlkVal("IO_SCHEDULER_sda", "none", &tblck, blckOK)
		if val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}

	val, info = OptBlkVal("IO_SCHEDULER_sda", "NoOP", &tblck, blckOK)
	if val != "NoOP" && val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal("IO_SCHEDULER_sda", "NoNE", &tblck, blckOK)
		if val != "NoNE" && val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "deadline", &tblck, blckOK)
	if val != "deadline" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal("IO_SCHEDULER_sda", "mq-deadline", &tblck, blckOK)
		if val != "mq-deadline" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "NoOp,NoNe", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", " noop , none ", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "hugo", &tblck, blckOK)
	if val != "hugo" && info != "NA" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
	}

	val, info = OptBlkVal("NRREQ_sda", "512", &tblck, blckOK)
	if val != "512" {
		t.Error(val)
	}
	val, info = OptBlkVal("NRREQ_sdb", "0", &tblck, blckOK)
	if val != "1024" {
		t.Error(val)
	}
	val, info = OptBlkVal("NRREQ_sdc", "128", &tblck, blckOK)
	if val != "128" {
		t.Error(val)
	}
}

func TestSetBlkVal(t *testing.T) {
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info, err := GetBlkVal("IO_SCHEDULER_sda", &tblck)
	oval := val
	if err != nil {
		t.Error(err)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" {
		t.Error(val, info)
	}
	// apply - value not used, but map changed above in optimise
	err = SetBlkVal("IO_SCHEDULER_sda", "notUsed", &tblck, false)
	// revert - value will be used to change map before applying
	err = SetBlkVal("IO_SCHEDULER_sda", oval, &tblck, true)
}

//GetLimitsVal
func TestOptLimitsVal(t *testing.T) {
	val := OptLimitsVal("@sdba soft nofile NA", "@sdba soft nofile 32800")
	if val != "@sdba soft nofile 32800" {
		t.Error(val)
	}
	val = OptLimitsVal("@sdba soft nofile 75536", "@sdba soft nofile 32800")
	if val != "@sdba soft nofile 32800" {
		t.Error(val)
	}
}

//SetLimitsVal apply and revert

func TestGetVMVal(t *testing.T) {
	val := GetVMVal("THP")
	if val != "always" && val != "madvise" && val != "never" {
		t.Errorf("wrong value '%+v' for THP.\n", val)
	}
	val = GetVMVal("KSM")
	if val != "1" && val != "0" {
		t.Errorf("wrong value '%+v' for KSM.\n", val)
	}
}

func TestOptVMVal(t *testing.T) {
	val := OptVMVal("THP", "always")
	if val != "always" {
		t.Error(val)
	}
	val = OptVMVal("THP", "unknown")
	if val != "never" {
		t.Error(val)
	}
	val = OptVMVal("KSM", "1")
	if val != "1" {
		t.Error(val)
	}
	val = OptVMVal("KSM", "2")
	if val != "0" {
		t.Error(val)
	}
	val = OptVMVal("UNKOWN_PARAMETER", "unknown")
	if val != "unknown" {
		t.Error(val)
	}
}

func TestSetVMVal(t *testing.T) {
	newval := ""
	oldval := GetVMVal("THP")
	if oldval == "never" {
		newval = "always"
	} else {
		newval = "never"
	}
	err := SetVMVal("THP", newval)
	if err != nil {
		t.Error(err)
	}
	val := GetVMVal("THP")
	if val != newval {
		t.Error(val)
	}
	// set test value back
	err = SetVMVal("THP", oldval)
	if err != nil {
		t.Error(err)
	}

	oldval = GetVMVal("KSM")
	if oldval == "0" {
		newval = "1"
	} else {
		newval = "0"
	}
	err = SetVMVal("KSM", newval)
	if err != nil {
		t.Error(err)
	}
	val = GetVMVal("KSM")
	if val != newval {
		t.Error(val)
	}
	// set test value back
	err = SetVMVal("KSM", oldval)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCPUVal(t *testing.T) {
	val, _, _ := GetCPUVal("force_latency")
	if val != "all:none" {
		t.Logf("force_latency supported: '%s'\n", val)
	}
	val, _, _ = GetCPUVal("energy_perf_bias")
	if val != "all:none" {
		t.Logf("energy_perf_bias supported: '%s'\n", val)
	}
	val, _, _ = GetCPUVal("governor")
	if val != "all:none" && val != "" {
		t.Logf("governor supported: '%s'\n", val)
	}
}

func TestOptCPUVal(t *testing.T) {
	val := OptCPUVal("force_latency", "1000", "70")
	if val != "70" {
		t.Error(val)
	}

	val = OptCPUVal("energy_perf_bias", "all:15", "performance")
	if val != "all:0" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "cpu0:15 cpu1:6 cpu2:0", "performance")
	if val != "cpu0:0 cpu1:0 cpu2:0" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:15", "normal")
	if val != "all:6" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:15", "powersave")
	if val != "all:15" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:15", "unknown")
	if val != "all:0" {
		t.Error(val)
	}

	/* future feature
	val = OptCPUVal("energy_perf_bias", "cpu0:6 cpu1:6 cpu2:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:0 cpu1:6 cpu2:15" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:performance cpu1:normal cpu2:powersave" {
		t.Error(val)
	}
	*/

	val = OptCPUVal("governor", "all:powersave", "performance")
	if val != "all:performance" {
		t.Error(val)
	}
	val = OptCPUVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "performance")
	if val != "cpu0:performance cpu1:performance cpu2:performance" {
		t.Error(val)
	}
	/* future feature
	val = OptCPUVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Error(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Error(val)
	}
	*/
}

//SetCPUVal

func TestGetMemVal(t *testing.T) {
	val := GetMemVal("VSZ_TMPFS_PERCENT")
	if val == "-1" {
		t.Log("/dev/shm not found")
	}
	val = GetMemVal("ShmFileSystemSizeMB")
	if val == "-1" {
		t.Log("/dev/shm not found")
	}
	val = GetMemVal("UNKOWN_PARAMETER")
	if val != "" {
		t.Error(val)
	}
}

func TestOptMemVal(t *testing.T) {
	val := OptMemVal("VSZ_TMPFS_PERCENT", "47", "80", "80")
	if val != "80" {
		t.Error(val)
	}
	val = OptMemVal("VSZ_TMPFS_PERCENT", "-1", "75", "75")
	if val != "75" {
		t.Error(val)
	}

	size75 := uint64(system.GetTotalMemSizeMB()) * 75 / 100
	size80 := uint64(system.GetTotalMemSizeMB()) * 80 / 100

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "80")
	if val != strconv.FormatUint(size80, 10) {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "80")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "0")
	if val != strconv.FormatUint(size75, 10) {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "0")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "80")
	if val != "25605" {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "80")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "0")
	if val != "25605" {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "0")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("UNKOWN_PARAMETER", "16043", "0", "0")
	if val != "" {
		t.Error(val)
	}
	val = OptMemVal("UNKOWN_PARAMETER", "-1", "0", "0")
	if val != "" {
		t.Error(val)
	}
}

//SetMemVal

func TestGetRpmVal(t *testing.T) {
	val := GetRpmVal("rpm:glibc")
	if val == "" {
		t.Log("rpm 'glibc' not found")
	}
}

func TestOptRpmVal(t *testing.T) {
	val := OptRpmVal("rpm:glibc", "NO_OPT")
	if val != "NO_OPT" {
		t.Error(val)
	}
}

func TestSetRpmVal(t *testing.T) {
	val := SetRpmVal("NO_OPT")
	if val != nil {
		t.Error(val)
	}
}

func TestGetGrubVal(t *testing.T) {
	val := GetGrubVal("grub:processor.max_cstate")
	if val == "NA" {
		t.Log("'processor.max_cstate' not found in kernel cmdline")
	}
	val = GetGrubVal("grub:UNKNOWN")
	if val != "NA" {
		t.Error(val)
	}
}

func TestOptGrubVal(t *testing.T) {
	val := OptGrubVal("grub:processor.max_cstate", "NO_OPT")
	if val != "NO_OPT" {
		t.Error(val)
	}
}

func TestSetGrubVal(t *testing.T) {
	val := SetGrubVal("NO_OPT")
	if val != nil {
		t.Error(val)
	}
}

func TestGetServiceVal(t *testing.T) {
	wrong := false
	state := ""
	val := GetServiceVal("UnkownService")
	if val != "NA" {
		t.Error(val)
	}
	val = GetServiceVal("uuidd.socket")
	for _, st := range strings.Split(val, ",") {
		state = strings.TrimSpace(st)
		if state != "start" && state != "stop" && state != "NA" && state != "enable" && state != "disable" {
			wrong = true
		}
	}
	if wrong {
		t.Error(val)
	}
}

func TestOptServiceVal(t *testing.T) {
	val := OptServiceVal("UnkownService", "start")
	if val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("uuidd.socket", "start")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("uuidd.socket", "stop")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("uuidd.socket", "unknown")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("sysstat", "start")
	if val != "start" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("sysstat.service", "stop")
	if val != "stop" && val != "NA" {
		t.Error(val)
	}
	val = OptServiceVal("sysstat", "unknown")
	if val != "" && val != "NA" {
		t.Error(val)
	}
	wrong := false
	state := ""
	val = OptServiceVal("sysstat", "stop, start, unknown, disable, enable")
	for _, st := range strings.Split(val, ",") {
		state = strings.TrimSpace(st)
		if state != "stop" && state != "disable" && state != "NA" {
			wrong = true
		}
	}
	if wrong {
		t.Error(val)
	}
	wrong = false
	val = OptServiceVal("uuidd.socket", "enable")
	for _, st := range strings.Split(val, ",") {
		state = strings.TrimSpace(st)
		if state != "start" && state != "enable" && state != "NA" {
			wrong = true
		}
	}
	if wrong {
		t.Error(val)
	}
}

func TestSetServiceVal(t *testing.T) {
	val := SetServiceVal("UnkownService", "start")
	if val != nil {
		t.Error(val)
	}
	_ = system.SystemctlDisable("sysstat.service")
	val = SetServiceVal("sysstat.service", "enable")
	if val != nil {
		t.Error(val)
	}
	val = SetServiceVal("sysstat.service", "disable")
	if val != nil {
		t.Error(val)
	}
}

func TestGetLoginVal(t *testing.T) {
	val, err := GetLoginVal("Unkown")
	if val != "" || err != nil {
		t.Error(val)
	}

	val, err = GetLoginVal("UserTasksMax")
	if _, errno := os.Stat("/etc/systemd/logind.conf.d/saptune-UserTasksMax.conf"); errno != nil {
		if !os.IsNotExist(errno) {
			if val != "" || err == nil {
				t.Error(val)
			}
		} else {
			if val != "NA" || err != nil {
				t.Error(val)
			}
		}
	}
}

func TestOptLoginVal(t *testing.T) {
	val := OptLoginVal("unkown")
	if val != "unkown" {
		t.Error(val)
	}
	val = OptLoginVal("infinity")
	if val != "infinity" {
		t.Error(val)
	}
	val = OptLoginVal("")
	if val != "" {
		t.Error(val)
	}
}

func TestSetLoginVal(t *testing.T) {
	utmFile := "/etc/systemd/logind.conf.d/saptune-UserTasksMax.conf"
	val := "18446744073709"

	err := SetLoginVal("UserTasksMax", val, false)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(utmFile); err != nil {
		t.Error(err)
	}
	if !system.CheckForPattern(utmFile, val) {
		t.Errorf("wrong value in file '%s'\n", utmFile)
	}
	val = "infinity"
	err = SetLoginVal("UserTasksMax", val, false)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(utmFile); err != nil {
		t.Error(err)
	}
	if !system.CheckForPattern(utmFile, val) {
		t.Errorf("wrong value in file '%s'\n", utmFile)
	}
	val = "10813"
	err = SetLoginVal("UserTasksMax", val, true)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(utmFile); err == nil {
		os.Remove(utmFile)
		t.Errorf("file '%s' still exists\n", utmFile)
	}
}

func TestGetPagecacheVal(t *testing.T) {
	prepare := LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val := GetPagecacheVal("ENABLE_PAGECACHE_LIMIT", &prepare)
	if val != "yes" && val != "no" {
		t.Error(val)
	}
	if prepare.VMPagecacheLimitMB == 0 && val != "no" {
		t.Error(val)
	}
	if prepare.VMPagecacheLimitMB > 0 && val != "yes" {
		t.Error(val)
	}

	prepare = LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val = GetPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, &prepare)
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Error(val)
	}

	prepare = LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val = GetPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", &prepare)
	if prepare.VMPagecacheLimitMB == 0 && val != "" {
		t.Error(val)
	}
	if prepare.VMPagecacheLimitMB > 0 && val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Error(val)
	}

	prepare = LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val = GetPagecacheVal("UNKOWN", &prepare)
	if val != "" {
		t.Error(val)
	}
}

func TestOptPagecacheVal(t *testing.T) {
	initPrepare, _ := LinuxPagingImprovements{PagingConfig: PCTestBaseConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare := initPrepare.(LinuxPagingImprovements)

	val := OptPagecacheVal("UNKNOWN", "unknown", &prepare)
	if val != "unknown" {
		t.Error(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "yes", &prepare)
	if val != "yes" {
		t.Error(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "no", &prepare)
	if val != "no" {
		t.Error(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "unknown", &prepare)
	if val != "no" {
		t.Error(val)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "2", &prepare)
	if val != "2" {
		t.Error(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Error(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "1", &prepare)
	if val != "1" {
		t.Error(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Error(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "0", &prepare)
	if val != "0" {
		t.Error(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Error(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "unknown", &prepare)
	if val != "1" {
		t.Error(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Error(val, prepare.VMPagecacheLimitIgnoreDirty)
	}

	PCTestConf := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest1.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Error(val, prepare.VMPagecacheLimitMB)
	}

	calc := system.GetMainMemSizeMB() * 2 / 100
	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest2.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Error(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Error(val, prepare.VMPagecacheLimitMB)
	}
	if val != strconv.FormatUint(calc, 10) {
		t.Error(val, calc)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest3.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Error(val, prepare.VMPagecacheLimitMB)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest4.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Error(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Error(val, prepare.VMPagecacheLimitMB)
	}
	if val != strconv.FormatUint(calc, 10) {
		t.Error(val, calc)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest5.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Error(val, prepare.VMPagecacheLimitMB)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest6.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Error(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Error(val, prepare.VMPagecacheLimitMB)
	}
	if val != "641" {
		t.Error(val)
	}

}

func TestSetPagecacheVal(t *testing.T) {
	prepare := LinuxPagingImprovements{PagingConfig: PCTestBaseConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}
	val := SetPagecacheVal("UNKNOWN", &prepare)
	if val != nil {
		t.Error(val)
	}
}
