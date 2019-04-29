package note

import (
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"
)

var PCTestBaseConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/note/1557506")

func TestGetServiceName(t *testing.T) {
	val := system.GetServiceName("uuidd.socket")
	if val != "uuidd.socket" && val != "" {
		t.Fatal(val)
	}
	val = system.GetServiceName("sysstat")
	if val != "sysstat.service" && val != "" {
		t.Fatal(val)
	}
	val = system.GetServiceName("sysstat.service")
	if val != "sysstat.service" && val != "" {
		t.Fatal(val)
	}
	val = system.GetServiceName("UnkownService")
	if val != "" {
		t.Fatal(val)
	}
}

func TestOptSysctlVal(t *testing.T) {
	op := txtparser.Operator("=")
	val := OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Fatal(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300 200", "100 330 180")
	if val != "100	330	180" {
		t.Fatal(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300", "100 330 180")
	if val != "" {
		t.Fatal(val)
	}
}

//GetBlkVal
func TestOptBlkVal(t *testing.T) {
	tblck := param.BlockDeviceQueue{param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val := OptBlkVal("IO_SCHEDULER_sda", "noop", &tblck)
	if val != "noop" {
		t.Fatal(val)
	}
	val = OptBlkVal("IO_SCHEDULER_sdb", "NoOP", &tblck)
	if val != "noop" {
		t.Fatal(val)
	}
	val = OptBlkVal("IO_SCHEDULER_sdc", "cfq", &tblck)
	if val != "cfq" {
		t.Fatal(val)
	}
	val = OptBlkVal("NRREQ_sda", "512", &tblck)
	if val != "512" {
		t.Fatal(val)
	}
	val = OptBlkVal("NRREQ_sdb", "0", &tblck)
	if val != "1024" {
		t.Fatal(val)
	}
	val = OptBlkVal("NRREQ_sdc", "128", &tblck)
	if val != "128" {
		t.Fatal(val)
	}
}

//SetBlkVal

//GetLimitsVal
func TestOptLimitsVal(t *testing.T) {
	val := OptLimitsVal("@sdba soft nofile NA", "@sdba soft nofile 32800")
	if val != "@sdba soft nofile 32800" {
		t.Fatal(val)
	}
	val = OptLimitsVal("@sdba soft nofile 75536", "@sdba soft nofile 32800")
	if val != "@sdba soft nofile 32800" {
		t.Fatal(val)
	}
}

//SetLimitsVal

func TestGetVMVal(t *testing.T) {
	if runtime.GOARCH != "ppc64le" {
		val := GetVMVal("THP")
		if val != "always" && val != "madvise" && val != "never" {
			t.Fatalf("wrong value '%+v' for THP.\n", val)
		}
	}
	val := GetVMVal("KSM")
	if val != "1" && val != "0" {
		t.Fatalf("wrong value '%+v' for KSM.\n", val)
	}
}

func TestOptVMVal(t *testing.T) {
	val := OptVMVal("THP", "always")
	if val != "always" {
		t.Fatal(val)
	}
	val = OptVMVal("THP", "unknown")
	if val != "never" {
		t.Fatal(val)
	}
	val = OptVMVal("KSM", "1")
	if val != "1" {
		t.Fatal(val)
	}
	val = OptVMVal("KSM", "2")
	if val != "0" {
		t.Fatal(val)
	}
	val = OptVMVal("UNKOWN_PARAMETER", "unknown")
	if val != "unknown" {
		t.Fatal(val)
	}
}

//SetVMVal

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
		t.Fatal(val)
	}

	val = OptCPUVal("energy_perf_bias", "all:15", "performance")
	if val != "all:0" {
		t.Fatal(val)
	}
	val = OptCPUVal("energy_perf_bias", "cpu0:15 cpu1:6 cpu2:0", "performance")
	if val != "cpu0:0 cpu1:0 cpu2:0" {
		t.Fatal(val)
	}

	/* future feature
	val = OptCPUVal("energy_perf_bias", "cpu0:6 cpu1:6 cpu2:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:0 cpu1:6 cpu2:15" {
		t.Fatal(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:performance cpu1:normal cpu2:powersave" {
		t.Fatal(val)
	}
	*/

	val = OptCPUVal("governor", "all:powersave", "performance")
	if val != "all:performance" {
		t.Fatal(val)
	}
	val = OptCPUVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "performance")
	if val != "cpu0:performance cpu1:performance cpu2:performance" {
		t.Fatal(val)
	}
	/* future feature
	val = OptCPUVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Fatal(val)
	}
	val = OptCPUVal("energy_perf_bias", "all:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Fatal(val)
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
		t.Fatal(val)
	}
}

func TestOptMemVal(t *testing.T) {
	val := OptMemVal("VSZ_TMPFS_PERCENT", "47", "80", "0", "80")
	if val != "80" {
		t.Fatal(val)
	}
	val = OptMemVal("VSZ_TMPFS_PERCENT", "-1", "75", "0", "75")
	if val != "75" {
		t.Fatal(val)
	}

	size75 := uint64(system.GetTotalMemSizeMB()) * 75 / 100
	size80 := uint64(system.GetTotalMemSizeMB()) * 80 / 100

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "0", "80")
	if val != strconv.FormatUint(size80, 10) {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "0", "80")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "0", "0")
	if val != strconv.FormatUint(size75, 10) {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "0", "0")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "25605", "80")
	if val != "25605" {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "25605", "80")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "25605", "0")
	if val != "25605" {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "25605", "0")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("UNKOWN_PARAMETER", "16043", "0", "0", "0")
	if val != "" {
		t.Fatal(val)
	}
	val = OptMemVal("UNKOWN_PARAMETER", "-1", "0", "0", "0")
	if val != "" {
		t.Fatal(val)
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
		t.Fatal(val)
	}
}

func TestSetRpmVal(t *testing.T) {
	val := SetRpmVal("NO_OPT")
	if val != nil {
		t.Fatal(val)
	}
}

func TestGetGrubVal(t *testing.T) {
	val := GetGrubVal("grub:processor.max_cstate")
	if val == "NA" {
		t.Log("'processor.max_cstate' not found in kernel cmdline")
	}
	val = GetGrubVal("grub:UNKNOWN")
	if val != "NA" {
		t.Fatal(val)
	}
}

func TestOptGrubVal(t *testing.T) {
	val := OptGrubVal("grub:processor.max_cstate", "NO_OPT")
	if val != "NO_OPT" {
		t.Fatal(val)
	}
}

func TestSetGrubVal(t *testing.T) {
	val := SetGrubVal("NO_OPT")
	if val != nil {
		t.Fatal(val)
	}
}

func TestGetServiceVal(t *testing.T) {
	val := GetServiceVal("UnkownService")
	if val != "NA" {
		t.Fatal(val)
	}
	val = GetServiceVal("uuidd.socket")
	if val != "start" && val != "stop" && val != "NA" {
		t.Fatal(val)
	}
}

func TestOptServiceVal(t *testing.T) {
	val := OptServiceVal("UnkownService", "start")
	if val != "NA" {
		t.Fatal(val)
	}
	val = OptServiceVal("uuidd.socket", "start")
	if val != "start" && val != "NA" {
		t.Fatal(val)
	}
	val = OptServiceVal("uuidd.socket", "stop")
	if val != "start" && val != "NA" {
		t.Fatal(val)
	}
	val = OptServiceVal("uuidd.socket", "unknown")
	if val != "start" && val != "NA" {
		t.Fatal(val)
	}
	val = OptServiceVal("sysstat", "start")
	if val != "start" && val != "NA" {
		t.Fatal(val)
	}
	val = OptServiceVal("sysstat.service", "stop")
	if val != "stop" && val != "NA" {
		t.Fatal(val)
	}
	val = OptServiceVal("sysstat", "unknown")
	if val != "start" && val != "NA" {
		t.Fatal(val)
	}
}

func TestSetServiceVal(t *testing.T) {
	val := SetServiceVal("UnkownService", "start")
	if val != nil {
		t.Fatal(val)
	}
}

func TestGetLoginVal(t *testing.T) {
	val, err := GetLoginVal("Unkown")
	if val != "" || err != nil {
		t.Fatal(val)
	}

	val, err = GetLoginVal("UserTasksMax")
	if _, errno := os.Stat("/etc/systemd/logind.conf.d/saptune-UserTasksMax.conf"); errno != nil {
		if !os.IsNotExist(errno) {
			if val != "" || err == nil {
				t.Fatal(val)
			}
		} else {
			if val != "NA" || err != nil {
				t.Fatal(val)
			}
		}
	}
}

func TestOptLoginVal(t *testing.T) {
	val := OptLoginVal("unkown")
	if val != "unkown" {
		t.Fatal(val)
	}
	val = OptLoginVal("infinity")
	if val != "infinity" {
		t.Fatal(val)
	}
	val = OptLoginVal("")
	if val != "" {
		t.Fatal(val)
	}
}

// SetLoginVal

func TestGetPagecacheVal(t *testing.T) {
	prepare := LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val := GetPagecacheVal("ENABLE_PAGECACHE_LIMIT", &prepare)
	if val != "yes" && val != "no" {
		t.Fatal(val)
	}
	if prepare.VMPagecacheLimitMB == 0 && val != "no" {
		t.Fatal(val)
	}
	if prepare.VMPagecacheLimitMB > 0 && val != "yes" {
		t.Fatal(val)
	}

	prepare = LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val = GetPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, &prepare)
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val)
	}

	prepare = LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val = GetPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", &prepare)
	if prepare.VMPagecacheLimitMB == 0 && val != "" {
		t.Fatal(val)
	}
	if prepare.VMPagecacheLimitMB > 0 && val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val)
	}

	prepare = LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
	val = GetPagecacheVal("UNKOWN", &prepare)
	if val != "" {
		t.Fatal(val)
	}
}

func TestOptPagecacheVal(t *testing.T) {
	initPrepare, _ := LinuxPagingImprovements{PagingConfig: PCTestBaseConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare := initPrepare.(LinuxPagingImprovements)

	val := OptPagecacheVal("UNKNOWN", "unknown", &prepare)
	if val != "unknown" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "yes", &prepare)
	if val != "yes" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "no", &prepare)
	if val != "no" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "unknown", &prepare)
	if val != "no" {
		t.Fatal(val)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "2", &prepare)
	if val != "2" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "1", &prepare)
	if val != "1" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "0", &prepare)
	if val != "0" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal(system.SysctlPagecacheLimitIgnoreDirty, "unknown", &prepare)
	if val != "1" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}

	PCTestConf := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest1.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}

	calc := system.GetMainMemSizeMB() * 2 / 100
	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest2.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Fatal(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}
	if val != strconv.FormatUint(calc, 10) {
		t.Fatal(val, calc)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest3.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest4.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Fatal(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}
	if val != strconv.FormatUint(calc, 10) {
		t.Fatal(val, calc)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest5.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest6.ini")
	initPrepare, _ = LinuxPagingImprovements{PagingConfig: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Fatal(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}
	if val != "641" {
		t.Fatal(val)
	}

}

func TestSetPagecacheVal(t *testing.T) {
	prepare := LinuxPagingImprovements{PagingConfig: PCTestBaseConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}
	val := SetPagecacheVal("UNKNOWN", &prepare)
	if val != nil {
		t.Fatal(val)
	}
}
