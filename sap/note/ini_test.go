package note

import (
	"fmt"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestCalculateOptimumValue(t *testing.T) {
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThan, "21", "20"); val != "21" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThan, "10", "20"); val != "21" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThan, "", "20"); val != "21" || err != nil {
		t.Fatal(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "21", "20"); val != "21" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "20", "20"); val != "20" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "10", "20"); val != "20" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorMoreThanEqual, "", "20"); val != "20" || err != nil {
		t.Fatal(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorLessThan, "10", "20"); val != "10" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThan, "10", "10"); val != "9" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThan, "", "10"); val != "9" || err != nil {
		t.Fatal(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "10", "8"); val != "8" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "10", "20"); val != "10" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "10", "10"); val != "10" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorLessThanEqual, "", "10"); val != "10" || err != nil {
		t.Fatal(val, err)
	}

	if val, err := CalculateOptimumValue(txtparser.OperatorEqual, "21", "20"); val != "20" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorEqual, "10", "20"); val != "20" || err != nil {
		t.Fatal(val, err)
	}
	if val, err := CalculateOptimumValue(txtparser.OperatorEqual, "", "20"); val != "20" || err != nil {
		t.Fatal(val, err)
	}
}

func TestVendorSettings(t *testing.T) {
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_test.ini")
	ini := INISettings{ConfFilePath: iniPath}

	if ini.Name() == "" {
		t.Fatal(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("ini_test: SAP Note file for ini_test\n\t\t\tVersion 2 from 02.11.2017 ") {
		t.Fatal(ini.Name())
	}

	initialised, err := ini.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	initialisedINI := initialised.(INISettings)
	for _, key := range []string{"vm.dirty_ratio", "vm.dirty_background_ratio", "vm.swappiness"} {
		if initialisedINI.SysctlParams[key] == "" {
			t.Fatal(initialisedINI.SysctlParams)
		}
	}

	optimised, err := initialisedINI.Optimise()
	if err != nil {
		t.Fatal(err)
	}
	optimisedINI := optimised.(INISettings)
	//if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_ratio"], 10, 64); err != nil || i < 11 {
	if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_ratio"], 10, 64); err != nil || i != 10 {
		t.Fatal(i, err)
	}
	//if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_background_ratio"], 10, 64); err != nil || i > 9 {
	if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.dirty_background_ratio"], 10, 64); err != nil || i != 10 {
		t.Fatal(i, err)
	}
	if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.swappiness"], 10, 64); err != nil || i != 10 {
		t.Fatal(i, err)
	}

	valApplyList := make([]string, 3)
	valApplyList[0] = "vm.dirty_ratio"
	valApplyList[1] = "vm.dirty_background_ratio"
	valApplyList[2] = "vm.swappiness"
	valapp := optimisedINI.SetValuesToApply(valApplyList)
	if valapp.(INISettings).ValuesToApply["vm.dirty_ratio"] != "vm.dirty_ratio" {
		t.Fatal(valapp.(INISettings).ValuesToApply["vm.dirty_ratio"])
	}
	if valapp.(INISettings).ValuesToApply["vm.swappiness"] == "vm.dirty_background_ratio" {
		t.Fatal(valapp.(INISettings).ValuesToApply["vm.swappiness"])
	}
}

func TestAllSettings(t *testing.T) {
	testString := []string{"vm.nr_hugepages", "THP", "KSM", "Sysstat"}
	if runtime.GOARCH == "ppc64le" {
		testString = []string{"KSM", "Sysstat"}
	}
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_all_test.ini")
	ini := INISettings{ConfFilePath: iniPath}

	if ini.Name() == "" {
		t.Fatal(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("ini_all_test: SAP Note file for ini_all_test\n\t\t\tVersion 3 from 02.01.2019 ") {
		t.Fatal(ini.Name())
	}

	initialised, err := ini.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	initialisedINI := initialised.(INISettings)
	for _, key := range testString {
		if initialisedINI.SysctlParams[key] == "" {
			t.Fatal(initialisedINI.SysctlParams)
		}
	}

	optimised, err := initialisedINI.Optimise()
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(optimisedINI.SysctlParams)
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
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["THP"] != "always" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["KSM"] != "1" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["energy_perf_bias"]) {
		fields := strings.Split(entry, ":")
		bval = bval + fmt.Sprintf("%s:%s ", fields[0], "15")
	}
	if optimisedINI.SysctlParams["energy_perf_bias"] != strings.TrimSpace(bval) {
		t.Fatal(optimisedINI.SysctlParams)
	}
	bval = ""
	for _, entry := range strings.Fields(initialisedINI.SysctlParams["governor"]) {
		fields := strings.Split(entry, ":")
		bval = bval + fmt.Sprintf("%s:%s ", fields[0], "performance")
	}
	if optimisedINI.SysctlParams["governor"] != strings.TrimSpace(bval) {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_SOFT"] != "sybase:28571380 " {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_HARD"] != "sybase:28571380 " {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_DOMAIN"] != "sybase " {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["LIMIT_ITEM"] != "sybase:memlock " {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["ShmFileSystemSizeMB"] != "25605" && optimisedINI.SysctlParams["ShmFileSystemSizeMB"] != "-1" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["VSZ_TMPFS_PERCENT"] != "60" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["Sysstat"] != "stop" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["UuiddSocket"] != "start" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["UnkownService"] != "" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["grub:transparent_hugepage"] != "never" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["rpm:glibc"] != "2.22-51.6" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["UserTasksMax"] != "setinpostinstall" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if runtime.GOARCH != "ppc64le" {
		if i, err := strconv.ParseInt(optimisedINI.SysctlParams["vm.nr_hugepages"], 10, 64); err != nil || i != 128 {
			t.Fatal(i, err)
		}
	}
	txt2chk := `# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
`
	if optimisedINI.SysctlParams["reminder"] != txt2chk {
		t.Fatal(optimisedINI.SysctlParams)
	}
}

func TestPageCacheSettings(t *testing.T) {
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest6/usr/share/saptune/notes/1557506")
	ini := INISettings{ConfFilePath: iniPath}

	if ini.Name() == "" {
		t.Fatal(ini.Name())
	}
	if ini.Name() != fmt.Sprintf("Linux paging improvements\n\t\t\tVersion 14 from 10.08.2015 ") {
		t.Fatal(ini.Name())
	}

	initialised, err := ini.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	initialisedINI := initialised.(INISettings)

	optimised, err := initialisedINI.Optimise()
	if err != nil {
		t.Fatal(err)
	}
	optimisedINI := optimised.(INISettings)
	if optimisedINI.SysctlParams["ENABLE_PAGECACHE_LIMIT"] != "yes" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["PAGECACHE_LIMIT_IGNORE_DIRTY"] != "1" {
		t.Fatal(optimisedINI.SysctlParams)
	}
	if optimisedINI.SysctlParams["OVERRIDE_PAGECACHE_LIMIT_MB"] != "641" {
		t.Fatal(optimisedINI.SysctlParams)
	}
}
