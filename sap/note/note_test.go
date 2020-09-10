package note

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"testing"
)

var OSNotesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes")
var OSPackageInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/")
var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra")

func jsonMarshalAndBack(original interface{}, receiver interface{}, t *testing.T) {
	serialised, err := json.Marshal(original)
	if err != nil {
		t.Fatal(original, err)
	}
	json.Unmarshal(serialised, &receiver)
}

func TestNoteSerialisation(t *testing.T) {
	// All notes must be tested
	paging := LinuxPagingImprovements{VMPagecacheLimitMB: 1000, VMPagecacheLimitIgnoreDirty: 2, UseAlgorithmForHANA: true}
	newPaging := LinuxPagingImprovements{}
	jsonMarshalAndBack(paging, &newPaging, t)
	if eq, diff, valapply := CompareNoteFields(paging, newPaging); !eq {
		t.Fatal(diff, valapply)
	}

	sysctl := INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "300", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	newSysctl := INISettings{}
	jsonMarshalAndBack(sysctl, &newSysctl, t)
	if eq, diff, valapply := CompareNoteFields(sysctl, newSysctl); !eq {
		t.Fatal(diff, valapply)
	}

	sysctl = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "300", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	newSysctl = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "150", "net.ipv4.tcp_keepalive_intvl": "175", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	if eq, diff, valapply := CompareNoteFields(sysctl, newSysctl); eq {
		t.Fatal(diff, valapply)
	}

	sysctl = INISettings{ConfFilePath: path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/fl_test.ini"), SysctlParams: map[string]string{"force_latency": "70", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	newSysctl = INISettings{}
	jsonMarshalAndBack(sysctl, &newSysctl, t)
	if eq, diff, valapply := CompareNoteFields(sysctl, newSysctl); !eq {
		t.Fatal(diff, valapply)
	}
}

func TestCmpMapValue(t *testing.T) {
	var key reflect.Value
	actualNote := INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "300", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	expectedNote := INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "150", "net.ipv4.tcp_keepalive_intvl": "175", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}

	for _, key = range reflect.ValueOf(actualNote).Field(3).MapKeys() {
		if key.String() == "net.ipv4.tcp_keepalive_time" {
			break
		}
	}
	actualValue := reflect.ValueOf(actualNote).Field(3).MapIndex(key).Interface()
	expectedValue := reflect.ValueOf(expectedNote).Field(3).MapIndex(key).Interface()
	expectedComparison := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "net.ipv4.tcp_keepalive_time", ActualValue: "300", ExpectedValue: "150", ActualValueJS: "300", ExpectedValueJS: "150", MatchExpectation: false}

	comparisons := cmpMapValue("SysctlParams", key, actualValue, expectedValue)
	if comparisons != expectedComparison {
		t.Error(comparisons, expectedComparison)
	}

	actualNote = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "4711"), ID: "4711", DescriptiveName: "", SysctlParams: map[string]string{"force_latency": "120", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	expectedNote = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "4711"), ID: "4711", DescriptiveName: "", SysctlParams: map[string]string{"force_latency": "70", "net.ipv4.tcp_keepalive_intvl": "175", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}

	for _, key = range reflect.ValueOf(actualNote).Field(3).MapKeys() {
		if key.String() == "force_latency" {
			break
		}
	}
	actualValue = reflect.ValueOf(actualNote).Field(3).MapIndex(key).Interface()
	expectedValue = reflect.ValueOf(expectedNote).Field(3).MapIndex(key).Interface()
	expectedComparison = FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "force_latency", ActualValue: "120", ExpectedValue: "70", ActualValueJS: "120", ExpectedValueJS: "70", MatchExpectation: false}

	comparisons = cmpMapValue("SysctlParams", key, actualValue, expectedValue)
	if comparisons != expectedComparison {
		t.Error(comparisons, expectedComparison)
	}

	actualNote = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "4711"), ID: "4711", DescriptiveName: "", SysctlParams: map[string]string{"force_latency": "all:none", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	expectedNote = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "4711"), ID: "4711", DescriptiveName: "", SysctlParams: map[string]string{"force_latency": "70", "net.ipv4.tcp_keepalive_intvl": "175", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}

	for _, key = range reflect.ValueOf(actualNote).Field(3).MapKeys() {
		if key.String() == "force_latency" {
			break
		}
	}
	actualValue = reflect.ValueOf(actualNote).Field(3).MapIndex(key).Interface()
	expectedValue = reflect.ValueOf(expectedNote).Field(3).MapIndex(key).Interface()
	expectedComparison = FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "force_latency", ActualValue: "all:none", ExpectedValue: "70", ActualValueJS: "all:none", ExpectedValueJS: "70", MatchExpectation: false}

	comparisons = cmpMapValue("SysctlParams", key, actualValue, expectedValue)
	if comparisons != expectedComparison {
		t.Error(comparisons, expectedComparison)
	}

	actualNote = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "123456"), ID: "123456", DescriptiveName: "", SysctlParams: map[string]string{"rpm:libopenssl1_0_0": "1.0.2p-2.11", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	expectedNote = INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "123456"), ID: "123456", DescriptiveName: "", SysctlParams: map[string]string{"rpm:libopenssl1_0_0": "1.0.2n-3.3.1", "net.ipv4.tcp_keepalive_intvl": "175", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}

	for _, key = range reflect.ValueOf(actualNote).Field(3).MapKeys() {
		if key.String() == "rpm:libopenssl1_0_0" {
			break
		}
	}
	actualValue = reflect.ValueOf(actualNote).Field(3).MapIndex(key).Interface()
	expectedValue = reflect.ValueOf(expectedNote).Field(3).MapIndex(key).Interface()
	expectedComparison = FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "rpm:libopenssl1_0_0", ActualValue: "1.0.2p-2.11", ExpectedValue: "1.0.2n-3.3.1", ActualValueJS: "1.0.2p-2.11", ExpectedValueJS: "1.0.2n-3.3.1", MatchExpectation: true}

	comparisons = cmpMapValue("SysctlParams", key, actualValue, expectedValue)
	if comparisons != expectedComparison {
		t.Error(comparisons, expectedComparison)
	}
}

func TestCmpFieldValue(t *testing.T) {
	actualNote := INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "300", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	expectedNote := INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "150", "net.ipv4.tcp_keepalive_intvl": "175", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	expectedComparison := FieldComparison{ReflectFieldName: "ID", ReflectMapKey: "", ActualValue: "1410736", ExpectedValue: "1410736", ActualValueJS: "1410736", ExpectedValueJS: "1410736", MatchExpectation: true}

	comparisons := cmpFieldValue(1, "ID", reflect.ValueOf(actualNote), reflect.ValueOf(expectedNote))
	if comparisons != expectedComparison {
		t.Error(comparisons, expectedComparison)
	}
}

func TestChkGrubCompliance(t *testing.T) {
	// grub:numa_balancing - false
	comp1 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:numa_balancing", ActualValue: "NA", ExpectedValue: "disable", ActualValueJS: "NA", ExpectedValueJS: "disable", MatchExpectation: false}
	// kernel.numa_balancing - true
	comp2 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "kernel.numa_balancing", ActualValue: "0", ExpectedValue: "0", ActualValueJS: "0", ExpectedValueJS: "0", MatchExpectation: true}
	// kernel.numa_balancing - false
	comp3 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "kernel.numa_balancing", ActualValue: "0", ExpectedValue: "0", ActualValueJS: "0", ExpectedValueJS: "0", MatchExpectation: false}
	// grub:transparent_hugepage - false
	comp4 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:transparent_hugepage", ActualValue: "NA", ExpectedValue: "never", ActualValueJS: "NA", ExpectedValueJS: "never", MatchExpectation: false}
	// THP - true
	comp5 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "THP", ActualValue: "always", ExpectedValue: "never", ActualValueJS: "always", ExpectedValueJS: "never", MatchExpectation: true}
	// THP - false
	comp6 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "THP", ActualValue: "always", ExpectedValue: "never", ActualValueJS: "always", ExpectedValueJS: "never", MatchExpectation: false}
	// grub:intel_idle.max_cstate - false
	comp7 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:intel_idle.max_cstate", ActualValue: "NA", ExpectedValue: "1", ActualValueJS: "NA", ExpectedValueJS: "1", MatchExpectation: false}
	// grub:intel_idle.max_cstate - true
	comp10 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:intel_idle.max_cstate", ActualValue: "NA", ExpectedValue: "1", ActualValueJS: "NA", ExpectedValueJS: "1", MatchExpectation: true}
	// grub:processor.max_cstate - true
	comp8 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:processor.max_cstate", ActualValue: "NA", ExpectedValue: "1", ActualValueJS: "NA", ExpectedValueJS: "1", MatchExpectation: true}
	// grub:processor.max_cstate - false
	comp11 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "grub:processor.max_cstate", ActualValue: "NA", ExpectedValue: "1", ActualValueJS: "NA", ExpectedValueJS: "1", MatchExpectation: false}
	// force_latency - false and all:none
	comp9 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "force_latency", ActualValue: "all:none", ExpectedValue: "70", ActualValueJS: "all:none", ExpectedValueJS: "70", MatchExpectation: false}
	// force_latency - true and all:none
	comp12 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "force_latency", ActualValue: "all:none", ExpectedValue: "70", ActualValueJS: "all:none", ExpectedValueJS: "70", MatchExpectation: true}
	// force_latency - false and !all:none
	comp13 := FieldComparison{ReflectFieldName: "SysctlParams", ReflectMapKey: "force_latency", ActualValue: "33", ExpectedValue: "70", ActualValueJS: "33", ExpectedValueJS: "70", MatchExpectation: false}

	// grub:numa_balancing - false, kernel.numa_balancing - true = true
	t.Run("grub:numa_balancing - false, kernel.numa_balancing - true", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:numa_balancing]"] = comp1
		comparison["SysctlParams[kernel.numa_balancing]"] = comp2

		allMatch = chkGrubCompliance(comparison, allMatch)
		if !allMatch {
			t.Errorf("grub:numa_balancing = false and kernel.numa_balancing = true should be true and NOT false")
		}
	})

	// grub:numa_balancing - false, kernel.numa_balancing - false = false
	t.Run("grub:numa_balancing - false, kernel.numa_balancing - false", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:numa_balancing]"] = comp1
		comparison["SysctlParams[kernel.numa_balancing]"] = comp3

		allMatch = chkGrubCompliance(comparison, allMatch)
		if allMatch {
			t.Errorf("grub:numa_balancing = false and kernel.numa_balancing = false should be false and NOT true")
		}
	})

	// grub:transparent_hugepage - false, THP - true = true
	t.Run("grub:transparent_hugepage - false, THP - true", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:transparent_hugepage]"] = comp4
		comparison["SysctlParams[THP]"] = comp5

		allMatch = chkGrubCompliance(comparison, allMatch)
		if !allMatch {
			t.Errorf("grub:transparent_hugepage = false and THP = true should be true and NOT false")
		}
	})

	// grub:transparent_hugepage - false, THP - false = false
	t.Run("grub:transparent_hugepage - false, THP - false", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:transparent_hugepage]"] = comp4
		comparison["SysctlParams[THP]"] = comp6

		allMatch = chkGrubCompliance(comparison, allMatch)
		if allMatch {
			t.Errorf("grub:transparent_hugepage = false and THP = false should be false and NOT true")
		}
	})

	// grub:intel_idle.max_cstate - false, grub:processor.max_cstate - true , force_latency - false and =all:none  = true
	t.Run("grub:intel_idle.max_cstate - false, grub:processor.max_cstate - true , force_latency - false and =all:none", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:intel_idle.max_cstate]"] = comp7
		comparison["SysctlParams[grub:processor.max_cstate]"] = comp8
		comparison["SysctlParams[force_latency]"] = comp9

		allMatch = chkGrubCompliance(comparison, allMatch)
		if !allMatch {
			t.Errorf("grub:intel_idle.max_cstate = false, grub:processor.max_cstate = true, force_latency = false and force_latency.act_val = all:none should be true and NOT false")
		}
	})

	// grub:intel_idle.max_cstate - true, grub:processor.max_cstate - false , force_latency - false and =all:none  = true
	t.Run("grub:intel_idle.max_cstate - true, grub:processor.max_cstate - false , force_latency - false and =all:none", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:intel_idle.max_cstate]"] = comp10
		comparison["SysctlParams[grub:processor.max_cstate]"] = comp11
		comparison["SysctlParams[force_latency]"] = comp9

		allMatch = chkGrubCompliance(comparison, allMatch)
		if !allMatch {
			t.Errorf("grub:intel_idle.max_cstate = true, grub:processor.max_cstate = false, force_latency = false and force_latency.act_val = all:none should be true and NOT false")
		}
	})

	// grub:intel_idle.max_cstate - false, grub:processor.max_cstate - false , force_latency - true and =all:none  = true
	t.Run("grub:intel_idle.max_cstate - false, grub:processor.max_cstate - false , force_latency - true and =all:none", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:intel_idle.max_cstate]"] = comp7
		comparison["SysctlParams[grub:processor.max_cstate]"] = comp11
		comparison["SysctlParams[force_latency]"] = comp12

		allMatch = chkGrubCompliance(comparison, allMatch)
		if !allMatch {
			t.Errorf("grub:intel_idle.max_cstate = false, grub:processor.max_cstate = false, force_latency = true and force_latency.act_val = all:none should be true and NOT false")
		}
	})

	// grub:intel_idle.max_cstate - true, grub:processor.max_cstate - false , force_latency - false and !=all:none  = false
	t.Run("grub:intel_idle.max_cstate - true, grub:processor.max_cstate - false , force_latency - false and !=all:none", func(t *testing.T) {
		allMatch := true
		comparison := make(map[string]FieldComparison)
		comparison["SysctlParams[grub:intel_idle.max_cstate]"] = comp10
		comparison["SysctlParams[grub:processor.max_cstate]"] = comp11
		comparison["SysctlParams[force_latency]"] = comp13

		allMatch = chkGrubCompliance(comparison, allMatch)
		if allMatch {
			t.Errorf("grub:intel_idle.max_cstate = true, grub:processor.max_cstate = false, force_latency = false and force_latency.act_val != all:none should be false and NOT true")
		}
	})
}

func TestGetTuningOptions(t *testing.T) {
	allOpts := GetTuningOptions(OSNotesInGOPATH, "")
	if sorted := allOpts.GetSortedIDs(); len(allOpts) != len(sorted) {
		t.Fatal(sorted, allOpts)
	}
	allOpts = GetTuningOptions("", TstFilesInGOPATH)
	if sorted := allOpts.GetSortedIDs(); len(allOpts) != len(sorted) {
		t.Fatal(sorted, allOpts)
	}
}

func TestCompareJSValu(t *testing.T) {
	op := ""
	v1 := "tst_string"
	v2 := "tst_string"
	v1i := "1"
	v2i := "1"
	r1, r2, match := CompareJSValue(v1, v2, op)
	if !match {
		t.Fatal(r1, v1, r2, v2, match)
	}
	r1, r2, match = CompareJSValue(v1, v2i, op)
	if match {
		t.Fatal(r1, v1, r2, v2i, match)
	}
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if !match {
		t.Fatal(r1, v1i, r2, v2i, match)
	}
	v1 = "newtst_string"
	r1, r2, match = CompareJSValue(v1, v2, op)
	if match {
		t.Fatal(r1, v1, r2, v2, match)
	}
	v1i = "2"
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if match {
		t.Fatal(r1, v1i, r2, v2i, match)
	}

	op = "=="
	v1 = "tst_string"
	v1i = "1"
	r1, r2, match = CompareJSValue(v1, v2, op)
	if !match {
		t.Fatal(r1, v1, r2, v2, match)
	}
	r1, r2, match = CompareJSValue(v1, v2i, op)
	if match {
		t.Fatal(r1, v1, r2, v2i, match)
	}
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if !match {
		t.Fatal(r1, v1i, r2, v2i, match)
	}
	v1 = "newtst_string"
	r1, r2, match = CompareJSValue(v1, v2, op)
	if match {
		t.Fatal(r1, v1, r2, v2, match)
	}
	v1i = "2"
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if match {
		t.Fatal(r1, v1i, r2, v2i, match)
	}

	// if op="<=" or op=">="
	// compare 'normal' strings will give unpredictable results
	// so no tests with strings like 'tst_value'.
	// calling functions will ensure, that v1 and v2 are strings
	// representing integer values
	op = "<="
	v1i = "1"
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if !match {
		t.Fatalf("compare '%+v' and '%+v', return '%s' and '%s', match: '%+v'\n", v1i, v2i, r1, r2, match)
	}
	v1i = "2"
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if match {
		t.Fatalf("compare '%+v' and '%+v', return '%s' and '%s', match: '%+v'\n", v1i, v2i, r1, r2, match)
	}
	r1, r2, match = CompareJSValue(v2i, v1i, op)
	if !match {
		t.Fatalf("compare '%+v' and '%+v', return '%s' and '%s', match: '%+v'\n", v1i, v2i, r1, r2, match)
	}

	op = ">="
	v1i = "1"
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if !match {
		t.Fatalf("compare '%+v' and '%+v', return '%s' and '%s', match: '%+v'\n", v1i, v2i, r1, r2, match)
	}
	v1i = "2"
	r1, r2, match = CompareJSValue(v1i, v2i, op)
	if !match {
		t.Fatalf("compare '%+v' and '%+v', return '%s' and '%s', match: '%+v'\n", v1i, v2i, r1, r2, match)
	}
	r1, r2, match = CompareJSValue(v2i, v1i, op)
	if match {
		t.Fatalf("compare '%+v' and '%+v', return '%s' and '%s', match: '%+v'\n", v1i, v2i, r1, r2, match)
	}
}
