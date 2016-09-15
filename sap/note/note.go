/*
SAP notes tune one or more parameters at a time.

Each note is capable of inspecting the system to determine current parameter values, assist each parameter
in calculating an optimised value, and apply all parameters.

A system can be tuned for more than one note at a time.
*/
package note

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

/*
An SAP note consisting of a series of tunable parameters that can be applied and reverted.
Parameter is immutable. Internal state changes can only be made to copies.
*/
type Note interface {
	Initialise() (Note, error) // Inspect all tuning parameters and return.
	Optimise() (Note, error)   // (Re)calculate all parameters, but do not apply.
	Apply() error              // Apply all tuning parameters without further calculation.
	Name() string              // The original note name.
}

var AllNotes = map[string]Note{
	"2205917":       HANARecommendedOSSettings{},
	"1557506":       LinuxPagingImprovements{},
	"1275776":       PrepareForSAPEnvironments{},
	"1984787":       AfterInstallation{},
	"2161991":       VmwareGuestIOElevator{},
	"SUSE-GUIDE-01": SUSESysOptimisation{},
	"SUSE-GUIDE-02": SUSENetCPUOptimisation{},
} // All tunable SAP notes and their IDs

// Return all note numbers, sorted in ascending order.
func GetSortedNoteIDs() (ret []string) {
	ret = make([]string, 0, len(AllNotes))
	for id := range AllNotes {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}

// Record the actual value versus expected value for a note field. The field name has to be the actual name in Go struct.
type NoteFieldComparison struct {
	ReflectFieldName               string
	ActualValue, ExpectedValue     interface{}
	ActualValueJS, ExpectedValueJS string
	MatchExpectation               bool
}

// Compare the content of two notes and return differences in their fields in a human-readable text.
func CompareNoteFields(actualNote, expectedNote Note) (allMatch bool, comparisons map[string]NoteFieldComparison) {
	comparisons = make(map[string]NoteFieldComparison)
	allMatch = true
	// Compare all fields
	refNoteActual := reflect.ValueOf(actualNote)
	refNoteExpected := reflect.ValueOf(expectedNote)
	for i := 0; i < refNoteActual.NumField(); i++ {
		// Retrieve field value from actual and expected note
		fieldName := reflect.TypeOf(actualNote).Field(i).Name
		actualValue := refNoteActual.Field(i).Interface()
		expectedValue := refNoteExpected.Field(i).Interface()
		// Compare field value by their serialised JSON string
		jsValNoteActual, err := json.Marshal(actualValue)
		if err != nil {
			panic(fmt.Errorf("JSON error in JSONCompare for field %s: %v", fieldName, err))
		}
		jsValNoteExpected, err := json.Marshal(expectedValue)
		if err != nil {
			panic(fmt.Errorf("JSON error in JSONCompare for field %s: %v", fieldName, err))
		}
		fieldComparison := NoteFieldComparison{
			ReflectFieldName: fieldName,
			ActualValue:      actualValue,
			ExpectedValue:    expectedValue,
			ActualValueJS:    string(jsValNoteActual),
			ExpectedValueJS:  string(jsValNoteExpected),
			MatchExpectation: string(jsValNoteActual) == string(jsValNoteExpected),
		}
		comparisons[fieldName] = fieldComparison
		if !fieldComparison.MatchExpectation {
			allMatch = false
		}
	}
	return
}
