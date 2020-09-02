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
	SolutionSheet         = "/usr/share/saptune/solutions"
	OverrideSolutionSheet = "/etc/saptune/override/solutions"
	DeprecSolutionSheet   = "/usr/share/saptune/solsdeprecated"
	NoteTuningSheets      = "/usr/share/saptune/notes/"
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
var OverrideSolutions = GetOverrideSolution(OverrideSolutionSheet, NoteTuningSheets)

// DeprecSolutions contains a list of all solutions witch are deprecated
var DeprecSolutions = GetDeprecatedSolution(DeprecSolutionSheet)

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
		if param.Section == "reminder" {
			continue
		}
		if param.Section != currentArch {
			// start a new arch
			if currentArch != "" {
				// save previous arch settings
				if system.IsPagecacheAvailable() {
					sols[pcarch] = sol
				}
				sols[arch] = sol
			}
			currentArch = param.Section
			sol = make(map[string]Solution)
			switch currentArch {
			case "ArchPPC64LE":
				arch = "ppc64le"
				pcarch = "ppc64le_PC"
			case "ArchX86":
				arch = "amd64"
				pcarch = "amd64_PC"
			}
		}

		// looking for override solution
		if len(OverrideSolutions[arch]) != 0 && len(OverrideSolutions[arch][param.Key]) != 0 {
			param.Value = strings.Join(OverrideSolutions[arch][param.Key], " ")
		}
		sol[param.Key] = strings.Split(param.Value, "\t")
	}
	switch currentArch {
	case "ArchPPC64LE":
		if system.IsPagecacheAvailable() {
			sols[ArchPPC64LEPC] = sol
		}
		sols[ArchPPC64LE] = sol
	case "ArchX86":
		if system.IsPagecacheAvailable() {
			sols[ArchX86PC] = sol
		}
		sols[ArchX86] = sol
	}
	return sols
}

// GetOverrideSolution reads solution override definition from file
// build same structure for AllSolutions as before
// can be simplyfied later
func GetOverrideSolution(fileName, noteFiles string) map[string]map[string]Solution {
	sols := make(map[string]map[string]Solution)
	sol := make(map[string]Solution)
	currentArch := ""
	arch := ""
	pcarch := ""
	// looking for override file
	content, err := txtparser.ParseINIFile(fileName, false)
	if err != nil {
		return sols
	}

	for _, param := range content.AllValues {
		//check, if all note files used in the override file are available in /usr/share/saptune/note
		notesOK := true
		for _, noteID := range strings.Split(content.KeyValue[param.Section][param.Key].Value, "\t") {
			if _, err := os.Stat(fmt.Sprintf("%s%s", noteFiles, noteID)); err != nil {
				system.WarningLog("Definition for note '%s' used for solution '%s' in override file '%s' not found in %s", noteID, param.Key, fileName, noteFiles)
				notesOK = false
			}
		}
		if !notesOK {
			// skip solution definition, because one or more notes
			// referenced in the solution definition do not have
			// a note configuration file on the system
			continue
		}

		if param.Section == "reminder" {
			continue
		}
		if param.Section != currentArch {
			// start a new arch
			if currentArch != "" {
				// save previous arch settings
				if system.IsPagecacheAvailable() {
					sols[pcarch] = sol
				}
				sols[arch] = sol
			}
			currentArch = param.Section
			sol = make(map[string]Solution)
			switch currentArch {
			case "ArchPPC64LE":
				arch = "ppc64le"
				pcarch = "ppc64le_PC"
			case "ArchX86":
				arch = "amd64"
				pcarch = "amd64_PC"
			}
		}
		sol[param.Key] = strings.Split(param.Value, "\t")
	}
	switch currentArch {
	case "ArchPPC64LE":
		if system.IsPagecacheAvailable() {
			sols[ArchPPC64LEPC] = sol
		}
		sols[ArchPPC64LE] = sol
	case "ArchX86":
		if system.IsPagecacheAvailable() {
			sols[ArchX86PC] = sol
		}
		sols[ArchX86] = sol
	}
	return sols
}

// GetDeprecatedSolution reads solution deprecated definition from file
func GetDeprecatedSolution(fileName string) map[string]map[string]string {
	sols := make(map[string]map[string]string)
	sol := make(map[string]string)
	currentArch := ""
	arch := ""
	pcarch := ""
	// looking for deprecated solution file
	content, err := txtparser.ParseINIFile(fileName, false)
	if err != nil {
		return sols
	}

	for _, param := range content.AllValues {
		if param.Section == "reminder" {
			continue
		}
		if param.Section != currentArch {
			// start a new arch
			if currentArch != "" {
				// save previous arch settings
				if system.IsPagecacheAvailable() {
					sols[pcarch] = sol
				}
				sols[arch] = sol
			}
			currentArch = param.Section
			sol = make(map[string]string)
			switch currentArch {
			case "ArchPPC64LE":
				arch = "ppc64le"
				pcarch = "ppc64le_PC"
			case "ArchX86":
				arch = "amd64"
				pcarch = "amd64_PC"
			}
		}
		sol[param.Key] = param.Value
	}
	switch currentArch {
	case "ArchPPC64LE":
		if system.IsPagecacheAvailable() {
			sols[ArchPPC64LEPC] = sol
		}
		sols[ArchPPC64LE] = sol
	case "ArchX86":
		if system.IsPagecacheAvailable() {
			sols[ArchX86PC] = sol
		}
		sols[ArchX86] = sol
	}
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
