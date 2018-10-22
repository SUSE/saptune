package txtparser

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

var fileName = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/notes/1410736")
var descName = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s", "TCP/IP: setting keepalive interval", "", "3", "25.11.2009 ")

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
		"Section": "reminder",
		"Key": "reminder",
		"Operator": "",
		"Value": ""
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
		"reminder": {
			"reminder": {
				"Section": "reminder",
				"Key": "reminder",
				"Operator": "",
				"Value": ""
			}
		}
	}
}`

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
}

func TestGetINIFileDescriptiveName(t *testing.T) {
	str := GetINIFileDescriptiveName(fileName)
	if str != descName {
		t.Fatalf("\n'%+v'\nis not\n'%+v'\n", str, descName)
	}
}
