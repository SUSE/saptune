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
	"github.com/HouzuoGuo/saptune/system"
	"log"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

type TuningOptions map[string]Note // Collection of tuning options from SAP notes and 3rd party vendors.

// Return all built-in tunable SAP notes together with those defined by 3rd party vendors.
func GetTuningOptions(thirdPartyTuningDir string) TuningOptions {
	ret := TuningOptions{
		"1680803":       ASERecommendedOSSettings{},
		"2205917":       HANARecommendedOSSettings{},
		"1557506":       LinuxPagingImprovements{},
		"1275776":       PrepareForSAPEnvironments{},
		"1984787":       AfterInstallation{},
		"2161991":       VmwareGuestIOElevator{},
		"SUSE-GUIDE-01": SUSESysOptimisation{},
		"SUSE-GUIDE-02": SUSENetCPUOptimisation{},
	}
	// Collect those defined by 3rd party
	_, files, err := system.ListDir(thirdPartyTuningDir)
	if err != nil {
		// Not a fatal error
		log.Printf("GetTuningOptions: failed to read 3rd party tuning definitions - %v", err)
	}
	for _, fileName := range files {
		// By convention, the portion before dash makes up the ID.
		idName := strings.SplitN(fileName, "-", 2)
		if len(idName) != 2 {
			log.Printf("GetTuningOptions: skip bad file name \"%s\"", fileName)
			continue
		}
		id := idName[0]
		// Just for the cosmetics, remove suffix .conf from description
		name := strings.TrimSuffix(idName[1], ".conf")
		// Do not allow vendor to override built-in
		if _, exists := ret[id]; exists {
			log.Printf("GetTuningOptions: vendor's \"%s\" will not override built-in tuning implementation", fileName)
			continue
		}
		ret[id] = INISettings{
			ConfFilePath:    path.Join(thirdPartyTuningDir, fileName),
			ID:              id,
			DescriptiveName: name,
		}
	}
	return ret
}

// Return all tuning option IDs, sorted in ascending order.
func (opts *TuningOptions) GetSortedIDs() (ret []string) {
	ret = make([]string, 0, len(*opts))
	for id := range *opts {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}

// Record the actual value versus expected value for a note field. The field name has to be the actual name in Go struct.
type NoteFieldComparison struct {
	ReflectFieldName               string // Structure field name
	ReflectMapKey                  string // If structure field is a map, this is the map key
	ActualValue, ExpectedValue     interface{}
	ActualValueJS, ExpectedValueJS string
	MatchExpectation               bool
}

// Compare JSON representation of two values and see if they match.
func CompareJSValue(v1, v2 interface{}) (v1JS, v2JS string, match bool) {
	v1JSBytes, err := json.Marshal(v1)
	if err != nil {
		log.Panicf("CompareJSValue: failed to serialise \"%+v\" - %v", v1, err)
	}
	v2JSBytes, err := json.Marshal(v2)
	if err != nil {
		log.Panicf("CompareJSValue: failed to serialise \"%+v\" - %v", v2, err)
	}
	v1JS, err = strconv.Unquote(string(v1JSBytes))
	if err != nil {
		v1JS = string(v1JSBytes)
	}
	v2JS, err = strconv.Unquote(string(v2JSBytes))
	if err != nil {
		v2JS = string(v2JSBytes)
	}
	match = v1JS == v2JS
	return
}

// Compare the content of two notes and return differences in their fields in a human-readable text.
func CompareNoteFields(actualNote, expectedNote Note) (allMatch bool, comparisons map[string]NoteFieldComparison) {
	comparisons = make(map[string]NoteFieldComparison)
	allMatch = true
	// Compare all fields
	refActualNote := reflect.ValueOf(actualNote)
	refExpectedNote := reflect.ValueOf(expectedNote)
	for i := 0; i < refActualNote.NumField(); i++ {
		var fieldComparison NoteFieldComparison
		// Retrieve actualField value from actual and expected note
		fieldName := reflect.TypeOf(actualNote).Field(i).Name
		actualField := refActualNote.Field(i)
		// Compare map value or actualField value
		if actualField.Type().Kind() == reflect.Map {
			// Compare map values
			expectedMap := refExpectedNote.Field(i)
			for _, key := range actualField.MapKeys() {
				actualValue := actualField.MapIndex(key).Interface()
				expectedValue := expectedMap.MapIndex(key).Interface()
				actualValueJS, expectedValueJS, match := CompareJSValue(actualValue, expectedValue)
				fieldComparison = NoteFieldComparison{
					ReflectFieldName: fieldName,
					ReflectMapKey:    key.String(),
					ActualValue:      actualValue,
					ExpectedValue:    expectedValue,
					ActualValueJS:    actualValueJS,
					ExpectedValueJS:  expectedValueJS,
					MatchExpectation: match,
				}
				comparisons[fmt.Sprintf("%s[%s]", fieldName, key.String())] = fieldComparison
				if !fieldComparison.MatchExpectation {
					allMatch = false
				}
			}
		} else {
			// Compare ordinary field value
			actualValue := refActualNote.Field(i).Interface()
			expectedValue := refExpectedNote.Field(i).Interface()
			actualValueJS, expectedValueJS, match := CompareJSValue(actualValue, expectedValue)
			fieldComparison = NoteFieldComparison{
				ReflectFieldName: fieldName,
				ActualValue:      actualValue,
				ExpectedValue:    expectedValue,
				ActualValueJS:    actualValueJS,
				ExpectedValueJS:  expectedValueJS,
				MatchExpectation: match,
			}
			comparisons[fieldName] = fieldComparison
			if !fieldComparison.MatchExpectation {
				allMatch = false
			}
		}
	}
	return
}
