package note

import (
	"fmt"
	"testing"
)

var paramNote1 = ParameterNoteEntry{
	NoteID: "start",
	Value:  "StartValue",
}
var paramNote2 = ParameterNoteEntry{
	NoteID: "entry2",
	Value:  "AdditionalValue",
}
var paramNote3 = ParameterNoteEntry{
	NoteID: "entry3",
	Value:  "LastValue",
}

func TestGetPathToParameter(t *testing.T) {
	val := GetPathToParameter("FILENAME4TEST")
	if val != "/var/lib/saptune/parameter/FILENAME4TEST" {
		t.Fatalf("parameter file name: %v.\n", val)
	}
}

func TestGetSavedParameterNotes(t *testing.T) {
	val := GetSavedParameterNotes("TEST_PARAMETER")
	if len(val.AllNotes) > 0 {
		t.Fatalf("parameter file for 'TEST_PARAMETER' exists. content: %+v.\n", val)
	}
}

func TestIDInParameterList(t *testing.T) {
	pNotes := ParameterNotes{
		AllNotes: make([]ParameterNoteEntry, 0, 8),
	}

	pNotes.AllNotes = append(pNotes.AllNotes, paramNote1)
	pNotes.AllNotes = append(pNotes.AllNotes, paramNote2)
	pNotes.AllNotes = append(pNotes.AllNotes, paramNote3)
	if !IDInParameterList("entry2", pNotes.AllNotes) {
		t.Fatalf("'entry2' not part of list '%+v'\n", pNotes)
	}
	if IDInParameterList("HUGO", pNotes.AllNotes) {
		t.Fatalf("'HUGO' is part of list '%+v'\n", pNotes)
	}
}

func TestListParams(t *testing.T) {
	val, tsterr := ListParams()
	if tsterr == nil && len(val) > 0 {
		t.Fatalf("there are parameter files stored: '%+v'\n", val)
	}
}

func TestCreateParameterStartValues(t *testing.T) {
	CreateParameterStartValues("TEST_PARAMETER", "TestStartValue")
	val := GetSavedParameterNotes("TEST_PARAMETER")
	if len(val.AllNotes) == 0 {
		t.Fatalf("missing parameter state file 'TEST_PARAMETER': '%+v'\n", val)
	}
	if val.AllNotes[0].NoteID != "start" {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong content in state file 'TEST_PARAMETER', 'start' is NOT the first entry, instead it's '%+v'\n", val.AllNotes[0].NoteID)
	}
	if val.AllNotes[0].Value != "TestStartValue" {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong start value in state file 'TEST_PARAMETER': '%+v'\n", val.AllNotes[0].Value)
	}
	CleanUpParamFile("TEST_PARAMETER")

	// empty start value
	CreateParameterStartValues("TEST_PARAMETER", "")
	val = GetSavedParameterNotes("TEST_PARAMETER")
	if len(val.AllNotes) == 0 {
		t.Fatalf("missing parameter state file 'TEST_PARAMETER': '%+v'\n", val)
	}
	if val.AllNotes[0].NoteID != "start" {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong content in state file 'TEST_PARAMETER', 'start' is NOT the first entry, instead it's '%+v'\n", val.AllNotes[0].NoteID)
	}
	if val.AllNotes[0].Value != "" {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong start value in state file 'TEST_PARAMETER': '%+v'\n", val.AllNotes[0].Value)
	}
	CleanUpParamFile("TEST_PARAMETER")
}

func TestAddParameterNoteValues(t *testing.T) {
	AddParameterNoteValues("TEST_PARAMETER", "TestAddValue", "4711")
	val := GetSavedParameterNotes("TEST_PARAMETER")
	if len(val.AllNotes) != 0 {
		t.Fatalf("parameter state file 'TEST_PARAMETER' exists. content: '%+v'\n", val)
	}

	CreateParameterStartValues("TEST_PARAMETER", "TestStartValue")
	AddParameterNoteValues("TEST_PARAMETER", "TestAddValue", "4711")
	val = GetSavedParameterNotes("TEST_PARAMETER")
	if len(val.AllNotes) == 0 {
		t.Fatalf("missing parameter state file 'TEST_PARAMETER': '%+v'\n", val)
	}
	if val.AllNotes[0].NoteID != "start" && val.AllNotes[1].NoteID != "4711" {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong content in state file 'TEST_PARAMETER': '%+v'\n", val)
	}
	if val.AllNotes[0].Value != "TestStartValue" && val.AllNotes[1].Value != "TestAddValue" {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong content in state file 'TEST_PARAMETER': '%+v'\n", val)
	}
	if !IDInParameterList("4711", val.AllNotes) {
		CleanUpParamFile("TEST_PARAMETER")
		t.Fatalf("wrong content in state file 'TEST_PARAMETER': '%+v'\n", val)
	}
	CleanUpParamFile("TEST_PARAMETER")
}

func TestSaveLimitsParameter(t *testing.T) {
	tkey := "TEST_LIMIT_PARAMETER"
	tdom := "TDOMAIN"
	titem := "TITEM"
	sval := "TestLimitStartValue"
	aval := "TestLimitAddValue"

	paramFile := fmt.Sprintf("%s_%s_%s", tkey, titem, tdom)

	SaveLimitsParameter(tkey, tdom, titem, sval, "start", "")
	val := GetSavedParameterNotes(tkey)
	if len(val.AllNotes) != 0 {
		val = GetSavedParameterNotes(paramFile)
		if len(val.AllNotes) != 0 {
			t.Fatalf("parameter state file exists. content: '%+v'\n", val)
		}
	}

	SaveLimitsParameter(tkey, tdom, titem, aval, "add", "47114711")
	val = GetSavedParameterNotes(tkey)
	if len(val.AllNotes) != 0 {
		val = GetSavedParameterNotes(paramFile)
		if len(val.AllNotes) != 0 {
			t.Fatalf("parameter state file exists. content: '%+v'\n", val)
		}
	}

	tkey = "LIMIT_HARD"
	paramFile = fmt.Sprintf("%s_%s_%s", tkey, titem, tdom)

	SaveLimitsParameter(tkey, tdom, titem, sval, "start", "")
	val = GetSavedParameterNotes(paramFile)
	if len(val.AllNotes) != 0 {
		t.Fatalf("parameter state file exists. content: '%+v'\n", val)
	}

	sval = "TDOMAIN:TestLimitStartValue"
	sout := "TDOMAIN:TestLimitStartValue "
	SaveLimitsParameter(tkey, tdom, titem, sval, "start", "")
	val = GetSavedParameterNotes(paramFile)
	if len(val.AllNotes) == 0 {
		t.Fatalf("missing parameter state file '%s'\n", paramFile)
	}
	if val.AllNotes[0].NoteID != "start" {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong content in state file '%s': '%+v'\n", paramFile, val)
	}
	if val.AllNotes[0].Value != sout {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong content in state file '%s': '%+v'\n", paramFile, val)
	}

	aval = "TDOMAIN:TestLimitAddValue"
	aout := "TDOMAIN:TestLimitAddValue "
	SaveLimitsParameter(tkey, tdom, titem, aval, "add", "47114711")
	val = GetSavedParameterNotes(paramFile)
	if len(val.AllNotes) == 0 {
		t.Fatalf("missing parameter state file '%+v'\n", paramFile)
	}
	if val.AllNotes[0].NoteID != "start" && val.AllNotes[1].NoteID != "47114711" {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong content in state file '%s': '%+v'\n", paramFile, val)
	}
	if val.AllNotes[0].Value != sout && val.AllNotes[1].Value != aout {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong content in state file '%s': '%+v'\n", paramFile, val)
	}
	if IDInParameterList("4711", val.AllNotes) {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong content in state file '%s': '%+v'\n", paramFile, val)
	}
	if !IDInParameterList("47114711", val.AllNotes) {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong content in state file '%s': '%+v'\n", paramFile, val)
	}
	CleanUpParamFile(paramFile)
}

func TestGetAllSavedParameters(t *testing.T) {
	CreateParameterStartValues("TEST_PARAMETER_1", "TestStartValue1")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue1", "4711")
	CreateParameterStartValues("TEST_PARAMETER_2", "TestStartValue2")
	AddParameterNoteValues("TEST_PARAMETER_2", "TestAddValue2", "4712")
	CreateParameterStartValues("TEST_PARAMETER_3", "TestStartValue3")
	AddParameterNoteValues("TEST_PARAMETER_3", "TestAddValue3", "4713")

	val := GetAllSavedParameters()
	if val["TEST_PARAMETER_1"].AllNotes[0].NoteID != "start" && val["TEST_PARAMETER_1"].AllNotes[1].NoteID != "4711" {
		CleanUpParamFile("TEST_PARAMETER_1")
		CleanUpParamFile("TEST_PARAMETER_2")
		CleanUpParamFile("TEST_PARAMETER_3")
		t.Fatalf("wrong content in state file '%s': '%+v'\n", "TEST_PARAMETER_1", val["TEST_PARAMETER_1"].AllNotes)
	}
	if val["TEST_PARAMETER_1"].AllNotes[0].Value != "TestStartValue1" && val["TEST_PARAMETER_1"].AllNotes[1].Value != "TestAddValue1" {
		CleanUpParamFile("TEST_PARAMETER_1")
		CleanUpParamFile("TEST_PARAMETER_2")
		CleanUpParamFile("TEST_PARAMETER_3")
		t.Fatalf("wrong content in state file '%s': '%+v'\n", "TEST_PARAMETER_1", val["TEST_PARAMETER_1"].AllNotes)
	}
	if val["TEST_PARAMETER_2"].AllNotes[0].NoteID != "start" && val["TEST_PARAMETER_2"].AllNotes[1].NoteID != "4712" {
		CleanUpParamFile("TEST_PARAMETER_1")
		CleanUpParamFile("TEST_PARAMETER_2")
		CleanUpParamFile("TEST_PARAMETER_3")
		t.Fatalf("wrong content in state file '%s': '%+v'\n", "TEST_PARAMETER_2", val["TEST_PARAMETER_2"].AllNotes)
	}
	if val["TEST_PARAMETER_2"].AllNotes[0].Value != "TestStartValue2" && val["TEST_PARAMETER_2"].AllNotes[1].Value != "TestAddValue2" {
		CleanUpParamFile("TEST_PARAMETER_1")
		CleanUpParamFile("TEST_PARAMETER_2")
		CleanUpParamFile("TEST_PARAMETER_3")
		t.Fatalf("wrong content in state file '%s': '%+v'\n", "TEST_PARAMETER_2", val["TEST_PARAMETER_2"].AllNotes)
	}
	if val["TEST_PARAMETER_3"].AllNotes[0].NoteID != "start" && val["TEST_PARAMETER_3"].AllNotes[1].NoteID != "4713" {
		CleanUpParamFile("TEST_PARAMETER_1")
		CleanUpParamFile("TEST_PARAMETER_2")
		CleanUpParamFile("TEST_PARAMETER_3")
		t.Fatalf("wrong content in state file '%s': '%+v'\n", "TEST_PARAMETER_3", val["TEST_PARAMETER_3"].AllNotes)
	}
	if val["TEST_PARAMETER_3"].AllNotes[0].Value != "TestStartValue3" && val["TEST_PARAMETER_3"].AllNotes[1].Value != "TestAddValue3" {
		CleanUpParamFile("TEST_PARAMETER_1")
		CleanUpParamFile("TEST_PARAMETER_2")
		CleanUpParamFile("TEST_PARAMETER_3")
		t.Fatalf("wrong content in state file '%s': '%+v'\n", "TEST_PARAMETER_3", val["TEST_PARAMETER_3"].AllNotes)
	}
	CleanUpParamFile("TEST_PARAMETER_1")
	CleanUpParamFile("TEST_PARAMETER_2")
	CleanUpParamFile("TEST_PARAMETER_3")
}

func TestStoreParameter(t *testing.T) {
	paramList := ParameterNotes{
		AllNotes: make([]ParameterNoteEntry, 0, 64),
	}
	param := ParameterNoteEntry{
		NoteID: "start",
		Value:  "TestStartValue1",
	}
	paramList.AllNotes = append(paramList.AllNotes, param)
	param = ParameterNoteEntry{
		NoteID: "4711",
		Value:  "TestAddValue1",
	}
	paramList.AllNotes = append(paramList.AllNotes, param)
	err := StoreParameter("TEST_PARAMETER_1", paramList, true)
	if err != nil {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("failed to store values for parameter '%s' in file: '%+v'\n", "TEST_PARAMETER_1", paramList)
	}
	CleanUpParamFile("TEST_PARAMETER_1")
}

func TestPositionInParameterList(t *testing.T) {
	CreateParameterStartValues("TEST_PARAMETER_1", "TestStartValue1")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue1", "4711")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue2", "4712")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue3", "4713")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue4", "4714")
	noteList := GetSavedParameterNotes("TEST_PARAMETER_1")
	val := PositionInParameterList("4712", noteList.AllNotes)
	if val != 2 {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong position for note '%s': '%v'\n", "4712", val)
	}
	val = PositionInParameterList("start", noteList.AllNotes)
	if val != 0 {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong position for note '%s': '%v'\n", "start", val)
	}
	val = PositionInParameterList("TEST_NON_EXIST", noteList.AllNotes)
	if val != 0 {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong position for note '%s': '%v'\n", "TEST_NON_EXIST", val)
	}
	CleanUpParamFile("TEST_PARAMETER_1")
}

func TestRevertParameter(t *testing.T) {
	// test with non existing parameter file
	val := RevertParameter("TEST_PARAMETER_1", "4712")
	if val != "" {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong parameter '%s' reverted from parameter file '%s'\n", val, "TEST_PARAMETER_1")
	}

	CreateParameterStartValues("TEST_PARAMETER_1", "TestStartValue1")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue1", "4711")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue2", "4712")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue3", "4713")
	AddParameterNoteValues("TEST_PARAMETER_1", "TestAddValue4", "4714")
	val = RevertParameter("TEST_PARAMETER_1", "4712")
	if val != "TestAddValue4" {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4712")
	}
	val = RevertParameter("TEST_PARAMETER_1", "4714")
	if val != "TestAddValue3" {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4714")
	}
	val = RevertParameter("TEST_PARAMETER_1", "4711")
	if val != "TestAddValue3" {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4711")
	}
	val = RevertParameter("TEST_PARAMETER_1", "4713")
	if val != "TestStartValue1" {
		CleanUpParamFile("TEST_PARAMETER_1")
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4713")
	}
	CleanUpParamFile("TEST_PARAMETER_1")
}

func TestRevertLimitsParameter(t *testing.T) {
	tkey := "LIMIT_HARD"
	tdom := "TDOMAIN"
	titem := "TITEM"
	sval := "TDOMAIN:TestLimitStartValue1"
	aval1 := "TDOMAIN:TestLimitAddValue1"
	aval2 := "TDOMAIN:TestLimitAddValue2"
	aval3 := "TDOMAIN:TestLimitAddValue3"
	aval4 := "TDOMAIN:TestLimitAddValue4"
	paramFile := fmt.Sprintf("%s_%s_%s", tkey, titem, tdom)

	CleanUpParamFile(paramFile)

	SaveLimitsParameter(tkey, tdom, titem, sval, "start", "")
	SaveLimitsParameter(tkey, tdom, titem, aval1, "add", "4711")
	SaveLimitsParameter(tkey, tdom, titem, aval2, "add", "4712")
	SaveLimitsParameter(tkey, tdom, titem, aval3, "add", "4713")
	SaveLimitsParameter(tkey, tdom, titem, aval4, "add", "4714")

	// test with 'wrong' key, return should be an empty string
	val := RevertLimitsParameter("LIMIT_TEST_KEY", tdom, titem, "4712")
	if val != "" {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong parameter '%s' reverted for key 'LIMIT_TEST_KEY' and for note '%s'\n", val, "4712")
	}

	val = RevertLimitsParameter(tkey, tdom, titem, "4712")
	if val != "TDOMAIN:TestLimitAddValue4 " {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4712")
	}
	val = RevertLimitsParameter(tkey, tdom, titem, "4714")
	if val != "TDOMAIN:TestLimitAddValue3 " {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4714")
	}
	val = RevertLimitsParameter(tkey, tdom, titem, "4711")
	if val != "TDOMAIN:TestLimitAddValue3 " {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4711")
	}
	val = RevertLimitsParameter(tkey, tdom, titem, "4713")
	if val != "TDOMAIN:TestLimitStartValue1 " {
		CleanUpParamFile(paramFile)
		t.Fatalf("wrong parameter '%s' reverted for note '%s'\n", val, "4713")
	}
	CleanUpParamFile(paramFile)
}
