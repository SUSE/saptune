package note

/*
SAP notes tune one or more parameters at a time.

Each note is capable of inspecting the system to determine current parameter values, assist each parameter
in calculating an optimised value, and apply all parameters.

A system can be tuned for more than one note at a time.
*/

import (
	"encoding/json"
	"fmt"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Note defines the structure and actions for a SAP Note
// An SAP note consisting of a series of tunable parameters that can be
// applied and reverted.
// Parameter is immutable. Internal state changes can only be made to copies.
type Note interface {
	Initialise() (Note, error) // Inspect all tuning parameters and return.
	Optimise() (Note, error)   // (Re)calculate all parameters, but do not apply.
	Apply() error              // Apply all tuning parameters without further calculation.
	Name() string              // The original note name.
}

// TuningOptions is the collection of tuning options from SAP notes and
// 3rd party vendors.
type TuningOptions map[string]Note

// Note2Convert checks, if there is a note in an older saptune format to revert
// support revert from older saptune versions
func Note2Convert(noteID string) string {
	fileName := fmt.Sprintf("/var/lib/saptune/saved_state/%s_n2c", noteID)
	if _, err := os.Stat(fileName); err == nil {
		noteID = fmt.Sprintf("%s_n2c", noteID)
	}
	return noteID
}

// GetTuningOptions returns all built-in tunable SAP notes together with those
// defined by 3rd party vendors.
func GetTuningOptions(saptuneTuningDir, thirdPartyTuningDir string) TuningOptions {
	ret := TuningOptions{}
	// Collect those defined by saptune
	_, files, err := system.ListDir(saptuneTuningDir)
	if err != nil {
		// Not a fatal error
		system.WarningLog("GetTuningOptions: failed to read saptune tuning definitions - %v", err)
	}
	for _, fileName := range files {
		ret[fileName] = INISettings{
			ConfFilePath:    path.Join(saptuneTuningDir, fileName),
			ID:              fileName,
			DescriptiveName: "",
		}
	}

	// Collect those defined by 3rd party
	_, files, err = system.ListDir(thirdPartyTuningDir)
	if err != nil {
		// Not a fatal error
		system.WarningLog("GetTuningOptions: failed to read 3rd party tuning definitions - %v", err)
	}
	for _, fileName := range files {
		// ignore left over files (BOBJ and ASE definition files) from
		// the migration of saptune version 1 to saptune version 2
		if fileName == "SAP_BOBJ-SAP_Business_OBJects.conf" || fileName == "SAP_ASE-SAP_Adaptive_Server_Enterprise.conf" {
			system.WarningLog("GetTuningOptions: skip old note definition \"%s\" from saptune version 1.", fileName)
			system.WarningLog("For more information refer to the man page saptune-migrate(7)")
			continue
		}
		id := ""
		// get the description of the note from the header inside the file
		name := txtparser.GetINIFileDescriptiveName(path.Join(thirdPartyTuningDir, fileName))
		if name == "" {
			// no header found in the vendor file
			// fall back to the old style vendor file names
			// support of old style vendor file names for compatibility reasons
			system.WarningLog("GetTuningOptions: no header information found in file \"%s\"", fileName)
			system.WarningLog("falling back to old style vendor file names")
			// By convention, the portion before dash makes up the ID.
			idName := strings.SplitN(fileName, "-", 2)
			if len(idName) != 2 {
				system.WarningLog("GetTuningOptions: skip bad file name \"%s\"", fileName)
				continue
			}
			id = idName[0]
			// Just for the cosmetics, remove suffix .conf from description
			name = strings.TrimSuffix(idName[1], ".conf")
		} else {
			// description found in header of the file
			// let name empty, to get the right information during 'note list'
			id = strings.TrimSuffix(fileName, ".conf")
		}
		// Do not allow vendor to override built-in
		if _, exists := ret[id]; exists {
			system.WarningLog("GetTuningOptions: vendor's \"%s\" will not override built-in tuning implementation", fileName)
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

// GetSortedIDs returns all tuning option IDs, sorted in ascending order.
func (opts *TuningOptions) GetSortedIDs() (ret []string) {
	ret = make([]string, 0, len(*opts))
	for id := range *opts {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}

// FieldComparison records the actual value versus expected value for
// a note field. The field name has to be the actual name in Go struct.
type FieldComparison struct {
	ReflectFieldName               string // Structure field name
	ReflectMapKey                  string // If structure field is a map, this is the map key
	ActualValue, ExpectedValue     interface{}
	ActualValueJS, ExpectedValueJS string
	MatchExpectation               bool
}

// CompareJSValue compares JSON representation of two values and see
// if they match.
func CompareJSValue(v1, v2 interface{}) (v1JS, v2JS string, match bool) {
	v1JSBytes, err := json.Marshal(v1)
	if err != nil {
		system.ErrorLog("CompareJSValue: failed to serialise \"%+v\" - %v", v1, err)
		panic(err)
	}
	v2JSBytes, err := json.Marshal(v2)
	if err != nil {
		system.ErrorLog("CompareJSValue: failed to serialise \"%+v\" - %v", v2, err)
		panic(err)
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

// CompareNoteFields compares the content of two notes and return differences
// in their fields in a human-readable text.
func CompareNoteFields(actualNote, expectedNote Note) (allMatch bool, comparisons map[string]FieldComparison, valApplyList []string) {
	comparisons = make(map[string]FieldComparison)
	allMatch = true
	// Compare all fields
	refActualNote := reflect.ValueOf(actualNote)
	refExpectedNote := reflect.ValueOf(expectedNote)
	for i := 0; i < refActualNote.NumField(); i++ {
		var fieldComparison FieldComparison
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
				if strings.Split(key.String(), ":")[0] == "rpm" {
					match = system.CmpRpmVers(actualValue.(string), expectedValue.(string))
				}
				fieldComparison = FieldComparison{
					ReflectFieldName: fieldName,
					ReflectMapKey:    key.String(),
					ActualValue:      actualValue,
					ExpectedValue:    expectedValue,
					ActualValueJS:    actualValueJS,
					ExpectedValueJS:  expectedValueJS,
					MatchExpectation: match,
				}
				comparisons[fmt.Sprintf("%s[%s]", fieldName, key.String())] = fieldComparison
				if !fieldComparison.MatchExpectation && fieldComparison.ReflectFieldName == "SysctlParams" {
					valApplyList = append(valApplyList, fieldComparison.ReflectMapKey)
				}
				if !fieldComparison.MatchExpectation {
					allMatch = false
				}
			}
		} else {
			// Compare ordinary field value
			actualValue := refActualNote.Field(i).Interface()
			expectedValue := refExpectedNote.Field(i).Interface()
			actualValueJS, expectedValueJS, match := CompareJSValue(actualValue, expectedValue)
			fieldComparison = FieldComparison{
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
