package note

import (
	"fmt"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"strconv"
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
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/sap/note/ini_test.ini")
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
