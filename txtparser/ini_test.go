package txtparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var fileNotExist = "/file_does_not_exist"
var tstFile = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_all_test.ini")
var tst2File = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/wrong_limit_test.ini")
var fileName = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes/1557506")
var descName = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s", "Linux paging improvements", "", "14", "10.08.2015 ")
var category = "LINUX"
var fileVersion = "14"

var iniExample = `
# comment
[Section A]
alpha.beta-charlie_delta < 1	a
echo.foxtrot > 2	bb

[Section B]
golf-hotel = 3	ccc
india_julia < 4	dddd

[Section C]

[Section D]
lima > 5	eeeee

[Section E]
mike.november+oscar_papa-quebeck > 6	ffffff
`

// iniExample parsed and serialised into JSON
var iniJSON = `
{
	"AllValues": [{
		"Section": "Section A",
		"Key": "alpha.beta-charlie_delta",
		"Operator": "\u003c",
		"Value": "1\ta"
	}, {
		"Section": "Section A",
		"Key": "echo.foxtrot",
		"Operator": "\u003e",
		"Value": "2\tbb"
	}, {
		"Section": "Section B",
		"Key": "golf-hotel",
		"Operator": "=",
		"Value": "3\tccc"
	}, {
		"Section": "Section B",
		"Key": "india_julia",
		"Operator": "\u003c",
		"Value": "4\tdddd"
	}, {
		"Section": "Section D",
		"Key": "lima",
		"Operator": "\u003e",
		"Value": "5\teeeee"
	}, {
		"Section": "Section E",
		"Key": "mike.november+oscar_papa-quebeck",
		"Operator": "\u003e",
		"Value": "6\tffffff"
	}],
	"KeyValue": {
		"Section A": {
			"alpha.beta-charlie_delta": {
				"Section": "Section A",
				"Key": "alpha.beta-charlie_delta",
				"Operator": "\u003c",
				"Value": "1\ta"
			},
			"echo.foxtrot": {
				"Section": "Section A",
				"Key": "echo.foxtrot",
				"Operator": "\u003e",
				"Value": "2\tbb"
			}
		},
		"Section B": {
			"golf-hotel": {
				"Section": "Section B",
				"Key": "golf-hotel",
				"Operator": "=",
				"Value": "3\tccc"
			},
			"india_julia": {
				"Section": "Section B",
				"Key": "india_julia",
				"Operator": "\u003c",
				"Value": "4\tdddd"
			}
		},
		"Section C": {},
		"Section D": {
			"lima": {
				"Section": "Section D",
				"Key": "lima",
				"Operator": "\u003e",
				"Value": "5\teeeee"
			}
		},
		"Section E": {
			"mike.november+oscar_papa-quebeck": {
				"Section": "Section E",
				"Key": "mike.november+oscar_papa-quebeck",
				"Operator": "\u003e",
				"Value": "6\tffffff"
			}
		}
	}
}`

var iniWrongJSON = `
{
	"AllValues": [{
		"Section": "limits",
		"Key": "limits_NA",
		"Operator": "=",
		"Value": "NA"
	}, {
		"Section": "reminder",
		"Key": "reminder",
		"Operator": "",
		"Value": "# Text to ignore for apply but to display.\n# Everything the customer should know about this note, especially\n# which parameters are NOT handled and the reason.\n"
	}],
	"KeyValue": {
		"limits": {
			"limits_NA": {
				"Section": "limits",
				"Key": "limits_NA",
				"Operator": "=",
				"Value": "NA"
			}
		},
		"reminder": {
			"reminder": {
				"Section": "reminder",
				"Key": "reminder",
				"Operator": "",
				"Value": "# Text to ignore for apply but to display.\n# Everything the customer should know about this note, especially\n# which parameters are NOT handled and the reason.\n"
			}
		}
	}
}`

func TestParseINIFile(t *testing.T) {
	content, err := ParseINIFile(fileName, false)
	if err != nil {
		t.Fatal(content, err)
	}
	newFile := path.Join(os.TempDir(), "saptunetest1")
	content, err = ParseINIFile(newFile, true)
	if err != nil {
		t.Fatal(content, err)
	}
	if _, err = os.Stat(newFile); err != nil {
		t.Fatalf("file '%s' does not exist\n", newFile)
	}
	os.Remove(newFile)
	newFile2 := path.Join(os.TempDir(), "saptunetest2")
	content, err = ParseINIFile(newFile2, false)
	if err == nil {
		t.Fatal(content, err)
	}
	if _, err = os.Stat(newFile); err == nil {
		os.Remove(newFile2)
		t.Fatalf("file '%s' exists\n", newFile2)
	}
}

func TestParseINI(t *testing.T) {
	actualINI := ParseINI(iniExample)
	var expectedINI INIFile
	if err := json.Unmarshal([]byte(iniJSON), &expectedINI); err != nil {
		t.Fatal(err)
	}
	//b, err := json.Marshal(actualINI)
	//t.Log(string(b), err)
	if !reflect.DeepEqual(*actualINI, expectedINI) {
		t.Fatalf("\n%+v\n%+v\n", *actualINI, expectedINI)
	}
	content, err := ioutil.ReadFile(tstFile)
	if err != nil {
		t.Fatal(err)
	}
	newINI := ParseINI(string(content))

	content, err = ioutil.ReadFile(tst2File)
	if err != nil {
		t.Fatal(err)
	}
	newINI = ParseINI(string(content))
	var wrongINI INIFile
	if err := json.Unmarshal([]byte(iniWrongJSON), &wrongINI); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*newINI, wrongINI) {
		t.Fatalf("\n%+v\n%+v\n", *newINI, wrongINI)
	}
}

func TestGetINIFileDescriptiveName(t *testing.T) {
	str := GetINIFileDescriptiveName(fileName)
	if str != descName {
		t.Fatalf("\n'%+v'\nis not\n'%+v'\n", str, descName)
	}
	str = GetINIFileDescriptiveName(fileNotExist)
	if str != "" {
		t.Fatalf(str)
	}
}

func TestGetINIFileVersionSectionEntry(t *testing.T) {
	str := GetINIFileVersionSectionEntry(fileName, "category")
	if str != category {
		t.Fatalf("\n'%+v'\nis not\n'%+v'\n", str, category)
	}
	str = GetINIFileVersionSectionEntry(fileNotExist, "category")
	if str != "" {
		t.Fatalf(str)
	}
	str = GetINIFileVersionSectionEntry(fileName, "version")
	if str != fileVersion {
		t.Fatalf("\n'%+v'\nis not\n'%+v'\n", str, fileVersion)
	}
	str = GetINIFileVersionSectionEntry(fileNotExist, "version")
	if str != "" {
		t.Fatalf(str)
	}
	str = GetINIFileVersionSectionEntry(fileName, "not_avail")
	if str != "" {
		t.Fatalf("\n'%+v'\nis not\n'%+v'\n", str, "")
	}
}
