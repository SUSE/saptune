package app

import (
	"fmt"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"strconv"
	"strings"
)

// collectChangedParameterInfo extracts the changed parameter and their values
// from the note definition file and/or override file
// search for new added, deleted or changed parameter
func collectChangedParameterInfo(noteID, fileName string, comparisons map[string]note.FieldComparison, app *App) map[string]map[string]interface{} {
	chgParams := make(map[string]map[string]interface{})
	ovKeyValue := make(map[string]map[string]txtparser.INIEntry)

	// check, if note file has changed
	// content of override file, if exists
	override, ovCont := txtparser.GetOverrides("ovw", noteID)
	// note in working area == the 'new' note with (perhaps) changes
	content, err := txtparser.ParseINIFile(fileName, false)
	// runtime section info of the note == the 'old' note content
	sectCont, secterr := txtparser.GetSectionInfo("rosi", noteID, false)
	if err != nil || secterr != nil {
		system.ErrorLog("Problems while getting parameter information of note '%s'. - %v - %v", noteID, err, secterr)
		return chgParams
	}
	// check for changed or added parameter
	// to get the changes: diff between note in working area and info stored
	// in sections file /run/saptune/sections/<noteid>.sections
	for _, param := range content.AllValues {
		switch param.Section {
		case note.INISectionVersion, note.INISectionRpm, note.INISectionGrub, note.INISectionFS, note.INISectionReminder:
			// These parameters are only checked, but not applied.
			// So nothing to do during refresh
			continue
		case note.INISectionCPU, note.INISectionMEM, note.INISectionService, note.INISectionBlock, note.INISectionLimits, note.INISectionLogin, note.INISectionPagecache:
			// currently not supported for 'refresh'
			system.InfoLog("parameters from section '%s' currently not supported and not evaluated for 'refresh' operation", param.Section)
			continue
		}

		// initialise parameter entry
		paramEntry := initChangedParams(param, noteID, "newchange")
		// check for new parameter
		changed := chkForNewOrChangedParams(sectCont, paramEntry)

		// check, if parameter is in override file and
		if override && ovCont != nil {
			chkOverrideParameter(ovCont.KeyValue, paramEntry, changed)
		}
		// check, if parameter value has changed in Note and/or
		// override file by checking against value in parameter file
		changed = chkParameterValue(paramEntry, changed, app)

		if paramEntry["wasUntouched"].(bool) && !paramEntry["isUntouched"].(bool) {
			// parameter WAS untouched, but has a value now
			paramEntry["changedParam"] = true
			changed = true
		}
		if changed {
			// parameter changed in note file OR in override file
			// or both
			chgParams[param.Key] = paramEntry
		}
	}
	// check for deleted parameter
	if ovCont != nil {
		ovKeyValue = ovCont.KeyValue
	}
	chkForDeletedParams(noteID, sectCont.AllValues, content.KeyValue, ovKeyValue, chgParams)
	return chgParams
}

// initChangedParams initialises the parameter entry
func initChangedParams(param txtparser.INIEntry, noteID, mode string) map[string]interface{} {
	paramEntry := make(map[string]interface{})
	paramEntry["noteID"] = noteID
	// key (parameter name) from note file
	paramEntry["key"] = param.Key
	// section from note file
	paramEntry["section"] = param.Section
	// value from note file
	paramEntry["nValue"] = param.Value
	// operator from note file
	paramEntry["nOp"] = param.Operator
	// index position for 'insert'
	paramEntry["idx"] = -1
	// value from section file
	paramEntry["sectValue"] = param.Value
	// value from override file
	paramEntry["overValue"] = ""
	paramEntry["overOp"] = ""
	// parameter is untouched
	paramEntry["isUntouched"] = false
	paramEntry["isUntouchedinNote"] = false
	paramEntry["isUntouchedinOver"] = false
	// parameter was untouched
	paramEntry["wasUntouched"] = false
	paramEntry["wasUntouchedinNote"] = false
	paramEntry["wasUntouchedinOver"] = false
	// parameter in note file deleted
	paramEntry["deletedParam"] = false
	paramEntry["newParam"] = false
	// parameter in note file changed
	paramEntry["changedParam"] = false
	// parameter in note file changed
	paramEntry["changedInNote"] = false
	paramEntry["override"] = false
	// parameter in override file changed
	paramEntry["changedInOver"] = false
	// noteChain values from parameter file
	paramEntry["noteChainPreID"] = ""
	paramEntry["noteChainPostID"] = ""
	paramEntry["noteChainPreValue"] = ""
	paramEntry["noteChainPostValue"] = ""
	paramEntry["isLastNote"] = false

	// adapt values
	if param.Value == "" && mode != "del" {
		paramEntry["isUntouched"] = true
		paramEntry["isUntouchedinNote"] = true
	}
	return paramEntry
}

// check for new or changed parameter in note definition file
func chkForNewOrChangedParams(sectCont *txtparser.INIFile, paramEntry map[string]interface{}) bool {
	system.DebugLog("chkForNewOrChangedParams - paramEntry is '%+v'", paramEntry)
	changed := false
	if sectCont == nil {
		return changed
	}
	section := paramEntry["section"].(string)
	key := paramEntry["key"].(string)
	/*
		// ANGI TODO - do I need section handling or is it handled by
		// the current checks?
		if _, ok := sectCont.KeyValue[section]; !ok {
			// new section
			paramEntry["newParam"] = true
			changed = true
		} else if _, ok := sectCont.KeyValue[section][key]; !ok {
	*/
	if _, ok := sectCont.KeyValue[section][key]; !ok {
		// not in stored section info - new parameter
		paramEntry["newParam"] = true
		changed = true
		return changed
	}
	if sectCont.KeyValue[section][key].Value != paramEntry["nValue"].(string) {
		// check for changed parameter in note file
		paramEntry["changedInNote"] = true
		paramEntry["changedParam"] = true
		paramEntry["sectValue"] = sectCont.KeyValue[section][key].Value
		changed = true
	}
	// if the parameter was untouched in Note during
	// apply, the NoteID is not in the parameter state file, but
	// in the section file with empty ('') value
	if sectCont.KeyValue[section][key].Value == "" {
		paramEntry["wasUntouched"] = true
		paramEntry["wasUntouchedinNote"] = true
		paramEntry["sectValue"] = "untouched"
	}
	return changed
}

// chkOverrideParameter checks, if the parameter is available in the override
// file and if the override file was changed
// This is the value which is the expected value from the Note file OR the
// override file.
func chkOverrideParameter(ovKeyValues map[string]map[string]txtparser.INIEntry, paramEntry map[string]interface{}, changed bool) {
	section := paramEntry["section"].(string)
	key := paramEntry["key"].(string)
	if _, ok := ovKeyValues[section][key]; !ok {
		// parameter NOT available in override file
		return
	}
	// parameter available in override file
	if changed {
		// parameter was changed in note file
		system.WarningLog("Attention: found override entry for parameter '%s' currently changed in note '%s'. Please check if value in override file is still valid!", key, paramEntry["noteID"].(string))
	}

	ovValue := ovKeyValues[section][key].Value
	if ovValue == "" {
		paramEntry["isUntouched"] = true
		paramEntry["isUntouchedinOver"] = true
		ovValue = "untouched"
	}
	paramEntry["override"] = true
	paramEntry["overValue"] = ovValue
	paramEntry["overOp"] = ovKeyValues[section][key].Operator
}

// chkParameterValue checks the parameter state file
func chkParameterValue(paramEntry map[string]interface{}, changed bool, app *App) bool {
	key := paramEntry["key"].(string)
	paramStateValues := note.GetSavedParameterNotes(key)
	system.DebugLog("chkParameterValue - paramEntry is '%+v', paramStateValues are '%+v'", paramEntry, paramStateValues)
	noteFound := false
	paramStateVal := "NA"
	for _, psv := range paramStateValues.AllNotes {
		if psv.NoteID == paramEntry["noteID"].(string) {
			noteFound = true
			paramStateVal = psv.Value
			break
		}
	}

	// check against value from parameter file
	// BUT - if the parameter WAS 'untouched' in the note
	// or in the override file
	// -> the note is NOT listed in the parameter state file
	if noteFound && paramEntry["sectValue"].(string) != paramStateVal {
		// noteFound and paramEntry["sectValue"].(string) == "untouched"
		// -> value in parameter file is from an override file
		// "untouched" != paramStateVal => case covered by this if query

		// Value in section File != value in parameter file
		// value in parameter file is from former override file
		system.DebugLog("parameter value from section file (old note) '%s' differs from value of the parameter state file '%s'", paramEntry["sectValue"].(string), paramStateVal)
		system.DebugLog("This means value in parameter file is from override file")
		if paramEntry["override"].(bool) {
			if paramEntry["overValue"].(string) != paramStateVal {
				system.DebugLog("parameter changed in override file")
				// value changed in override file
				paramEntry["changedInOver"] = true
				paramEntry["changedParam"] = true
				changed = true
			}
		} else {
			// no override file available
			system.DebugLog("... but overide file did not exist any more")
			// --> There was an override file, but now it's removed
			// value in parameter file == value in old override file
			// now the value from note file will be applied and not
			// the former value from the removed override file
			// so we need to set 'changed' to true to get it applied
			//paramEntry["changedInOver"] = true ??
			changed = true
			paramEntry["changedParam"] = true
		}
	}
	if !noteFound && !paramEntry["newParam"].(bool) {
		// if the parameter was 'untouched' in the note or in
		// the override file -> the note is NOT listed in the
		// parameter state file
		//
		// noteID not found in parameter state file and
		// parameter is not a new parameter
		// -> parameter was untouched
		paramEntry["wasUntouched"] = true

		// check, if it was untouched in the override file
		noteSavedState, err := valuesFromSavedState(paramEntry["noteID"].(string), app)
		if err != nil {
			return changed
		}
		savedStateOverVal := noteSavedState.OverrideParams[key]
		if savedStateOverVal == "untouched" {
			paramEntry["wasUntouchedinOver"] = true
			paramEntry["wasUntouched"] = true
			if paramEntry["overValue"] != savedStateOverVal {
				// value changed in override file
				paramEntry["changedInOver"] = true
			}
		}
	}
	return changed
}

// chkForDeletedParams checks, if a parameter was deleted from the note
// definition file
// check for all parameter in the section file, if the parameter is still
// available in the note definition file
func chkForDeletedParams(noteID string, sectCont []txtparser.INIEntry, noteCont map[string]map[string]txtparser.INIEntry, ovCont map[string]map[string]txtparser.INIEntry, chgParams map[string]map[string]interface{}) {
	system.DebugLog("chkForDeletedParams - chgParams is '%+v'", chgParams)
	deleted := false
	for _, sectParam := range sectCont {
		paramEntry := initChangedParams(sectParam, noteID, "del")
		if _, ok := noteCont[sectParam.Section]; !ok {
			// section no longer available in note definition file
			// deleted section
			deleted = true
		} else if _, ok := noteCont[sectParam.Section][sectParam.Key]; !ok {
			// parameter no longer available in note definition file
			// deleted parameter
			deleted = true
		}
		if deleted {
			paramEntry["deletedParam"] = true
			if sectParam.Value == "" {
				paramEntry["wasUntouched"] = true
			}
			chgParams[sectParam.Key] = paramEntry

			// check if deleted parameter is available in override
			// file and print a warning
			if _, ok := ovCont[sectParam.Section]; ok {
				if _, ok = ovCont[sectParam.Section][sectParam.Key]; ok {
					system.WarningLog("Attention: found override entry for parameter '%s' currently deleted in note '%s'. Please remove parameter from override file. Setting will be ignored anyway as parameter no longer available in note file!", paramEntry["key"], paramEntry["noteID"].(string))
				}
			}
		}
	}
}

// adjustStateFiles will adjust the parameter values in the related state
// files
func adjustStateFiles(noteID string, app *App, changedParameter map[string]map[string]interface{}, comparisons map[string]note.FieldComparison) ([]string, error) {
	system.DebugLog("adjustStateFiles - changedParameter is '%+v'", changedParameter)
	valApplyList := []string{}
	needApply := false
	noteApplyOrder := app.NoteApplyOrder
	err := adjustSectionFile(noteID, changedParameter)
	if err != nil {
		return valApplyList, err
	}
	for key, param := range changedParameter {
		savedStateChange := map[string]string{}
		needApply, savedStateChange = adjustParameterFile(noteApplyOrder, param, comparisons[fmt.Sprintf("SysctlParams[%s]", key)], app)
		if needApply {
			// parameter need to be applied later
			valApplyList = append(valApplyList, key)
		}
		system.DebugLog("valApplyList is '%+v'", valApplyList)
		if savedStateChange["needChange"] == "delete" && param["changedInOver"].(bool) {
			savedStateChange["over"] = "over:delete"
		}
		if savedStateChange["needChange"] != "delete" && param["changedInOver"].(bool) {
			savedStateChange["over"] = "over:" + param["overValue"].(string)
		}
		// ANGI TODO - error inside loop - handling?
		err = adjustSavedStateFile(noteID, key, savedStateChange, app)
	}
	return valApplyList, err
}

// adjustSavedStateFile will adjust the parameter value in the saved_state file
// saved_state file contains all parameter of a note with the system values
// valid at the time of the apply of the note
//
// savedStateChange["needChange"] = "NO/add/delete/change",
func adjustSavedStateFile(noteID, key string, savedStateChange map[string]string, app *App) error {
	system.DebugLog("adjustSavedStateFile - noteID is '%s', key is '%s', savedStateChange is '%+v'", noteID, key, savedStateChange)
	var err error
	if savedStateChange["over"] != "" {
		// delete, add or change override part of saved_state file
		err = changeSavedStateEntry(noteID, key, savedStateChange["over"], app)
	}
	if savedStateChange["needChange"] == "NO" {
		return err
	}
	// ANGI TODO - error handling
	if savedStateChange["needChange"] == "delete" {
		// remove key entry, if available
		err = changeSavedStateEntry(noteID, key, "", app)
	}
	if savedStateChange["noteID"] != "" && savedStateChange["needChange"] != "delete" {
		// add or change
		err = changeSavedStateEntry(noteID, key, savedStateChange["noteID"], app)
	}
	if savedStateChange["postID"] != "" && savedStateChange["postValue"] != "" {
		// change successor
		err = changeSavedStateEntry(savedStateChange["postID"], key, savedStateChange["postValue"], app)
	}
	return err
}

// valuesFromSavedState reads content of saved state file
func valuesFromSavedState(noteID string, app *App) (note.INISettings, error) {
	noteSavedState := note.INISettings{
		ConfFilePath:    "",
		ID:              "",
		DescriptiveName: "",
	}
	err := app.State.Retrieve(noteID, &noteSavedState)
	if err != nil {
		system.ErrorLog("Failed to retrieve saved state of note %s - %v", noteID, err)
	}
	return noteSavedState, err
}

// changeSavedStateEntry changes the value of entry 'key' in the saved_state
// file of noteID
func changeSavedStateEntry(noteID, key, value string, app *App) error {
	noteSavedState, err := valuesFromSavedState(noteID, app)
	if err != nil {
		return err
	}
	if value == "" {
		// remove key entry
		delete(noteSavedState.SysctlParams, key)
		delete(noteSavedState.Inform, key)
	} else if strings.Contains(value, "over:") {
		ofields := strings.Split(value, ":")
		if len(ofields) > 1 {
			if ofields[1] == "delete" {
				//delete
				delete(noteSavedState.OverrideParams, key)
			} else {
				// add or change value
				noteSavedState.OverrideParams[key] = ofields[1]
			}
		}
	} else {
		// add or change value
		noteSavedState.SysctlParams[key] = value
	}
	err = app.State.Store(noteID, noteSavedState, true)
	if err != nil {
		system.ErrorLog("Failed to save current state of note %s - %v", noteID, err)
	}
	return err
}

// adjustParameterFile will adjust the parameter value in the parameter file
// The parameter file contains the noteID and the value this note will set to
// the system (value from the note file or the override file)
// If the value is 'untouched' (expectedValue is empty), the noteID will NOT be
// part of the parameter file
func adjustParameterFile(noteApplyOrder []string, param map[string]interface{}, comparison note.FieldComparison, app *App) (bool, map[string]string) {
	noteID := param["noteID"].(string)
	key := param["key"].(string)
	if app.NoteApplyOrder[len(app.NoteApplyOrder)-1] == noteID {
		// note is last note in noteApplyOrder
		param["isLastNote"] = true
	}
	savedStateChange := make(map[string]string)
	needApply := false
	paramStateValues := note.GetSavedParameterNotes(key)
	paramStatePos := paramStateValues.PositionInParameterList(noteID)
	// paramStatePos == 0 (note ID not in file, no file or only 'start' in file)
	// len(paramStateValues.AllNotes) == 1 only the 'start' value in file
	// setup note chain - predecessor ID - noteID - successor ID
	// needed for saved_state adjustment later
	noteChainSetup(paramStateValues, paramStatePos, param)
	system.DebugLog("noteChaine values - noteChainPreValue is '%s', noteChainPostID is '%s', noteChainPostValue is '%s'", param["noteChainPreValue"].(string), param["noteChainPostID"].(string), param["noteChainPostValue"].(string))

	// savedStateChange["needChange"] = "NO/add/delete/change",
	// savedStateChange["noteID"] = value
	// savedStateChange["over"] = value
	// savedStateChange["postID"], savedStateChange["postValue"]
	// savedStateChange["preID"], savedStateChange["preValue"]
	savedStateChange["needChange"] = "NO"

	// check NoteApplyOrder for order of Notes and check, if a successor Note from
	// NoteApplyOrder is already available in the parameter file
	param["idx"] = pNoteInsertPosition(noteID, noteApplyOrder, paramStateValues)
	if len(paramStateValues.AllNotes) == 0 {
		// parameter file NOT available
		needApply = paramFileNotAvailorStart(param, comparison, "noFile", savedStateChange)
	} else if len(paramStateValues.AllNotes) == 1 {
		// only the 'start' value is available
		// so simply append new or changed Note parameter
		needApply = paramFileNotAvailorStart(param, comparison, "start", savedStateChange)
	} else if paramStatePos == 0 {
		// note ID not available in existing parameter file
		// append or insert new or changed Note parameter values
		needApply = paramFileNoNoteID(param, comparison, needApply, savedStateChange)
		if param["deletedParam"].(bool) && param["wasUntouched"].(bool) {
			return needApply, savedStateChange
		}
		if param["idx"].(int) > 0 {
			// found successor Note in parameter file and inserted
			// noteID at index position
			// need to setup noteChain again after inserting the
			// new NoteID to parameterfile
			postNote := param["noteChainPostID"].(string)
			pos := 0
			paramStateValuesNew := note.GetSavedParameterNotes(key)
			if param["isUntouched"].(bool) && param["newParam"].(bool) {
				pos = paramStateValuesNew.PositionInParameterList(postNote)
			} else {
				pos = paramStateValuesNew.PositionInParameterList(noteID)
			}
			noteChainSetup(paramStateValuesNew, pos, param)
			system.DebugLog("noteChaine values after re-setup - noteChainPreValue is '%s', noteChainPostID is '%s', noteChainPostValue is '%s'", param["noteChainPreValue"].(string), param["noteChainPostID"].(string), param["noteChainPostValue"].(string))
			savedStateChange["noteID"] = param["noteChainPreValue"].(string)
			savedStateChange["postID"] = param["noteChainPostID"].(string)
			savedStateChange["postValue"] = comparison.ExpectedValue.(string)
			if param["isUntouched"].(bool) && param["changedParam"].(bool) {
				savedStateChange["postValue"] = param["noteChainPreValue"].(string)
			}
		} else {
			savedStateChange["noteID"] = comparison.ActualValue.(string)
		}
	} else {
		// note ID in parameter file available
		needApply = paramFileNoteIDAvail(param, comparison, needApply, savedStateChange)
	}
	return needApply, savedStateChange
}

// paramFileNotAvailorStart handles the situation that the parameter file of the
// changed parameter is NOT available
// or contains only the 'start' value
// no successor Note available (or untouched successor - ANGI TODO)
func paramFileNotAvailorStart(param map[string]interface{}, comparison note.FieldComparison, fileState string, savedStateChange map[string]string) bool {
	noteID := param["noteID"].(string)
	key := param["key"].(string)
	needApply := false
	if fileState == "start" {
		system.DebugLog("parameter file for parameter '%s' contains only the start value, but not Note ID '%s'.", key, noteID)
	} else {
		system.DebugLog("parameter file for parameter '%s' not available.", key)
	}
	if param["deletedParam"].(bool) {
		// delete parameter entry
		if param["wasUntouched"].(bool) {
			system.DebugLog("But parameter reported as deleted from Note '%s' and was untouched. Adjust saved_state file", noteID)
			savedStateChange["needChange"] = "delete"
		} else {
			system.DebugLog("But parameter reported as deleted from Note '%s'. Mismatch, do nothing.", key, noteID)
		}
	}
	if param["newParam"].(bool) {
		system.DebugLog("But parameter is new in Note '%s'.", noteID)
		savedStateChange["needChange"] = "add"
	}
	if param["changedParam"].(bool) {
		system.DebugLog("But parameter reported as changed in Note '%s'. Scary, but add/change parameter anyway.", key, noteID)
		savedStateChange["needChange"] = "changed"
	}
	if param["newParam"].(bool) || param["changedParam"].(bool) {
		savedStateChange["noteID"] = comparison.ActualValue.(string)
		if fileState == "noFile" {
			// new or changed parameter, create parameter file
			system.DebugLog("Create parameter file for parameter '$s'.", key)
			note.CreateParameterStartValues(key, comparison.ActualValue.(string))
		}
		if param["isUntouched"].(bool) {
			return needApply
		}
		// add values to the parameter file (not, if untouched)
		system.DebugLog("Add Note to parameter file.", noteID, key, noteID)
		note.AddParameterNoteValues(key, comparison.ExpectedValue.(string), noteID, "add")
		needApply = true
	}
	return needApply
}

// paramFileNoNoteID handles the situation that the parameter file of the
// changed parameter does NOT contain the related noteID
func paramFileNoNoteID(param map[string]interface{}, comparison note.FieldComparison, needApply bool, savedStateChange map[string]string) bool {
	earlyReturn := false
	noteID := param["noteID"].(string)
	key := param["key"].(string)
	idx := param["idx"].(int)
	if param["deletedParam"].(bool) {
		if param["wasUntouched"].(bool) {
			system.DebugLog("Note ID '%s' is not available in parameter file of '%s', but parameter reported as deleted from Note '%s' and was untouched. Adjust saved_state file", noteID, key, noteID)
			savedStateChange["needChange"] = "delete"
		} else {
			system.DebugLog("Note ID '%s' is not available in parameter file of '%s', but parameter reported as deleted from Note '%s'. Mismatch, do nothing.", noteID, key, noteID)
		}
		earlyReturn = true
	}
	if idx < 0 {
		system.DebugLog("Seems index '%+v' for 'insert' operation was not correctly set. Do nothing.", idx)
		earlyReturn = true
	}
	if param["override"].(bool) && !param["changedInOver"].(bool) {
		system.DebugLog("parameter is available in override file and is NOT changed, so nothing to apply or to change in parameter and saved_state file")
		earlyReturn = true
	}
	if earlyReturn {
		return needApply
	}

	if param["isUntouched"].(bool) {
		if param["changedParam"].(bool) {
			system.DebugLog("parameter is untouched, but changed.")
			needApply = true
		} else {
			system.DebugLog("parameter is untouched, but not changed. Nothing to apply or to insert in parameter file. Only saved_state handling needed")
			savedStateChange["needChange"] = "add"
		}
		return needApply
	}
	if param["newParam"].(bool) {
		system.DebugLog("Note ID '%s' is not available in parameter file of '%s', but parameter is new in Note '%s'. So add Note to parameter file.", noteID, key, noteID)
		savedStateChange["needChange"] = "add"
	}
	if param["changedParam"].(bool) {
		system.DebugLog("Note ID '%s' is not available in parameter file of '%s', but parameter is reported as changed in Note '%s'. Scary, but add/change parameter anyway.", noteID, key, noteID)
		savedStateChange["needChange"] = "changed"
	}
	savedStateChange["noteID"] = comparison.ActualValue.(string)
	if param["newParam"].(bool) || param["changedParam"].(bool) {
		// add new parameter entry
		if idx > 0 {
			// found successor Note in parameter file
			// insert noteID at index position
			// no apply needed, as value from successor
			// note is still valid
			// remaining savedStateChange handling later
			// in the calling function
			system.DebugLog("Note ID '%s' is not available in parameter file of '%s', insert Note to parameter file at position '%v'.", noteID, key, idx)
			note.AddParameterNoteValues(key, comparison.ExpectedValue.(string), noteID, strconv.Itoa(idx))
		} else {
			// index is '0', no successor Note found in
			// parameter file
			// or noteID is last Note in NoteApplyOrder
			// append to end of parameter file
			// apply needed
			system.DebugLog("Note ID '%s' is not available in parameter file of '%s', append Note to parameter file.", noteID, key)
			needApply = true
			note.AddParameterNoteValues(key, comparison.ExpectedValue.(string), noteID, "add")
		}
	}
	return needApply
}

// paramFileNoteIDAvail handles the situation that the parameter file of the
// changed parameter already contains the note ID
func paramFileNoteIDAvail(param map[string]interface{}, comparison note.FieldComparison, needApply bool, savedStateChange map[string]string) bool {
	system.DebugLog("paramFileNoteIDAvail - param is '%+v', comparison is '%+v', needApply is '%+v', savedStateChange is '%+v'", param, comparison, needApply, savedStateChange)
	noteID := param["noteID"].(string)
	key := param["key"].(string)

	if param["newParam"].(bool) {
		system.DebugLog("Note ID '%s' is reported as available in parameter file of '%s', but is a new parameter in Note '%s'. Mismatch, do nothing.", noteID, key, noteID)
		// ANGI TODO - discuss, if returning an error, do nothing (only Log message) or simply change the entry in parameter file (handle similar to 'changedParam')
	}

	if param["deletedParam"].(bool) {
		// remove parameter entry
		system.DebugLog("Note ID '%s' available in parameter file of '%s', but deleted in Note '%s'. So delete Note from parameter file and saved_state file.", noteID, key, noteID)
		//_, _ := note.RevertParameter(key, noteID)
		savedStateChange["needChange"] = "delete"

		_, _ = note.RevertParameter(key, noteID)

		if param["isLastNote"].(bool) {
			// need to apply the value of the predecessor noteID
			system.DebugLog("Note ID '%s' is last Note in parameter file of '%s'. Need to apply value '%s' of predecessor note '%s'.", noteID, key, param["noteChainPreValue"].(string), param["noteChainPreID"].(string))
			needApply = true
			// change content of 'refreshed' data structure for later apply
			if param["noteChainPreValue"].(string) != "" {
				err := updateVendInfo(noteID, key, param["noteChainPreValue"].(string), param["section"].(string))
				if err != nil {
					system.DebugLog("Failed to update vend information for 'refreshed' data structur - %v", err)
				}
			}
			// no need for additional savedState handling, no successor note
		} else {
			// not last note, so NO apply needed
			// but savedState file of successor note needs changes
			system.DebugLog("Note ID '%s' is NOT the last Note in parameter file of '%s'. No apply needed, but adjust save_state file of successor Note '%s' with '%s'.", noteID, key, param["noteChainPostID"].(string), param["noteChainPreValue"].(string))
			savedStateChange["postID"] = param["noteChainPostID"].(string)
			savedStateChange["postValue"] = param["noteChainPreValue"].(string)
		}
	}
	if param["override"].(bool) && !param["changedInOver"].(bool) {
		system.DebugLog("parameter is available in override file and is NOT changed, so nothing to apply or to change in parameter and saved_state file")
		return needApply
	}
	if param["changedParam"].(bool) {
		exchgValue := comparison.ExpectedValue.(string)
		if param["isUntouched"].(bool) {
			exchgValue = param["noteChainPreValue"].(string)

			system.DebugLog("Note ID '%s' available in parameter file of '%s', but untouched in Note '%s'. So remove Note from parameter file.", noteID, key, noteID)
			// remove parameter entry
			_, _ = note.RevertParameter(key, noteID)
		} else {
			// change parameter entry
			system.DebugLog("Note ID '%s' available in parameter file of '%s', but changed in Note '%s'. So change Note value in parameter file.", noteID, key, noteID)
			note.AddParameterNoteValues(key, comparison.ExpectedValue.(string), noteID, "change")
		}
		if param["isUntouched"].(bool) && param["isLastNote"].(bool) {
			// need to apply the value of the predecessor noteID
			system.DebugLog("Note ID '%s' is last Note in parameter file of '%s', but is untouched. Need to apply value '%s' of predecessor note '%s'.", noteID, key, param["noteChainPreValue"].(string), param["noteChainPreID"].(string))
			// change content of 'refreshed' data structure for later apply
			if param["noteChainPreValue"].(string) != "" {
				err := updateVendInfo(noteID, key, exchgValue, param["section"].(string))
				if err != nil {
					system.DebugLog("Failed to update vend information for 'refreshed' data structur - %v", err)
				}
			}
		}
		if param["isLastNote"].(bool) {
			// no change of saved_state file from noteID needed
			// apply needed
			system.DebugLog("Note ID '%s' is last Note in parameter file of '%s'. No change of saved_state file needed, but need to apply changed value '%s'.", noteID, key, exchgValue)
			needApply = true
			// no need for additional savedState handling, no successor note
		} else {
			savedStateChange["needChange"] = "changed"
			// not last note, so NO apply needed
			// but savedState file of successor note needs changes
			system.DebugLog("Note ID '%s' is NOT the last Note in parameter file of '%s'. No apply needed, but adjust save_state file of successor Note '%s' with '%s'.", noteID, key, param["noteChainPostID"].(string), exchgValue)
			savedStateChange["postID"] = param["noteChainPostID"].(string)
			savedStateChange["postValue"] = exchgValue
		}
	}
	return needApply
}

// adjustSectionFile will adjust the parameter value in the section file
// the section file /run/saptune/sections/<NoteID>.sections contains the
// parameters and values from the note definition file, created during
// 'note apply' to ensure that a revert of the note values will work even that
// the note definition file was removed. Will be removed during 'note revert'
func adjustSectionFile(noteID string, changedParameter map[string]map[string]interface{}) error {
	// read the current stored section information
	sectCont, secterr := txtparser.GetSectionInfo("rosi", noteID, false)
	if secterr != nil {
		system.ErrorLog("Problems while getting section information of note '%s'. - %v", noteID, secterr)
		return secterr
	}

	for key, param := range changedParameter {
		section := param["section"].(string)
		entry := txtparser.INIEntry{
			Section:  section,
			Key:      key,
			Operator: param["nOp"].(txtparser.Operator),
			Value:    param["nValue"].(string),
		}
		if param["deletedParam"].(bool) {
			// delete parameter entry
			idx := positionInList(entry, sectCont.AllValues)
			if idx < 0 {
				system.ErrorLog("deleted parameter not found in section file. Can't change the value. Ignoring...")
			} else {
				sectCont.AllValues = append(sectCont.AllValues[:idx], sectCont.AllValues[idx+1:]...)
				delete(sectCont.KeyValue[section], key)
			}
			// ANGI TODO - check, if last key in 'section' and then
			// remove section too
			// delete(sectCont.KeyValue, section)
		}
		if param["newParam"].(bool) {
			// add new parameter entry
			sectCont.AllValues = append(sectCont.AllValues, entry)
			if _, ok := sectCont.KeyValue[section]; !ok {
				// add new section
				entryMap := make(map[string]txtparser.INIEntry)
				entryMap[entry.Key] = entry
				sectCont.KeyValue[section] = entryMap
			} else {
				sectCont.KeyValue[section][key] = entry
			}
		}
		if param["changedParam"].(bool) && param["changedInNote"].(bool) {
			// change parameter entry
			idx := positionInList(entry, sectCont.AllValues)
			if idx < 0 {
				system.ErrorLog("changed parameter not found in section file. Can't change the value. Ignoring...")
			} else {
				sectCont.AllValues[idx] = entry
				sectCont.KeyValue[section][key] = entry
			}
		}
		if param["changedParam"].(bool) && !param["changedInNote"].(bool) {
			system.DebugLog("parameter not changed in Note, so no need to change the section file entry")
		}
	}
	// override stored section info with the adjusted information
	err := txtparser.StoreSectionInfo(sectCont, "section", noteID, true)
	if err != nil {
		system.ErrorLog("Problems during storing of section information")
	}
	return err
}

// updateVendInfo updates the stored 'vend' info for later apply
func updateVendInfo(noteID, key, value, section string) error {
	// ANGI TODO: override handling
	// don't forget override - refreshed.OverrideParams[..] = ..

	// add 'start value' of 'deleted' parameter to vend data to get the
	// Apply working
	refreshed, err := note.GetVendInfo("vend", noteID)
	// ANGI TODO error handling
	refreshed.(note.INISettings).SysctlParams[key] = value
	err = txtparser.StoreSectionInfo(refreshed, "vend", noteID, true)
	// ANGI TODO error handling

	// add 'deleted' parameter back to ini.AllValues to get Apply working
	refreshAV, _ := txtparser.GetSectionInfo("del", noteID, false)

	entriesArray := make([]txtparser.INIEntry, 0, 8)
	entriesMap := make(map[string]txtparser.INIEntry)
	entry := txtparser.INIEntry{
		Section:  section,
		Key:      key,
		Operator: txtparser.Operator("="),
		Value:    value,
	}
	entriesArray = append(entriesArray, entry)
	entriesMap[entry.Key] = entry

	refreshAV.KeyValue[section] = entriesMap
	refreshAV.AllValues = append(refreshAV.AllValues, entriesArray...)
	err = txtparser.StoreSectionInfo(refreshAV, "del", noteID, true)
	return err
}

// positionInList returns the position of an entry in the given list
func positionInList(entry txtparser.INIEntry, list []txtparser.INIEntry) int {
	for cnt, ent := range list {
		if entry.Section == ent.Section && entry.Key == ent.Key {
			return cnt
		}
	}
	return -1
}

// noteChainSetup creates the note chain for the noteID from the parameter file
// predecessor ID - noteID - successor ID including the parameter values
// and sets 'isLastNote'
func noteChainSetup(paramStateValues note.ParameterNotes, notePosition int, param map[string]interface{}) {
	system.DebugLog("noteChainSetup - paramStateValues is '%+v', notePosition is '%+v', param is '%+v'", paramStateValues, notePosition, param)
	noteID := param["noteID"].(string)
	param["noteChainPreID"] = ""
	param["noteChainPostID"] = ""
	if len(paramStateValues.AllNotes) > 0 {
		if paramStateValues.AllNotes[len(paramStateValues.AllNotes)-1].NoteID == noteID {
			// note is on last position of existing parameter file
			param["isLastNote"] = true
		}
		// store predecessor and successor note ID and parameter value
		// for later use (saved_state file adjustment)
		pre := notePosition - 1
		post := notePosition + 1
		if pre >= 0 {
			param["noteChainPreID"] = paramStateValues.AllNotes[pre].NoteID
			param["noteChainPreValue"] = paramStateValues.AllNotes[pre].Value
		}
		if post <= len(paramStateValues.AllNotes)-1 {
			param["noteChainPostID"] = paramStateValues.AllNotes[post].NoteID
			param["noteChainPostValue"] = paramStateValues.AllNotes[post].Value
		}
	}
}

// pNoteInsertPosition returns the index of the postition inside the parameter
// file where to insert the noteID of the newly added parameter
// check NoteApplyOrder for order of Notes
// and check, if a successor Note from NoteApplyOrder is already available
// in the parameter file
func pNoteInsertPosition(noteID string, noteApplyOrder []string, paramStateValues note.ParameterNotes) int {
	index := 0
	searchNext := false
	for _, note := range noteApplyOrder {
		if note == noteID {
			searchNext = true
			continue
		}
		if searchNext {
			//index = paramStateValues.PositionInParameterList(noteID)
			index = paramStateValues.PositionInParameterList(note)
			// index == 0 (note ID not in parameter file, no file
			// or only 'start' in file)
			if index > 0 {
				// found successor note ID in parameter file
				// stop searching for insert position
				break
			}
		}
	}
	system.DebugLog("pNoteInsertPosition - noteID is '%s', noteApplyOrder is '%+v', paramStateValues is '%+v', index is '%+v'", noteID, noteApplyOrder, paramStateValues, index)
	return index
}
