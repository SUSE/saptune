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
	ShippedSolSheets       = "/usr/share/saptune/sols/"
	OverrideSolutionSheets = "/etc/saptune/override/"
	DeprecSolutionSheets   = "/usr/share/saptune/deprecated/"
	SolutionSheets         = "/var/lib/saptune/working/sols/"
	NoteTuningSheets       = "/var/lib/saptune/working/notes/"
	ExtraTuningSheets      = "/etc/saptune/extra/"
	ArchX86                = "amd64"      // ArchX86 is the GOARCH value for x86 platform.
	ArchPPC64LE            = "ppc64le"    // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform.
	ArchX86PC              = "amd64_PC"   // ArchX86 is the GOARCH value for x86 platform. PC indicates PageCache is available
	ArchPPC64LEPC          = "ppc64le_PC" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform. PC indicates PageCache is available
)

// Solution is identified by set of note numbers.
type Solution []string

// Architecture VS solution ID VS note numbers
// AllSolutions = map[string]map[string]Solution

// OverrideSolutions contains a list of all available override solutions with
// their related SAP Notes for all supported architectures
var OverrideSolutions = GetOtherSolution(OverrideSolutionSheets, NoteTuningSheets, ExtraTuningSheets)

// CustomSolutions contains a list of all available customer specific solutions
// with their related SAP Notes for all supported architectures
var CustomSolutions = GetOtherSolution(ExtraTuningSheets, NoteTuningSheets, ExtraTuningSheets)

// DeprecSolutions contains a list of all solutions witch are deprecated
var DeprecSolutions = GetOtherSolution(DeprecSolutionSheets, "", "")

// AllSolutions contains a list of all available solutions with their related
// SAP Notes for all supported architectures
var AllSolutions = GetSolutionDefintion(SolutionSheets, ExtraTuningSheets, NoteTuningSheets)

// GetSolutionDefintion reads solution definition from file
// build same structure for AllSolutions as before
// can be simplified later
func GetSolutionDefintion(solsDir, extraDir, noteDir string) map[string]map[string]Solution {
	sols := make(map[string]map[string]Solution)
	sol := make(map[string]Solution)
	currentArch := ""
	arch := ""
	pcarch := ""
	solAllVals := getAllSolsFromDir(solsDir, "", "")
	// add custom solutions to the list
	custAllVals := getAllSolsFromDir(extraDir, noteDir, extraDir)
	for _, p := range custAllVals {
		solAllVals = append(solAllVals, p)
	}

	for _, param := range solAllVals {
		if param.Section == "reminder" || param.Section == "version" {
			continue
		}
		if param.Section != "ArchX86" && param.Section != "ArchPPC64LE" {
			// as the function most of the time is called
			// before the logging is initialized use
			// Fprintf instead to give customers a hint.
			fmt.Fprintf(os.Stderr, "Warning: skip unsupported solution section '%s'\n", param.Section)
			//system.WarningLog("skip unsupported solution section '%s'", param.Section)
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

		// looking for override solution
		if len(OverrideSolutions[arch]) != 0 && len(OverrideSolutions[arch][param.Key]) != 0 {
			param.Value = strings.Join(OverrideSolutions[arch][param.Key], "\t")
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
func GetOtherSolution(solsDir, noteFiles, extraFiles string) map[string]map[string]Solution {
	sols := make(map[string]map[string]Solution)
	sol := make(map[string]Solution)
	currentArch := ""
	arch := ""
	pcarch := ""
	extra := false
	if solsDir == ExtraTuningSheets {
		extra = true
	}
	// looking for override or extra solution file
	solAllVals := getAllSolsFromDir(solsDir, noteFiles, extraFiles)

	for _, param := range solAllVals {
		if param.Section == "reminder" || param.Section == "version" {
			continue
		}
		if param.Section != "ArchX86" && param.Section != "ArchPPC64LE" {
			// as the function most of the time is called
			// before the logging is initialized use
			// Fprintf instead to give customers a hint.
			fmt.Fprintf(os.Stderr, "Warning: skip unsupported solution section '%s'\n", param.Section)
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
		// Do not allow customer/vendor to override built-in solutions
		if extra && IsShippedSolution(param.Key) {
			system.WarningLog("extra solution '%s' will not override built-in solution implementation", param.Key)
			continue
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
	newSol := make(map[string]Solution)
	if system.IsPagecacheAvailable() {
		for key, val := range sols[pcarch] {
			newSol[key] = val
		}
		for key, val := range sol {
			newSol[key] = val
		}
		//sols[pcarch] = sol
		if len(newSol) != 0 {
			sols[pcarch] = newSol
		}
		newSol = make(map[string]Solution)
	}
	for key, val := range sols[arch] {
		newSol[key] = val
	}
	for key, val := range sol {
		newSol[key] = val
	}
	//sols[arch] = sol
	if len(newSol) != 0 {
		sols[arch] = newSol
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

// getAllSolsFromDir retrieves all defined solutions from the solution files found in the given
// directory
func getAllSolsFromDir(solsDir, noteFiles, extraFiles string) []txtparser.INIEntry {
	solAllVals := make([]txtparser.INIEntry, 0, 64)
	_, files := system.ListDir(solsDir, "saptune solution definitions")
	for _, fName := range files {
		if strings.HasSuffix(fName, ".conf") {
			// skip custom defined note definition files
			continue
		}
		if fName != "solsdeprecated" && !strings.HasSuffix(fName, ".sol") {
			// silently skip filenames without .sol suffix
			// do not print a warning as in case of override we can not filter out
			// the node definition files as they do not have a suffix.
			continue
		}
		solName := strings.TrimSuffix(fName, ".sol")
		fileName := fmt.Sprintf("%s%s", solsDir, fName)
		content, err := txtparser.ParseINIFile(fileName, false)
		if err != nil {
			// as the function most of the time is called
			// before the logging is initialized use
			// Fprintf instead to give customers a hint.
			fmt.Fprintf(os.Stderr, "Error: Failed to read solution definition from file '%s'\n", fileName)
			continue
		}

		notesOK := true
		for _, param := range content.AllValues {
			param.Key = solName
			if noteFiles != "" {
				// check, if all note files used in the override or custom
				// solution file are available in the working area or in
				// /etc/saptune/extra
				notesOK = checkSolutionNotes(param, fileName, noteFiles, extraFiles)
				if !notesOK {
					// skip solution definition, because one or more notes
					// referenced in the solution definition do not have
					// a note configuration file on the system
					continue
				}
			}
			solAllVals = append(solAllVals, param)
			//solAllVals = append(solAllVals, content.AllValues...)
		}
	}
	return solAllVals
}

// IsAvailableSolution returns true, if the solution name already exists
func IsAvailableSolution(sol, arch string) bool {
	found := false
	for _, solName := range GetSortedSolutionNames(arch) {
		if sol == solName {
			found = true
			break
		}
	}
	return found
}

// IsShippedSolution returns true, if the solution is shipped by the
// saptune package (from /usr/share/saptune/solutions)
func IsShippedSolution(sol string) bool {
	fileName := fmt.Sprintf("%s%s.sol", ShippedSolSheets, sol)
	if _, err := os.Stat(fileName); err == nil {
		return true
	}
	return false
}

// Refresh refreshes the solution related variables
func Refresh() {
	CustomSolutions = GetOtherSolution(ExtraTuningSheets, NoteTuningSheets, ExtraTuningSheets)
	OverrideSolutions = GetOtherSolution(OverrideSolutionSheets, NoteTuningSheets, ExtraTuningSheets)
	AllSolutions = GetSolutionDefintion(SolutionSheets, ExtraTuningSheets, NoteTuningSheets)
}
