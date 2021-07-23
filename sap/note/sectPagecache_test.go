package note

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"strconv"
	"testing"
)

var PCTestBaseConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/note/1557506")

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
	val = GetPagecacheVal("UNKNOWN", &prepare)
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
