package txtparser

import (
	"encoding/json"
	"reflect"
	"testing"
)

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
