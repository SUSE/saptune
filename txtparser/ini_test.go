package txtparser

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

var fileNotExist = "/file_does_not_exist"
var tstFile = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_all_test.ini")
var tst2File = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/wrong_limit_test.ini")
var fileNameOld = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_test.ini")
var fileNameNew = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_new_test.ini")
var fileNameWrong = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_wrong_test.ini")
var fileNameMissing = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_missing_test.ini")
var fileName = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes/1557506")
var descName = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s\n\t\t\t%s", "Linux paging improvements", "", "16", "06.02.2020", "https://launchpad.support.sap.com/#/notes/1557506")
var descNameNew = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s", "Linux paging improvements", "", "16", "06.02.2020")
var noteVersion = "16"
var noteDate = "06.02.2020"
var noteTitle = "Linux paging improvements"
var noteRefs = "https://launchpad.support.sap.com/#/notes/1557506"
var oldDescName = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s", "ini_test: SAP Note file for ini_test", "", "2", "02.11.2017")
var oldNoteVersion = "2"
var oldNoteDate = "02.11.2017"
var oldNoteTitle = "ini_test: SAP Note file for ini_test"
var oldNoteCategory = "Linux"

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
		t.Error(content, err)
	}
	newFile := path.Join(os.TempDir(), "saptunetest1")
	content, err = ParseINIFile(newFile, true)
	if err != nil {
		t.Error(content, err)
	}
	if _, err = os.Stat(newFile); err != nil {
		t.Errorf("file '%s' does not exist\n", newFile)
	}
	os.Remove(newFile)
	newFile2 := path.Join(os.TempDir(), "saptunetest2")
	content, err = ParseINIFile(newFile2, false)
	if err == nil {
		t.Error(content, err)
	}
	if _, err = os.Stat(newFile); err == nil {
		os.Remove(newFile2)
		t.Errorf("file '%s' exists\n", newFile2)
	}
}

func TestParseINI(t *testing.T) {
	actualINI := ParseINI(iniExample)
	var expectedINI INIFile
	if err := json.Unmarshal([]byte(iniJSON), &expectedINI); err != nil {
		t.Error(err)
	}
	//b, err := json.Marshal(actualINI)
	//t.Log(string(b), err)
	if !reflect.DeepEqual(*actualINI, expectedINI) {
		t.Errorf("\n%+v\n%+v\n", *actualINI, expectedINI)
	}
	content, err := os.ReadFile(tstFile)
	if err != nil {
		t.Error(err)
	}
	_ = ParseINI(string(content))

	content, err = os.ReadFile(tst2File)
	if err != nil {
		t.Error(err)
	}
	newINI := ParseINI(string(content))
	var wrongINI INIFile
	if err := json.Unmarshal([]byte(iniWrongJSON), &wrongINI); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*newINI, wrongINI) {
		t.Errorf("\n%+v\n%+v\n", *newINI, wrongINI)
	}
}

func TestGetINIFileDescriptiveName(t *testing.T) {
	str := GetINIFileDescriptiveName(fileName)
	if str != descName {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, descName)
	}
	str = GetINIFileDescriptiveName(fileNameNew)
	if str != descNameNew {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, descNameNew)
	}
	str = GetINIFileDescriptiveName(fileNameOld)
	if str != oldDescName {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, oldDescName)
	}
	str = GetINIFileDescriptiveName(fileNotExist)
	if str != "" {
		t.Errorf(str)
	}
}

func TestGetINIFileVersionSectionRefs(t *testing.T) {
	refs := GetINIFileVersionSectionRefs(fileName)
	for _, ref := range refs {
		if ref != noteRefs {
			t.Errorf("\n'%+v'\nis not\n'%+v'\n", ref, noteRefs)
		}
	}
	refs = GetINIFileVersionSectionRefs(fileNameNew)
	if len(refs) > 0 {
		t.Errorf("refs contain '%+v'\n", refs)
	}
}

func TestGetINIFileVersionSectionEntry(t *testing.T) {
	str := GetINIFileVersionSectionEntry(fileName, "reference")
	if str != noteRefs {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, noteRefs)
	}
	str = GetINIFileVersionSectionEntry(fileNotExist, "reference")
	if str != "" {
		t.Errorf(str)
	}
	str = GetINIFileVersionSectionEntry(fileName, "version")
	if str != noteVersion {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, noteVersion)
	}
	str = GetINIFileVersionSectionEntry(fileNameOld, "version")
	if str != oldNoteVersion {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, oldNoteVersion)
	}
	str = GetINIFileVersionSectionEntry(fileNotExist, "version")
	if str != "" {
		t.Errorf(str)
	}
	str = GetINIFileVersionSectionEntry(fileName, "date")
	if str != noteDate {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, noteDate)
	}
	str = GetINIFileVersionSectionEntry(fileNameOld, "date")
	if str != oldNoteDate {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, oldNoteDate)
	}
	str = GetINIFileVersionSectionEntry(fileNotExist, "date")
	if str != "" {
		t.Errorf(str)
	}
	str = GetINIFileVersionSectionEntry(fileName, "name")
	if str != noteTitle {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, noteTitle)
	}
	str = GetINIFileVersionSectionEntry(fileNameOld, "name")
	if str != oldNoteTitle {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, oldNoteTitle)
	}
	str = GetINIFileVersionSectionEntry(fileNotExist, "name")
	if str != "" {
		t.Errorf(str)
	}
	str = GetINIFileVersionSectionEntry(fileNameNew, "category")
	if str != oldNoteCategory {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, oldNoteCategory)
	}
	str = GetINIFileVersionSectionEntry(fileNameOld, "category")
	if str != oldNoteCategory {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, oldNoteCategory)
	}
	str = GetINIFileVersionSectionEntry(fileName, "not_avail")
	if str != "" {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, "")
	}
	str = GetINIFileVersionSectionEntry(fileNameWrong, "name")
	if str != "" {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, "")
	}
	str = GetINIFileVersionSectionEntry(fileNameMissing, "name")
	if str != "" {
		t.Errorf("\n'%+v'\nis not\n'%+v'\n", str, "")
	}
}

func TestBlkInfoNeeded(t *testing.T) {
	sectFields := []string{"hugo", "blkvendor=HUGO", "blkmodel=EGON"}
	if !blkInfoNeeded(sectFields) {
		t.Error("should be 'true', but returns 'false'")
	}
	sectFields = []string{"hugo", "HUGO", "EGON"}
	if blkInfoNeeded(sectFields) {
		t.Error("should be 'false', but returns 'true'")
	}
}
