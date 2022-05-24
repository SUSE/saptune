package actions

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/solution"
	"github.com/SUSE/saptune/system"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

var solTemplate = "/usr/share/saptune/SolutionTemplate.conf"

// SolutionAction  Solution actions like apply, revert, verify asm.
func SolutionAction(actionName, solName, newSolName string, tuneApp *app.App) {
	switch actionName {
	case "apply":
		SolutionActionApply(os.Stdout, solName, tuneApp)
	case "list":
		SolutionActionList(os.Stdout, tuneApp)
	case "verify":
		SolutionActionVerify(os.Stdout, solName, tuneApp)
	case "simulate":
		SolutionActionSimulate(os.Stdout, solName, tuneApp)
	case "customise", "customize":
		SolutionActionCustomise(os.Stdout, solName, tuneApp)
	case "edit":
		SolutionActionEdit(os.Stdout, solName, tuneApp)
	case "create":
		SolutionActionCreate(os.Stdout, solName)
	case "show":
		SolutionActionShow(os.Stdout, solName)
	case "delete":
		SolutionActionDelete(os.Stdin, os.Stdout, solName, tuneApp)
	case "rename":
		SolutionActionRename(os.Stdin, os.Stdout, solName, newSolName, tuneApp)
	case "revert":
		SolutionActionRevert(os.Stdout, solName, tuneApp)
	case "applied":
		SolutionActionApplied(os.Stdout, tuneApp)
	case "enabled":
		SolutionActionEnabled(os.Stdout, tuneApp)
	default:
		PrintHelpAndExit(os.Stdout, 1)
	}
}

// SolutionActionApply applies parameter settings defined by the solution
// to the system
func SolutionActionApply(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(writer, 1)
	}
	if len(tuneApp.TuneForSolutions) > 0 {
		// already one solution applied.
		// do not apply another solution. Does not make sense
		system.ErrorExit("There is already one solution applied. Applying another solution is NOT supported.", 1)
	}
	removedAdditionalNotes, err := tuneApp.TuneSolution(solName)
	if err != nil {
		system.ErrorExit("Failed to tune for solution %s: %v", solName, err)
	}
	fmt.Fprintf(writer, "All tuning options for the SAP solution have been applied successfully.\n")
	if len(removedAdditionalNotes) > 0 {
		fmt.Fprintf(writer, "\nThe following previously-enabled notes are now tuned by the SAP solution:\n")
		for _, noteNumber := range removedAdditionalNotes {
			fmt.Fprintf(writer, "\t%s\t%s\n", noteNumber, tuneApp.AllNotes[noteNumber].Name())
		}
	}
	rememberMessage(writer)
}

// SolutionActionList lists all available solution definitions
func SolutionActionList(writer io.Writer, tuneApp *app.App) {
	jsolutionList := []system.JSolListEntry{}
	jsolutionListEntry := system.JSolListEntry{}
	setColor := false
	fmt.Fprintf(writer, "\nAll solutions (* denotes enabled solution, O denotes override file exists for solution, C denotes custom solutions, D denotes deprecated solutions):\n")
	for _, solName := range solution.GetSortedSolutionNames(solutionSelector) {
		jsolutionListEntry = system.JSolListEntry{
			SolName:     "",
			NotesList:   []string{},
			SolEnabled:  false,
			SolOverride: false,
			CustomSol:   false,
			DepSol:      false,
		}
		format := "\t%-18s -"
		if len(solution.OverrideSolutions[solutionSelector][solName]) != 0 {
			// override solution
			format = " O" + format
			jsolutionListEntry.SolOverride = true
		}
		if len(solution.CustomSolutions[solutionSelector][solName]) != 0 {
			// custom solution
			format = " C" + format
			jsolutionListEntry.CustomSol = true
		}
		if _, ok := solution.DeprecSolutions[solutionSelector][solName]; ok {
			// deprecated solution
			format = " D" + format
			jsolutionListEntry.DepSol = true
		}
		if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
			// enabled solution
			format = " " + setGreenText + "*" + format
			jsolutionListEntry.SolEnabled = true
			setColor = true
		}

		solNotes := ""
		for _, noteString := range solution.AllSolutions[solutionSelector][solName] {
			if setColor {
				// notes of an enabled solution
				// check for manually reverted notes
				j := tuneApp.PositionInNoteApplyOrder(noteString)
				if j < 0 {
					// noteID was reverted manually
					solNotes = solNotes + " " + setRedText + setStrikeText + noteString + resetTextColor
				} else {
					solNotes = solNotes + " " + setGreenText + noteString
				}
			} else {
				solNotes = solNotes + " " + noteString
			}
		}
		format = format + solNotes
		if setColor {
			format = format + resetTextColor
			setColor = false
		}
		format = format + "\n"
		//fmt.Printf(format, solName)
		fmt.Fprintf(writer, format, solName)
		jsolutionListEntry.SolName = solName
		jsolutionListEntry.NotesList = solution.AllSolutions[solutionSelector][solName]
		jsolutionList = append(jsolutionList, jsolutionListEntry)
	}
	remember := bytes.Buffer{}
	if system.GetFlagVal("format") == "json" {
		writer = &remember
	}
	rememberMessage(writer)
	result := system.JSolList{
		SolsList: jsolutionList,
		Msg:      remember.String(),
	}
	system.Jcollect(result)
}

// SolutionActionVerify compares all parameter settings from a solution
// definition against the system settings
func SolutionActionVerify(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		VerifyAllParameters(writer, tuneApp)
	} else {
		// Check system parameters against the specified solution, no matter the solution has been tuned for or not.
		unsatisfiedNotes, comparisons, err := tuneApp.VerifySolution(solName)
		if err != nil {
			system.ErrorExit("Failed to test the current system against the specified SAP solution: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true)
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "%s%sThe system fully conforms to the tuning guidelines of the specified SAP solution.%s%s\n", setGreenText, setBoldText, resetBoldText, resetTextColor)
		} else {
			system.ErrorExit("The parameters listed above have deviated from the specified SAP solution recommendations.\n", "colorPrint", setRedText, setBoldText, resetBoldText, resetTextColor)
		}
	}
}

// SolutionActionSimulate shows all changes that will be applied to the system if
// the solution will be applied.
func SolutionActionSimulate(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(writer, 1)
	}
	// Run verify and print out all fields of the note
	if _, comparisons, err := tuneApp.VerifySolution(solName); err != nil {
		system.ErrorExit("Failed to test the current system against the specified note: %v", err)
	} else {
		fmt.Fprintf(writer, "If you run `saptune solution apply %s`, the following changes will be applied to your system:\n", solName)
		PrintNoteFields(writer, "NONE", comparisons, false)
	}
}

// SolutionActionRevert reverts all parameter settings of a solution back to
// the state before 'apply'
func SolutionActionRevert(writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(writer, 1)
	}
	// 'ok' only used to control the log messages
	// call RevertSolution in any case to get the chance of clean up
	_, ok := tuneApp.IsSolutionApplied(solName)
	if err := tuneApp.RevertSolution(solName); err != nil {
		system.ErrorExit("Failed to revert tuning for solution %s: %v", solName, err)
	}
	if ok {
		system.InfoLog("Parameters tuned by the notes referred by the SAP solution have been successfully reverted.")
		fmt.Fprintf(writer, "Parameters tuned by the notes referred by the SAP solution have been successfully reverted.\n")
	} else {
		system.NoticeLog("Solution '%s' is not applied, so nothing to revert.", solName)
	}
}

// SolutionActionEnabled prints out the enabled solution definition
func SolutionActionEnabled(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.TuneForSolutions) != 0 {
		fmt.Fprintf(writer, "%s", tuneApp.TuneForSolutions[0])
	}
	system.Jcollect(tuneApp.TuneForSolutions)
	//system.Jcollect(strings.Join(tuneApp.TuneForSolutions, " "))
}

// SolutionActionApplied prints out the applied solution
func SolutionActionApplied(writer io.Writer, tuneApp *app.App) {
	solName := ""
	if len(tuneApp.TuneForSolutions) != 0 {
		solName = tuneApp.TuneForSolutions[0]
		if state, ok := tuneApp.IsSolutionApplied(solName); ok {
			if state == "partial" {
				fmt.Fprintf(writer, "%s (partial)", solName)
			} else {
				fmt.Fprintf(writer, "%s", solName)
			}
		}
	}
	system.Jcollect(strings.Split(solName, " "))
}

// SolutionActionCustomise creates an override file and allows to editing the
// solution definition override file
func SolutionActionCustomise(writer io.Writer, customSol string, tuneApp *app.App) {
	if customSol == "" {
		PrintHelpAndExit(writer, 1)
	}
	solFName := customSol
	if !strings.HasSuffix(customSol, ".sol") {
		solFName = fmt.Sprintf("%s.sol", customSol)
	} else {
		customSol = strings.TrimSuffix(customSol, ".sol")
	}
	if !solution.IsAvailableSolution(customSol, solutionSelector) {
		system.ErrorExit("Solution '%s' does not exist.", customSol)
	}

	editSrcFile := ""
	editDestFile := ""
	fileName, _ := getFileName(solFName, SolutionSheets, ExtraTuningSheets)
	ovFileName, overrideSol := getovFile(solFName, OverrideTuningSheets)
	if !overrideSol {
		editSrcFile = fileName
		editDestFile = ovFileName
	} else {
		system.NoticeLog("Solution override file already exists, using file '%s' as base for editing", ovFileName)
		editSrcFile = ovFileName
		editDestFile = ovFileName
	}

	changed, err := system.EditAndCheckFile(editSrcFile, editDestFile, customSol, "solution")
	if err != nil {
		system.ErrorExit("Problems while editing solution definition file '%s' - %v", editSrcFile, err)
	}
	if changed {
		// check, if solution is active - applied
		if i := sort.SearchStrings(tuneApp.TuneForSolutions, customSol); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == customSol {
			system.NoticeLog("Your just edited Solution is already applied. To get your changes to take effect, please 'revert' the Solution and apply again.\n")
		} else {
			system.NoticeLog("Do not forget to apply the just edited Solution to get your changes to take effect\n")
		}
	} else {
		system.NoticeLog("Nothing changed during the editor session, so no update of the solution definition file '%s'", editSrcFile)
	}
}

// SolutionActionEdit allows to editing the custom/vendor specific
// solution definition file and NOT the override file
func SolutionActionEdit(writer io.Writer, customSol string, tuneApp *app.App) {
	if customSol == "" {
		PrintHelpAndExit(writer, 1)
	}
	solFName := customSol
	if !strings.HasSuffix(customSol, ".sol") {
		solFName = fmt.Sprintf("%s.sol", customSol)
	} else {
		customSol = strings.TrimSuffix(customSol, ".sol")
	}
	if !solution.IsAvailableSolution(customSol, solutionSelector) {
		system.ErrorExit("Solution '%s' does not exist.", customSol)
	}

	fileName, extraSol := getFileName(solFName, SolutionSheets, ExtraTuningSheets)
	ovFileName, overrideSol := getovFile(solFName, OverrideTuningSheets)
	if !extraSol {
		system.ErrorExit("The Solution definition file you want to edit is a saptune internal (shipped) Solution and can NOT be edited. Use 'saptune solution customise' instead. Exiting ...")
	}

	changed, err := system.EditAndCheckFile(fileName, fileName, customSol, "solution")
	if err != nil {
		system.ErrorExit("Problems while editing Solution definition file '%s' - %v", fileName, err)
	}
	if changed {
		// check, if solution is active - applied
		if i := sort.SearchStrings(tuneApp.TuneForSolutions, customSol); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == customSol {
			system.NoticeLog("Your just edited Solution is already applied. To get your changes to take effect, please 'revert' the Solution and apply again.\n")
		} else {
			system.NoticeLog("Do not forget to apply the just edited Solution to get your changes to take effect\n")
		}
		if overrideSol {
			system.NoticeLog("Solution override file '%s' exists. Please check, if the content of this file is still valid", ovFileName)
		}
	} else {
		system.NoticeLog("Nothing changed during the editor session, so no update of the solution definition file '%s'", fileName)
	}
}

// SolutionActionCreate helps the customer to create an own solution definition
func SolutionActionCreate(writer io.Writer, customSol string) {
	fileName := ""
	if customSol == "" {
		PrintHelpAndExit(writer, 1)
	}
	if !strings.HasSuffix(customSol, ".sol") {
		fileName = fmt.Sprintf("%s%s.sol", ExtraTuningSheets, customSol)
	} else {
		fileName = fmt.Sprintf("%s%s", ExtraTuningSheets, customSol)
		customSol = strings.TrimSuffix(customSol, ".sol")

	}
	if solution.IsShippedSolution(customSol) {
		system.ErrorExit("Solution name '%s' already in use for a solution definition shipped by saptune. Please use another name for your solution definition.", customSol)
	}
	if solution.IsAvailableSolution(customSol, solutionSelector) {
		system.ErrorExit("Solution name '%s' already exists in %s. Please use 'saptune solution edit %s' instead to edit the custom solution file or choose another filename for your solution definition.", fileName, ExtraTuningSheets, customSol)
	}
	if _, err := os.Stat(fileName); err == nil {
		system.ErrorExit("File for solution definition '%s' already exists in %s. Please use 'saptune solution edit %s' instead to edit the custom solution file or choose another filename for your solution definition.", fileName, ExtraTuningSheets, customSol)
	}
	changed, err := system.EditAndCheckFile(solTemplate, fileName, customSol, "solution")
	if err != nil {
		system.ErrorExit("Problems while editing solution definition file '%s' - %v", fileName, err)
	}
	if !changed {
		system.NoticeLog("Nothing changed during the editor session, so no new, custom specific solution definition file will be created.")
	} else {
		system.NoticeLog("Solution '%s' created successfully. You can modify the content of your Solution definition file by using 'saptune solution edit %s' or create an override file by 'saptune solution customise %s'.", customSol, customSol, customSol)
	}
}

// SolutionActionShow shows the content of the Solution definition file
func SolutionActionShow(writer io.Writer, solName string) {
	if solName == "" {
		PrintHelpAndExit(writer, 1)
	}
	// check if solution really exists
	if !solution.IsAvailableSolution(solName, solutionSelector) {
		system.NoticeLog("Solution '%s' does not exist. Nothing to do.", solName)
		system.ErrorExit("", 0)
	}
	solFName := fmt.Sprintf("%s.sol", solName)
	fileName, _ := getFileName(solFName, SolutionSheets, ExtraTuningSheets)
	cont, err := ioutil.ReadFile(fileName)
	if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
	}
	fmt.Fprintf(writer, "\nContent of Solution %s:\n%s\n", solName, string(cont))
}

// SolutionActionDelete deletes a custom solution definition file and
// the corresponding override file
func SolutionActionDelete(reader io.Reader, writer io.Writer, solName string, tuneApp *app.App) {
	if solName == "" {
		PrintHelpAndExit(writer, 1)
	}
	// check if solution really exists
	if !solution.IsAvailableSolution(solName, solutionSelector) {
		system.NoticeLog("Solution '%s' does not exist. Nothing to do.", solName)
		system.ErrorExit("", 0)
	}
	solFName := fmt.Sprintf("%s.sol", solName)
	txtConfirm := fmt.Sprintf("Do you really want to delete Solution '%s'?", solName)
	fileName, extraSol := getFileName(solFName, SolutionSheets, ExtraTuningSheets)
	ovFileName, overrideSol := getovFile(solFName, OverrideTuningSheets)

	// check, if solution is active - applied
	if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
		system.ErrorExit("The Solution file you want to delete is currently in use, which means the Solution is already applied.\nSo please 'revert' the Solution first and then try deleting again.")
	}

	if !extraSol && !overrideSol {
		system.ErrorExit("The Solution file you want to delete is a saptune internal (shipped) Solution and can NOT be deleted. Exiting ...")
	}
	if !extraSol && overrideSol {
		// system solution, override file exists
		txtConfirm = fmt.Sprintf("Solution to delete is a saptune internal (shipped) Solution, so it can NOT be deleted. But an override file for the Solution exists.\nDo you want to remove the override file for Solution %s?", solName)
	}
	if extraSol && overrideSol {
		// custom solution with override file
		txtConfirm = fmt.Sprintf("Solution to delete is a customer/vendor specific Solution and an override file for the Solution exists.\nDo you want to remove the override file for Solution %s?", solName)
	}
	if overrideSol {
		// remove override file
		if readYesNo(txtConfirm, reader, writer) {
			deleteDefFile(ovFileName)
		}
	}
	if extraSol {
		// custom solution
		txtConfirm = fmt.Sprintf("Solution to delete is a customer/vendor specific Solution.\nDo you really want to delete this Solution '%s'?", solName)
		// remove customer/vendor specific solution definition file
		if readYesNo(txtConfirm, reader, writer) {
			deleteDefFile(fileName)
		}
	}
}

// SolutionActionRename renames a custom Solution definition file and
// the corresponding override file
func SolutionActionRename(reader io.Reader, writer io.Writer, solName, newSolName string, tuneApp *app.App) {
	if solName == "" || newSolName == "" {
		PrintHelpAndExit(writer, 1)
	}
	// check if old solution name really exists
	if !solution.IsAvailableSolution(solName, solutionSelector) {
		system.NoticeLog("Solution '%s' does not exist. Nothing to do.", solName)
		system.ErrorExit("", 0)
	}
	// check if new solution name already exists
	if solution.IsAvailableSolution(newSolName, solutionSelector) {
		system.ErrorExit("The new name '%s' for Solution '%s' already exists, can't rename.", newSolName, solName)
	}
	txtConfirm := fmt.Sprintf("Do you really want to rename Solution '%s' to '%s'?", solName, newSolName)
	solFName := fmt.Sprintf("%s.sol", solName)
	fileName, extraSol := getFileName(solFName, SolutionSheets, ExtraTuningSheets)
	if !extraSol {
		system.ErrorExit("The Solution definition file you want to rename is a saptune internal (shipped) Solution and can NOT be renamed. Exiting ...")
	}
	newFileName := fmt.Sprintf("%s%s.sol", ExtraTuningSheets, newSolName)
	ovFileName, overrideSol := getovFile(solFName, OverrideTuningSheets)
	newovFileName := fmt.Sprintf("%s%s.sol", OverrideTuningSheets, newSolName)

	// check, if solution is active - applied
	if i := sort.SearchStrings(tuneApp.TuneForSolutions, solName); i < len(tuneApp.TuneForSolutions) && tuneApp.TuneForSolutions[i] == solName {
		system.ErrorExit("The Solution definition file you want to rename is currently in use, which means the Solution is already applied.\nSo please 'revert' the Solution first and then try renaming again.")
	}

	if extraSol && overrideSol {
		// custom solution with override file
		txtConfirm = fmt.Sprintf("Solution to rename is a customer/vendor specific Solution.\nDo you really want to rename this Solution '%s' and the corresponding override file to the new name '%s'?", solName, newSolName)
	}
	if extraSol && !overrideSol {
		// custom solution
		txtConfirm = fmt.Sprintf("Solution to rename is a customer/vendor specific Solution.\nDo you really want to rename this Solution '%s' to the new name '%s'?", solName, newSolName)
	}

	if readYesNo(txtConfirm, reader, writer) {
		renameDefFile(fileName, newFileName)
		//rewriteSolName(solName, newSolName, newFileName)
		if overrideSol {
			renameDefFile(ovFileName, newovFileName)
			//rewriteSolName(solName, newSolName, newovFileName)
		}
		//solution.Refresh()
	}
}

// rewriteSolName rewrites the solution name inside the solution definition file
func rewriteSolName(oldName, newName, newFile string) {
	// open source file
	fn, err := os.Open(newFile)
	if err != nil {
		system.ErrorExit("Can not open file '%s' - %v", newFile, err)
	}
	defer fn.Close()
	// create temp file
	tmpfn, err := ioutil.TempFile("", "replace-*")
	if err != nil {
		system.ErrorExit("Unable to create temporary file - %v", err)
	}
	defer tmpfn.Close()
	// replace the solution name while copying from fn to tmpfn
	if err := replaceSolName(fn, tmpfn, oldName, newName); err != nil {
		system.ErrorExit("Unable to rewrite the solution name - %v", err)
	}
	// close files
	if err := tmpfn.Close(); err != nil {
		system.ErrorExit("Problems writing temporary file - %v", err)
	}
	if err := fn.Close(); err != nil {
		system.ErrorExit("Problems closing file '%s' - %v", newFile, err)
	}
	// overwrite source file with the temporary used file
	if err := os.Rename(tmpfn.Name(), newFile); err != nil {
		system.ErrorExit("Cannot copy the temporary used file to ''%s' - %v", newFile, err)
	}
}

// replaceSolName replaces the solution name in the file
func replaceSolName(r io.Reader, w io.Writer, oldSol, newSol string) error {
	var re = regexp.MustCompile(fmt.Sprintf("^%s[[:space:]]*=", oldSol))
	newSName := fmt.Sprintf("%s =", newSol)

	// use scanner to read line by line
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		line = re.ReplaceAllString(line, newSName)
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
	}
	return sc.Err()
}

// getNoteInSol checks, if a Note is part of a Solution
// returns Solution names
func getNoteInSol(tApp *app.App, noteName string) (string, string) {
	noteInSols := ""
	noteInCustomSols := ""
	sols := []string{}
	for sol := range tApp.AllSolutions {
		sols = append(sols, sol)
	}
	sort.Strings(sols)
	for _, sol := range sols {
		for _, noteID := range tApp.AllSolutions[sol] {
			if noteName != noteID {
				continue
			}
			// note is part of solution sol
			if len(noteInSols) == 0 {
				noteInSols = sol
			} else {
				noteInSols = fmt.Sprintf("%s, %s", noteInSols, sol)
			}
			// check for custom solution
			if len(solution.CustomSolutions[solutionSelector][sol]) != 0 {
				// sol is custom solution
				if len(noteInCustomSols) == 0 {
					noteInCustomSols = sol
				} else {
					noteInCustomSols = fmt.Sprintf("%s, %s", noteInCustomSols, sol)
				}
			}
		}
	}
	return noteInSols, noteInCustomSols
}
