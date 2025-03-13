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

// GetTuningOptions returns all built-in tunable SAP notes together with those
// defined by 3rd party vendors.
func GetTuningOptions(saptuneTuningDir, thirdPartyTuningDir string) TuningOptions {
	ret := TuningOptions{}
	// Collect those defined by saptune
	_, files := system.ListDir(saptuneTuningDir, "saptune tuning definitions")
	for _, fileName := range files {
		ret[fileName] = INISettings{
			ConfFilePath:    path.Join(saptuneTuningDir, fileName),
			ID:              fileName,
			DescriptiveName: txtparser.GetINIFileDescriptiveName(path.Join(saptuneTuningDir, fileName)),
		}
	}

	if thirdPartyTuningDir == "" {
		return ret
	}

	// Collect those defined by 3rd party
	_, files = system.ListDir(thirdPartyTuningDir, "3rd party tuning definitions")
	for _, fileName := range files {
		if fileName == "solutions" || strings.HasSuffix(fileName, ".sol") {
			continue
		}
		// ignore left over files (BOBJ and ASE definition files) from
		// the migration of saptune version 1 to saptune version 2
		if fileName == "SAP_BOBJ-SAP_Business_OBJects.conf" || fileName == "SAP_ASE-SAP_Adaptive_Server_Enterprise.conf" {
			system.InfoLog("GetTuningOptions: skip old note definition \"%s\" from saptune version 1.", fileName)
			system.InfoLog("For more information refer to the man page saptune-migrate(7)")
			continue
		}
		if !strings.HasSuffix(fileName, ".conf") {
			// skip filenames without .conf suffix
			system.InfoLog("skip file \"%s\", wrong filename syntax, missing '.conf' suffix", fileName)
			continue
		}

		id := strings.TrimSuffix(fileName, ".conf")
		// get the description of the note from the header inside the file
		name := txtparser.GetINIFileDescriptiveName(path.Join(thirdPartyTuningDir, fileName))
		if name == "" {
			// no header found in the vendor file
			// fall back to the old style vendor file names
			// support of old style vendor file names for compatibility reasons
			system.InfoLog("GetTuningOptions: no header information found in file \"%s\"", fileName)
			system.InfoLog("falling back to old style vendor file names")
			system.WarningLog("old style vendor files are deprecated. For future support add header information to the file - %s", fileName)
			// By convention, the portion before dash makes up the ID.
			idName := strings.SplitN(fileName, "-", 2)
			if len(idName) != 2 {
				system.InfoLog("GetTuningOptions: bad file name \"%s\"", fileName)
			} else {
				id = idName[0]
				// Just for the cosmetics, remove suffix .conf from description
				name = strings.TrimSuffix(idName[1], ".conf")
			}
		}
		// Do not allow vendor to override built-in
		if _, exists := ret[id]; exists {
			system.InfoLog("GetTuningOptions: vendor's \"%s\" will not override built-in tuning implementation", fileName)
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

// GetNoteHeadData provides description, reference, version and released date
// of a given noteObj
func GetNoteHeadData(obj Note) (desc, vers, date string, refs []string) {
	objConfFile := reflect.ValueOf(obj).FieldByName("ConfFilePath").String()
	if objConfFile != "" {
		desc = txtparser.GetINIFileVersionSectionEntry(objConfFile, "name")
		vers = txtparser.GetINIFileVersionSectionEntry(objConfFile, "version")
		date = txtparser.GetINIFileVersionSectionEntry(objConfFile, "date")
		refs = txtparser.GetINIFileVersionSectionRefs(objConfFile)
	}
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
func CompareJSValue(v1, v2 interface{}, op string) (v1JS, v2JS string, match bool) {
	v1JSBytes, err := json.Marshal(v1)
	if err != nil {
		_ = system.ErrorLog("CompareJSValue: failed to serialise \"%+v\" - %v", v1, err)
		panic(err)
	}
	v2JSBytes, err := json.Marshal(v2)
	if err != nil {
		_ = system.ErrorLog("CompareJSValue: failed to serialise \"%+v\" - %v", v2, err)
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

	switch op {
	case "", "==":
		match = v1JS == v2JS
	case ">=":
		v1JSi, _ := strconv.Atoi(v1JS)
		v2JSi, _ := strconv.Atoi(v2JS)
		match = v1JSi >= v2JSi
	case "<=":
		v1JSi, _ := strconv.Atoi(v1JS)
		v2JSi, _ := strconv.Atoi(v2JS)
		match = v1JSi <= v2JSi
	}
	return
}

// CompareNoteFields compares the content of two notes and return differences
// in their fields in a human-readable text.
func CompareNoteFields(actualNote, expectedNote Note) (allMatch bool, comparisons map[string]FieldComparison, valApplyList []string) {
	comparisons = make(map[string]FieldComparison)
	allMatch = true
	grubAvail := false
	// Compare all fields
	refActualNote := reflect.ValueOf(actualNote)
	refExpectedNote := reflect.ValueOf(expectedNote)
	for i := 0; i < refActualNote.NumField(); i++ {
		// Retrieve actualField value from actual and expected note
		fieldName := reflect.TypeOf(actualNote).Field(i).Name
		// Compare map value or actualField value
		if refActualNote.Field(i).Type().Kind() == reflect.Map {
			// Compare map values
			actualMap := refActualNote.Field(i)
			expectedMap := refExpectedNote.Field(i)
			for _, key := range actualMap.MapKeys() {
				if strings.Contains(key.String(), "grub") {
					grubAvail = true
				}
				actualValue := actualMap.MapIndex(key).Interface()
				expectedValue := expectedMap.MapIndex(key).Interface()
				ckey := fmt.Sprintf("%s[%s]", fieldName, key.String())
				comparisons[ckey] = cmpMapValue(fieldName, key, actualValue, expectedValue)
				if !comparisons[ckey].MatchExpectation && comparisons[ckey].ReflectFieldName == "SysctlParams" {
					valApplyList = append(valApplyList, comparisons[ckey].ReflectMapKey)
				} else if key.String() == "force_latency" && comparisons[ckey].ReflectFieldName == "SysctlParams" {
					valApplyList = append(valApplyList, comparisons[ckey].ReflectMapKey)
				}
				if !comparisons[ckey].MatchExpectation && fieldName == "SysctlParams" {
					// a parameter, which is not supported
					// by the system ("all:none") should not
					// influence the compare result
					//
					// and grub compliance of saptune
					// integrated notes will be handled
					// at the end of the compare
					// all other grub settings treated as
					// normal parameters
					// if this should change in the future use
					// !strings.Contains(key.String(), "grub")
					// instead of !isInternalGrub(key.String())
					if actualValue.(string) != "all:none" && !isInternalGrub(key.String()) && !(system.IsXFSOption.MatchString(key.String()) && actualValue.(string) == "NA") && actualValue.(string) != "PNA" && key.String() != "VSZ_TMPFS_PERCENT" {
						allMatch = false
					}
				}
			}
		} else {
			// Compare ordinary field value
			// ConfFilePath, ID, DescriptiveName
			comparisons[fieldName] = cmpFieldValue(i, fieldName, refActualNote, refExpectedNote)
			if !comparisons[fieldName].MatchExpectation {
				allMatch = false
			}
		}
	}
	if allMatch && grubAvail {
		allMatch = chkGrubCompliance(comparisons, allMatch)
	}
	return
}

// isInternalGrub - checks, if a grub setting found in the note definition
// is a saptune integrated grub parameter or a customer specific parameter
func isInternalGrub(val string) bool {
	// define saptune integrated grub parameter
	internalGrub := []string{"grub:numa_balancing", "grub:transparent_hugepage", "grub:intel_idle.max_cstate", "grub:processor.max_cstate"}

	for _, item := range internalGrub {
		if item == val {
			return true
		}
	}
	return false
}

// chkGrubCompliance grub special - check compliance of alternative settings
// only if one of these alternatives are not compliant, modify the result of
// the compare
// restricted to grub parameter shipped with saptune integrated notes
// grub parameter and 'alternative' setting have to be within the same note
func chkGrubCompliance(comparisons map[string]FieldComparison, allMatch bool) bool {
	// grub:numa_balancing, kernel.numa_balancing
	// grub:transparent_hugepage, THP
	// grub:intel_idle.max_cstate, grub:processor.max_cstate, force_latency
	entries := []string{"grub:numa_balancing#kernel.numa_balancing", "grub:transparent_hugepage#THP", "grub:intel_idle.max_cstate#force_latency", "grub:processor.max_cstate#force_latency"}

	for _, item := range entries {
		entFields := strings.Split(item, "#")
		grubEnt := entFields[0]
		alterEnt := entFields[1]
		grubEntry := fmt.Sprintf("SysctlParams[%s]", grubEnt)
		alterEntry := fmt.Sprintf("SysctlParams[%s]", alterEnt)

		if comparisons[grubEntry].ReflectMapKey == grubEnt && !comparisons[grubEntry].MatchExpectation {
			if alterEnt == "force_latency" {
				if (comparisons[alterEntry].ReflectMapKey == alterEnt && !comparisons[alterEntry].MatchExpectation && comparisons[alterEntry].ActualValue != "all:none") && allMatch {
					allMatch = false
				}
			} else {
				if comparisons[alterEntry].ReflectMapKey == alterEnt && !comparisons[alterEntry].MatchExpectation && allMatch {
					allMatch = false
				}
			}
		}
	}
	return allMatch
}

// cmpMapValue compares map values
func cmpMapValue(fieldName string, key reflect.Value, actVal, expVal interface{}) FieldComparison {
	op := ""
	if key.String() == "force_latency" && actVal.(string) != "all:none" {
		op = "<="
	}
	actualValueJS, expectedValueJS, match := CompareJSValue(actVal, expVal, op)
	if strings.Split(key.String(), ":")[0] == "rpm" {
		match = system.CmpRpmVers(actVal.(string), expVal.(string))
	}
	if strings.Split(key.String(), ":")[0] == "systemd" {
		match = system.CmpServiceStates(actVal.(string), expVal.(string))
	}
	if expVal == "" {
		// if the expected value is empty, the parameter value will
		// be untouched
		// this case should not influence the compare result
		// so set match to true
		match = true
	}

	if key.String() == "reminder" {
		// a diff in the reminder section should not influence the
		// compare result. So set macth to true
		match = true
	}

	fieldComparison := FieldComparison{
		ReflectFieldName: fieldName,
		ReflectMapKey:    key.String(),
		ActualValue:      actVal,
		ExpectedValue:    expVal,
		ActualValueJS:    actualValueJS,
		ExpectedValueJS:  expectedValueJS,
		MatchExpectation: match,
	}
	return fieldComparison
}

// cmpFieldValue compares ordinary field value
func cmpFieldValue(fNo int, fieldName string, actNote, expNote reflect.Value) FieldComparison {
	actualValue := actNote.Field(fNo).Interface()
	expectedValue := expNote.Field(fNo).Interface()
	actualValueJS, expectedValueJS, match := CompareJSValue(actualValue, expectedValue, "")
	fieldComparison := FieldComparison{
		ReflectFieldName: fieldName,
		ActualValue:      actualValue,
		ExpectedValue:    expectedValue,
		ActualValueJS:    actualValueJS,
		ExpectedValueJS:  expectedValueJS,
		MatchExpectation: match,
	}
	return fieldComparison
}

// GetVendInfo reads content of stored vend information.
func GetVendInfo(initype, ID string) (Note, error) {
	var vendConf INISettings
	vendFileName := fmt.Sprintf("%s/%s_%s.run", system.SaptuneSectionDir, initype, ID)
	content, err := system.ReadConfigFile(vendFileName, false)
	if err == nil && len(content) != 0 {
		err = json.Unmarshal(content, &vendConf)
	}
	return vendConf, err
}
