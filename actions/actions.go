package actions

import (
	"bufio"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/sap/note"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"strings"
)

//define constants and variables for the whole package
const (
	SaptuneService       = "saptune.service"
	SapconfService       = "sapconf.service"
	TunedService         = "tuned.service"
	NoteTuningSheets     = "/usr/share/saptune/notes/"
	OverrideTuningSheets = "/etc/saptune/override/"
	ExtraTuningSheets    = "/etc/saptune/extra/" // ExtraTuningSheets is a directory located on file system for external parties to place their tuning option files.
	setGreenText         = "\033[32m"
	setRedText           = "\033[31m"
	resetTextColor       = "\033[0m"
	exitSaptuneStopped   = 1
	exitNotTuned         = 3
	footnote1X86         = "[1] setting is not supported by the system"
	footnote1IBM         = "[1] setting is not relevant for the system"
	footnote2            = "[2] setting is not available on the system"
	footnote3            = "[3] value is only checked, but NOT set"
	footnote4            = "[4] cpu idle state settings differ"
	footnote5            = "[5] expected value does not contain a supported scheduler"
	footnote6            = "[6] grub settings are mostly covered by other settings. See man page saptune-note(5) for details"
	footnote7            = "[7] parameter value is untouched by default"
)

// RPMVersion is the package version from package build process
var RPMVersion = "undef"

// RPMDate is the date of package build
var RPMDate = "undef"

// set 'unsupported' footnote regarding the architecture
var footnote1 = footnote1X86

// Collection of tuning options from SAP notes and 3rd party vendors.
var tuningOptions = note.GetTuningOptions(NoteTuningSheets, ExtraTuningSheets)

// SelectAction selects the choosen action depending on the first command line
// argument
func SelectAction(stApp *app.App, saptuneVers string) {
	switch system.CliArg(1) {
	case "daemon":
		DaemonAction(system.CliArg(2), saptuneVers, stApp)
	case "note":
		NoteAction(system.CliArg(2), system.CliArg(3), system.CliArg(4), stApp)
	case "solution":
		SolutionAction(system.CliArg(2), system.CliArg(3), stApp)
	case "revert":
		RevertAction(os.Stdout, system.CliArg(2), stApp)
	case "service":
		ServiceAction(system.CliArg(2), saptuneVers, stApp)
	default:
		PrintHelpAndExit(1)
	}
}

// RevertAction Revert all notes and solutions
func RevertAction(writer io.Writer, actionName string, tuneApp *app.App) {
	if actionName != "all" {
		PrintHelpAndExit(1)
	}
	fmt.Fprintf(writer, "Reverting all notes and solutions, this may take some time...\n")
	if err := tuneApp.RevertAll(true); err != nil {
		system.ErrorExit("Failed to revert notes: %v", err)
		//panic(err)
	}
	fmt.Fprintf(writer, "Parameters tuned by the notes and solutions have been successfully reverted.\n")
}

// rememberMessage prints a reminder message
func rememberMessage(writer io.Writer) {
	if !system.SystemctlIsRunning("saptune.service") {
		fmt.Fprintf(writer, "\nRemember: if you wish to automatically activate the solution's tuning options after a reboot,"+
			"you must enable and start saptune.service by running:"+
			"\n    saptune service enablestart\n")
	}
}

// VerifyAllParameters Verify that all system parameters do not deviate from any of the enabled solutions/notes.
func VerifyAllParameters(writer io.Writer, tuneApp *app.App) {
	if len(tuneApp.NoteApplyOrder) == 0 {
		fmt.Fprintf(writer, "No notes or solutions enabled, nothing to verify.\n")
	} else {
		unsatisfiedNotes, comparisons, err := tuneApp.VerifyAll()
		if err != nil {
			system.ErrorExit("Failed to inspect the current system: %v", err)
		}
		PrintNoteFields(writer, "NONE", comparisons, true)
		tuneApp.PrintNoteApplyOrder(writer)
		if len(unsatisfiedNotes) == 0 {
			fmt.Fprintf(writer, "The running system is currently well-tuned according to all of the enabled notes.\n")
		} else {
			system.ErrorExit("The parameters listed above have deviated from SAP/SUSE recommendations.")
		}
	}
}

// getFileName returns the corresponding filename of a given noteID
// additional it returns a boolean value which is pointing out that
// the Note is a custom Note (extraNote = true) or an internal one
func getFileName(noteID, NoteTuningSheets, ExtraTuningSheets string) (string, bool) {
	extraNote := false
	fileName := fmt.Sprintf("%s%s", NoteTuningSheets, noteID)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		// Note is NOT an internal Note, but may be a custom Note
		extraNote = true
		_, files := system.ListDir(ExtraTuningSheets, "")
		for _, f := range files {
			if strings.HasPrefix(f, noteID) {
				fileName = fmt.Sprintf("%s%s", ExtraTuningSheets, f)
			}
		}
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			system.ErrorExit("Note %s not found in %s or %s.", noteID, NoteTuningSheets, ExtraTuningSheets)
		} else if err != nil {
			system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
		}
	} else if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", fileName, err)
	}
	return fileName, extraNote
}

// getovFile returns the corresponding override filename of a given noteID
// additional it returns a boolean value which is pointing out if the
// override file already exists (overrideNote = true) or not
func getovFile(noteID, OverrideTuningSheets string) (string, bool) {
	overrideNote := true
	ovFileName := fmt.Sprintf("%s%s", OverrideTuningSheets, noteID)
	if _, err := os.Stat(ovFileName); os.IsNotExist(err) {
		overrideNote = false
	} else if err != nil {
		system.ErrorExit("Failed to read file '%s' - %v", ovFileName, err)
	}
	return ovFileName, overrideNote
}

// readYesNo asks the user for yes/no answer.
// "y", "Y", "yes", "YES", and "Yes" following by "enter" count as confirmation
// "n", "N", "no", "NO", and "No" following by "enter" count as non-confirmation
func readYesNo(s string, in io.Reader, out io.Writer) bool {
	reader := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "%s [y/n]: ", s)
		response, err := reader.ReadString('\n')
		if err != nil {
			system.ErrorExit("Failed to read input: %v", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// renameNote will rename a Note to an new name
func renameNote(fileName, newFileName, ovFileName, newovFileName string, overrideNote, extraNote bool) {
	if overrideNote {
		if err := os.Rename(ovFileName, newovFileName); err != nil {
			system.ErrorExit("Failed to rename file '%s' to '%s' - %v", ovFileName, newovFileName, err)
		}
	}
	if extraNote {
		if err := os.Rename(fileName, newFileName); err != nil {
			system.ErrorExit("Failed to rename file '%s' to '%s' - %v", fileName, newFileName, err)
		}
	}
}

// deleteNote will delete a Note
func deleteNote(fileName, ovFileName string, overrideNote, extraNote bool) {
	if overrideNote {
		if err := os.Remove(ovFileName); err != nil {
			system.ErrorExit("Failed to remove file '%s' - %v", ovFileName, err)
		}
	}
	if extraNote {
		if err := os.Remove(fileName); err != nil {
			system.ErrorExit("Failed to remove file '%s' - %v", fileName, err)
		}
	}
}

// PrintHelpAndExit prints the usage and exit
func PrintHelpAndExit(exitStatus int) {
	fmt.Println(`saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | enable | disable | enablestart | stopdisable ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | enabled ]
  saptune note [ apply | simulate | verify | customise | create | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled ]
  saptune solution [ apply | simulate | verify | revert ] SolutionName
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Print current saptune version:
  saptune version
Print this message:
  saptune help`)
	system.ErrorExit("", exitStatus)
}
