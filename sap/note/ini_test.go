package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func cleanUp() {
	var parameterStateDir = "/var/lib/saptune/parameter"
	os.RemoveAll(parameterStateDir)
	defer os.RemoveAll(parameterStateDir)
	var saptuneSectionDir = "/var/lib/saptune/sections"
	os.RemoveAll(saptuneSectionDir)
	defer os.RemoveAll(saptuneSectionDir)
}

func TestCalculateOptimumValue(t *testing.T) {
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThan, "21", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThan, "10", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThan, "", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "21", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "20", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "10", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorLessThan, "10", "20"); val != "10" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThan, "10", "10"); val != "9" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThan, "", "10"); val != "9" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "10", "8"); val != "8" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "10", "20"); val != "10" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "10", "10"); val != "10" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "", "10"); val != "10" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorEqual, "21", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorEqual, "10", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorEqual, "", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
}

func TestVendorSettings(t *testing.T) {
	cleanUp()
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_test.ini")
	ini := INISettings{ConfFilePath: iniPath, ID: "471147"}

	if ini.Name() == "" {
		t.Error(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("ini_test: SAP Note file for ini_test\n\t\t\tVersion 2 from 02.11.2017 ") {
		t.Error(ini.Name())
	}

	initialised, err := ini.Initialise()
	if err != nil {
		t.Error(err)
	}
	initialisedINI := initialised.(INISettings)
	for _, key := range []string{"vm.dirty_ratio", "vm.dirty_background_ratio", "vm.swappiness"} {
		if initialisedINI.SysctlParams[key] == "" {
			t.Error(initialisedINI.SysctlParams)
		}
	}

	optimised, err := initialisedINI.Optimise()
	if err != nil {
		t.Error(err)
	}
	optimisedINI := optimised.(INISettings)
	//if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_ratio"], 10, 64); err != nil || i < 11 {
	if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_ratio"], 10, 64); err != nil || i != 10 {
		t.Error(i, err)
	}
	//if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_background_ratio"], 10, 64); err != nil || i > 9 {
	if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_background_ratio"], 10, 64); err != nil || i != 10 {
		t.Error(i, err)
	}
	if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.swappiness"], 10, 64); err != nil || i != 10 {
		t.Error(i, err)
	}

	valApplyList := make([]string, 3)
	valApplyList[0] = "vm.dirty_ratio"
	valApplyList[1] = "vm.dirty_background_ratio"
	valApplyList[2] = "vm.swappiness"
	valapp := optimisedINI.SetValuesToApply(valApplyList)
	if valapp.(INISettings).ValuesToApply["vm.dirty_ratio"] != "vm.dirty_ratio" {
		t.Error(valapp.(INISettings).ValuesToApply["vm.dirty_ratio"])
	}
	if valapp.(INISettings).ValuesToApply["vm.swappiness"] == "vm.dirty_background_ratio" {
		t.Error(valapp.(INISettings).ValuesToApply["vm.swappiness"])
	}
}

func TestNoConfig(t *testing.T) {
	iniPath := "/no_config_file"
	ini := INISettings{ConfFilePath: iniPath, ID: "47114711"}
	initialised, err := ini.Initialise()
	if err == nil {
		t.Error(err)
	}
	initialisedINI := initialised.(INISettings)
	_, err = initialisedINI.Optimise()
	if err == nil {
		t.Error(err)
	}
}

func TestAllSettings(t *testing.T) {
	cleanUp()
	testString := []string{"vm.nr_hugepages", "THP", "KSM", "systemd:sysstat"}
	if runtime.GOARCH == "ppc64le" {
		testString = []string{"KSM", "systemd:sysstat"}
	}
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_all_test.ini")
	ini := INISettings{ConfFilePath: iniPath, ID: "9876543"}
	//t.Logf("ini - %+v\n", ini)

	if ini.Name() == "" {
		t.Error(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("ini_all_test: SAP Note file for ini_all_test\n\t\t\tVersion 3 from 02.01.2019 ") {
		t.Error(ini.Name())
	}

	initialised, err := ini.Initialise()
	//t.Logf("initialised - %+v\n", initialised)
	if err != nil {
		t.Error(err)
	}
	initialisedINI := initialised.(INISettings)
	for _, key := range testString {
		if initialisedINI.SysctlParams[key] == "" {
			t.Error(initialisedINI.SysctlParams)
		}
	}

	optimised, err := initialisedINI.Optimise()
	//t.Logf("optimised - %+v\n", optimised)
	if err != nil {
		t.Error(err)
	}
	optimisedINI := optimised.(INISettings)
	bval := ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["IO_SCHEDULER"]) {
		fields := strings.Split(entry, "@")
		if bval == "" {
			bval = bval + fmt.Sprintf("%s@%s", fields[0], "noop")
		} else {
			bval = bval + " " + fmt.Sprintf("%s@%s", fields[0], "noop")
		}
	}
	if optimisedINI.SysctlParams["IO_SCHEDULER"] != bval {
		t.Error(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["NRREQ"]) {
		fields := strings.Split(entry, "@")
		if bval == "" {
			bval = bval + fmt.Sprintf("%s@%s", fields[0], "1022")
		} else {
			bval = bval + " " + fmt.Sprintf("%s@%s", fields[0], "1022")
		}
	}
	if optimisedINI.SysctlParams["NRREQ"] != bval {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["THP"] != "always" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["KSM"] != "1" {
		t.Error(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["energy_perf_bias"]) {
		fields := strings.Split(entry, ":")
		bval = bval + fmt.Sprintf("%s:%s ", fields[0], "15")
	}
	if optimisedINI.SysctlParams["energy_perf_bias"] != strings.TrimSpace(bval) {
		t.Error(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["governor"]) {
		fields := strings.Split(entry, ":")
		bval = bval + fmt.Sprintf("%s:%s ", fields[0], "performance")
	}
	if optimisedINI.SysctlParams["governor"] != strings.TrimSpace(bval) {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_sybase_hard_memlock"] != "sybase hard memlock 28571380" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_sybase_soft_memlock"] != "sybase soft memlock 28571380" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["ShmFileSystemSizeMB"] != "25605" && optimisedINI.SysctlParams["ShmFileSystemSizeMB"] != "-1" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["VSZ_TMPFS_PERCENT"] != "60" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["systemd:sysstat"] != "stop" && optimisedINI.SysctlParams["systemd:sysstat"] != "NA" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["systemd:uuidd.socket"] != "start" && optimisedINI.SysctlParams["systemd:uuidd.socket"] != "NA" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["systemd:UnkownService"] != "NA" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["grub:transparent_hugepage"] != "never" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["rpm:glibc"] != "2.22-51.6" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["UserTasksMax"] != "setinpostinstall" {
		t.Error(optimisedINI.SysctlParams)
	}
	if runtime.GOARCH != "ppc64le" {
		if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.nr_hugepages"], 10, 64); err != nil || i != 128 {
			t.Error(i, err)
		}
	}
	txt2chk := `# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
`
	if optimisedINI.SysctlParams["reminder"] != txt2chk {
		t.Error(optimisedINI.SysctlParams)
	}

	// apply
	valToApp := map[string]string{"THP": "THP", "KSM": "KSM", "LIMIT_sybase_hard_memlock": "LIMIT_sybase_hard_memlock", "LIMIT_sybase_soft_memlock": "LIMIT_sybase_soft_memlock", "vm.dirty_ratio": "vm.dirty_ratio", "vm.dirty_background_ratio": "vm.dirty_background_ratio", "ShmFileSystemSizeMB": "ShmFileSystemSizeMB", "systemd:sysstat": "systemd:sysstat", "systemd:uuidd.socket": "systemd:uuidd.socket"}
	optimisedINI.ValuesToApply = valToApp
	//t.Logf("optimisedINI - %+v\n", optimisedINI)

	err = optimisedINI.Apply()
	if err != nil {
		t.Error(err)
	}
	appl := INISettings{ConfFilePath: iniPath, ID: "9876543"}
	//t.Logf("appl - %+v\n", appl)
	applied, err := appl.Initialise()
	if err != nil {
		t.Error(err)
	}
	applyINI := applied.(INISettings)
	//t.Logf("applied: %+v,\n optimised: %+v\n", applyINI, optimisedINI)
	if applyINI.SysctlParams["THP"] != "always" {
		t.Error(applyINI.SysctlParams)
	}
	if applyINI.SysctlParams["KSM"] != "1" {
		t.Error(applyINI.SysctlParams)
	}
	if applyINI.SysctlParams["LIMIT_sybase_hard_memlock"] != "sybase hard memlock 28571380" {
		t.Error(applyINI.SysctlParams)
	}
	if applyINI.SysctlParams["LIMIT_sybase_soft_memlock"] != "sybase soft memlock 28571380" {
		t.Error(applyINI.SysctlParams)
	}
	if applyINI.SysctlParams["ShmFileSystemSizeMB"] != "25605" && applyINI.SysctlParams["ShmFileSystemSizeMB"] != "-1" {
		t.Error(applyINI.SysctlParams)
	}

	wrong := false
	for _, sval := range strings.Split(applyINI.SysctlParams["systemd:sysstat"], ",") {
		state := strings.TrimSpace(sval)
		if state != "start" && state != "stop" && state != "NA" && state != "enable" && state != "disable" {
			wrong = true
		}
	}
	if wrong {
		t.Error(applyINI.SysctlParams)
	}
	wrong = false
	for _, sval := range strings.Split(applyINI.SysctlParams["systemd:uuidd.socket"], ",") {
		state := strings.TrimSpace(sval)
		if state != "start" && state != "NA" && state != "enable" && state != "disable" {
			wrong = true
		}
	}
	if wrong {
		t.Error(applyINI.SysctlParams)
	}

	// revert
	valToApp = map[string]string{"revert": "revert"}
	initialisedINI.ValuesToApply = valToApp
	//t.Logf("initialisedINI - %+v\n", initialisedINI)

	err = initialisedINI.Apply()
	if err != nil {
		t.Error(err)
	}
	appl = INISettings{ConfFilePath: iniPath, ID: "9876543"}
	//t.Logf("appl - %+v\n", appl)
	applied, err = appl.Initialise()
	if err != nil {
		t.Error(err)
	}
	applyINI = applied.(INISettings)
	//t.Logf("applied: %+v,\n initialised: %+v\n", applyINI, initialisedINI)
}

func TestOverrideAllSettings(t *testing.T) {
	cleanUp()
	testString := []string{"vm.nr_hugepages", "THP", "KSM", "systemd:sysstat"}
	if runtime.GOARCH == "ppc64le" {
		testString = []string{"KSM", "systemd:sysstat"}
	}
	ovFile := "/etc/saptune/override/9876543"
	srcFile := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/override_ini_all_test.ini")
	_ = system.CopyFile(srcFile, ovFile)

	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_all_test.ini")
	ini := INISettings{ConfFilePath: iniPath, ID: "9876543"}
	t.Log(ini)

	if ini.Name() == "" {
		t.Error(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("ini_all_test: SAP Note file for ini_all_test\n\t\t\tVersion 3 from 02.01.2019 ") {
		t.Error(ini.Name())
	}

	initialised, err := ini.Initialise()
	if err != nil {
		t.Error(err)
	}
	initialisedINI := initialised.(INISettings)
	for _, key := range testString {
		if initialisedINI.SysctlParams[key] == "" {
			t.Error(initialisedINI.SysctlParams)
		}
	}

	optimised, err := initialisedINI.Optimise()
	if err != nil {
		t.Error(err)
	}
	optimisedINI := optimised.(INISettings)
	// clean up
	os.Remove(ovFile)

	bval := ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["IO_SCHEDULER"]) {
		fields := strings.Split(entry, "@")
		if bval == "" {
			bval = bval + fmt.Sprintf("%s@%s", fields[0], "noop")
		} else {
			bval = bval + " " + fmt.Sprintf("%s@%s", fields[0], "noop")
		}
	}
	if optimisedINI.SysctlParams["IO_SCHEDULER"] != bval {
		t.Error(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["NRREQ"]) {
		fields := strings.Split(entry, "@")
		if bval == "" {
			bval = bval + fmt.Sprintf("%s@%s", fields[0], "1024")
		} else {
			bval = bval + " " + fmt.Sprintf("%s@%s", fields[0], "1024")
		}
	}
	if optimisedINI.SysctlParams["NRREQ"] != bval {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["THP"] != "never" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["KSM"] != "0" {
		t.Error(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["energy_perf_bias"]) {
		fields := strings.Split(entry, ":")
		bval = bval + fmt.Sprintf("%s:%s ", fields[0], "0")
	}
	if optimisedINI.SysctlParams["energy_perf_bias"] != strings.TrimSpace(bval) {
		t.Error(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["governor"]) {
		fields := strings.Split(entry, ":")
		bval = bval + fmt.Sprintf("%s:%s ", fields[0], "performance")
	}
	if optimisedINI.SysctlParams["governor"] != strings.TrimSpace(bval) {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_sybase_hard_memlock"] != "sybase hard memlock 571380" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_sybase_soft_memlock"] != "sybase soft memlock 571380" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["ShmFileSystemSizeMB"] != "25605" && optimisedINI.SysctlParams["ShmFileSystemSizeMB"] != "-1" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["VSZ_TMPFS_PERCENT"] != "60" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["systemd:sysstat"] != "start" && optimisedINI.SysctlParams["systemd:sysstat"] != "NA" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["systemd:uuidd.socket"] != "start" && optimisedINI.SysctlParams["systemd:uuidd.socket"] != "NA" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["systemd:UnkownService"] != "NA" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["grub:transparent_hugepage"] != "never" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["rpm:glibc"] != "2.22-51.6" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["UserTasksMax"] != "infinity" {
		t.Error(optimisedINI.SysctlParams)
	}
	if runtime.GOARCH != "ppc64le" {
		if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.nr_hugepages"], 10, 64); err != nil || i != 126 {
			t.Error(i, err)
		}
	}
	txt2chk := `# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
`
	if optimisedINI.SysctlParams["reminder"] != txt2chk {
		t.Error(optimisedINI.SysctlParams)
	}
	// cleanup
	CleanUpRun()
}

func TestPageCacheSettings(t *testing.T) {
	cleanUp()
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest6.ini")
	ini := INISettings{ConfFilePath: iniPath}

	if ini.Name() == "" {
		t.Error(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("Linux paging improvements\n\t\t\tVersion 14 from 10.08.2015 ") {
		t.Error(ini.Name())
	}

	initialised, err := ini.Initialise()
	if err != nil {
		t.Error(err)
	}
	initialisedINI := initialised.(INISettings)

	optimised, err := initialisedINI.Optimise()
	if err != nil {
		t.Error(err)
	}
	optimisedINI := optimised.(INISettings)
	if optimisedINI.SysctlParams["ENABLE_PAGECACHE_LIMIT"] != "yes" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams[system.SysctlPagecacheLimitIgnoreDirty] != "1" {
		t.Error(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["OVERRIDE_PAGECACHE_LIMIT_MB"] != "641" {
		t.Error(optimisedINI.SysctlParams)
	}
	cleanUp()
}
