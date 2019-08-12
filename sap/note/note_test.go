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
