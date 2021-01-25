package solution

/*
Solutions are collections of relevant SAP notes, all of which are applicable to specific SAP products.
*/

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"sort"
	"strings"
)

// solution constant definitions
const (
	SolutionSheet         = "/var/lib/saptune/working/solutions"
	OverrideSolutionSheet = "/etc/saptune/override/solutions"
	ExtraSolutionSheet    = "/etc/saptune/extra/solutions"
	DeprecSolutionSheet   = "/usr/share/saptune/solsdeprecated"
	NoteTuningSheets      = "/var/lib/saptune/working/notes/"
	ExtraTuningSheets     = "/etc/saptune/extra/"
	ArchX86               = "amd64"      // ArchX86 is the GOARCH value for x86 platform.
	ArchPPC64LE           = "ppc64le"    // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform.
	ArchX86PC             = "amd64_PC"   // ArchX86 is the GOARCH value for x86 platform. PC indicates PageCache is available
	ArchPPC64LEPC         = "ppc64le_PC" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform. PC indicates PageCache is available
)

// Solution is identified by set of note numbers.
type Solution []string

// Architecture VS solution ID VS note numbers
// AllSolutions = map[string]map[string]Solution

// AllSolutions contains a list of all available solutions with their related
// SAP Notes for all supported architectures
var AllSolutions = GetSolutionDefintion(SolutionSheet)

// OverrideSolutions contains a list of all available override solutions with
// their related SAP Notes for all supported architectures
var OverrideSolutions = GetOtherSolution(OverrideSolutionSheet, NoteTuningSheets, ExtraTuningSheets)

// CustomSolutions contains a list of all available customer specific solutions
// with their related SAP Notes for all supported architectures
var CustomSolutions = GetOtherSolution(ExtraSolutionSheet, NoteTuningSheets, ExtraTuningSheets)

// DeprecSolutions contains a list of all solutions witch are deprecated
var DeprecSolutions = GetOtherSolution(DeprecSolutionSheet, "", "")

// GetSolutionDefintion reads solution definition from file
// build same structure for AllSolutions as before
// can be simplyfied later
func GetSolutionDefintion(fileName string) map[string]map[string]Solution {
	sols := make(map[string]map[string]Solution)
	sol := make(map[string]Solution)
	currentArch := ""
	arch := ""
	pcarch := ""
	content, err := txtparser.ParseINIFile(fileName, false)
	if err != nil {
		_ = system.ErrorLog("Failed to read solution definition from file '%s'", fileName)
		return sols
	}

	for _, param := range content.AllValues {
		if param.Section == "reminder" || param.Section == "version" {
			continue
		}
		if param.Section != currentArch {
			// start a new arch
			if currentArch != "" {
				// save previous arch settings
				if len(CustomSolutions) != 0 {
					// add custom solutions for previous arch
					for cKey, cVal := range CustomSolutions[arch] {
						sol[cKey] = cVal
					}
				}
				sols = storeSols(arch, pcarch, sol, sols)
			}
			currentArch = param.Section
			sol = make(map[string]Solution)
			arch, pcarch = setSolutionArch(currentArch)
		}

		// looking for override solution
		if len(OverrideSolutions[arch]) != 0 && len(OverrideSolutions[arch][param.Key]) != 0 {
			param.Value = strings.Join(OverrideSolutions[arch][param.Key], " ")
		}
		sol[param.Key] = strings.Split(param.Value, "\t")
	}
	if arch != "" {
		// add custom solutions for last arch
		if len(CustomSolutions) != 0 {
			for cKey, cVal := range CustomSolutions[arch] {
				sol[cKey] = cVal
			}
		}
	}
	sols = storeSols(arch, pcarch, sol, sols)
	return sols
}

// GetOtherSolution reads override, custom or deprecated solution definition
// from file
func GetOtherSolution(fileName, noteFiles, extraFiles string) map[string]map[string]Solution {
	sols := make(map[string]map[string]Solution)
	sol := make(map[string]Solution)
	currentArch := ""
	arch := ""
	pcarch := ""
	// looking for override or extra solution file
	content, err := txtparser.ParseINIFile(fileName, false)
	if err != nil {
		return sols
	}

	for _, param := range content.AllValues {
		if noteFiles != "" {
			//check, if all note files used in the override or custom
			// solution file are available in the working area or in
			// /etc/saptune/extra
			notesOK := true
			notesOK = checkSolutionNotes(param, fileName, noteFiles, extraFiles)
			if !notesOK {
				// skip solution definition, because one or more notes
				// referenced in the solution definition do not have
				// a note configuration file on the system
				continue
			}
		}

		if param.Section == "reminder" || param.Section == "version" {
			continue
		}
		if param.Section != currentArch {
			// start a new arch
			if currentArch != "" {
				// save previous arch settings
				sols = storeSols(arch, pcarch, sol, sols)
			}
			currentArch = param.Section
			sol = make(map[string]Solution)
			arch, pcarch = setSolutionArch(currentArch)
		}
		sol[param.Key] = strings.Split(param.Value, "\t")
	}
	if currentArch != "" {
		sols = storeSols(arch, pcarch, sol, sols)
	}
	return sols
}

// checkSolutionNotes checks, if all note files used in the override or custom
// solution file are available in the working area or in /etc/saptune/extra
func checkSolutionNotes(param txtparser.INIEntry, fileName, noteFiles, extraFiles string) bool {
	noteState := true
	// ANGI TODO additional check in /usr/share/saptune/note and WARNING
	// that the working area does not include the needed note for
	// the solution, but the package store (and/or staging area) does.
	for _, noteID := range strings.Split(param.Value, "\t") {
		// first check in the working area
		if _, err := os.Stat(fmt.Sprintf("%s%s", noteFiles, noteID)); err != nil {
			// noteID NOT found in working area
			if extraFiles != "" {
				// check for custom note files
				if _, err := os.Stat(fmt.Sprintf("%s%s.conf", extraFiles, noteID)); err != nil {
					// as the function most of the time is
					// called before the logging is
					// initialized use Fprintf instead to
					// give customers a hint.
					fmt.Fprintf(os.Stderr, "Attention: Definition for note '%s' used for solution '%s' in file '%s' not found in '%s' or '%s'\n", noteID, param.Key, fileName, noteFiles, extraFiles)
					//system.WarningLog("Definition for note '%s' used for solution '%s' in file '%s' not found in %s", noteID, param.Key, fileName, extraFiles)
					noteState = false
				}
			} else {
				// as the function most of the time is called
				// before the logging is initialized use
				// Fprintf instead to give customers a hint.
				fmt.Fprintf(os.Stderr, "Attention: Definition for note '%s' used for solution '%s' in file '%s' not found in '%s'\n", noteID, param.Key, fileName, noteFiles)
				//system.WarningLog("Definition for note '%s' used for solution '%s' in file '%s' not found in %s", noteID, param.Key, fileName, noteFiles)
				noteState = false
			}
		}
	}
	return noteState
}

// setSolutionArch sets arch and pcarch variables regarding the current
// architecture read from the solution file
func setSolutionArch(curArch string) (arch, pcarch string) {
	switch curArch {
	case "ArchPPC64LE":
		arch = "ppc64le"
		pcarch = "ppc64le_PC"
	case "ArchX86":
		arch = "amd64"
		pcarch = "amd64_PC"
	}
	return
}

// storeSols stores the collected solutions in the solution map
// related to the last current architecture read from the solution file
func storeSols(arch, pcarch string, sol map[string]Solution, sols map[string]map[string]Solution) map[string]map[string]Solution {
	if system.IsPagecacheAvailable() {
		sols[pcarch] = sol
	}
	sols[arch] = sol
	return sols
}

// GetSortedSolutionNames returns all solution names, sorted alphabetically.
func GetSortedSolutionNames(archName string) (ret []string) {
	ret = make([]string, 0, len(AllSolutions))
	for id := range AllSolutions[archName] {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}
