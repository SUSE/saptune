/*
Solutions are collections of relevant SAP notes, all of which are applicable to specific SAP products.

A system can be tuned for more than one solutions at a time.
*/
package solution

import (
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	SolutionSheet         = "/usr/share/saptune/solutions"
	OverrideSolutionSheet = "/etc/saptune/override/solutions"
	NoteTuningSheets      = "/usr/share/saptune/notes/"
	ArchX86     = "amd64"   // ArchX86 is the GOARCH value for x86 platform.
	ArchPPC64LE = "ppc64le" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform.
	ArchX86_PC     = "amd64_PC"   // ArchX86 is the GOARCH value for x86 platform. _PC indicates PageCache is available
	ArchPPC64LE_PC = "ppc64le_PC" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform. _PC indicates PageCache is available
)

type Solution []string // Solution is identified by set of note numbers.

// Architecture VS solution ID VS note numbers
// AllSolutions = map[string]map[string]Solution
var AllSolutions = GetSolutionDefintion(SolutionSheet)
var OverrideSolutions = GetOverrideSolution(OverrideSolutionSheet, NoteTuningSheets)

// read solution definition from file
// build same structure for AllSolutions as before
// can be simplyfied later
func GetSolutionDefintion(fileName string) (map[string]map[string]Solution) {
	sols := make(map[string]map[string]Solution)
	sol  := make(map[string]Solution)
	currentArch := ""
	arch := ""
	pcarch := ""
	content, err := txtparser.ParseINIFile(fileName, false)
	if err != nil {
		log.Printf("Failed to read solution definition from file '%s'", fileName)
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
		sol[param.Key] = strings.Split(param.Value,"\t")

		if system.IsPagecacheAvailable() {
			if len(OverrideSolutions[pcarch]) != 0 && len(OverrideSolutions[pcarch][param.Key]) != 0 {
				// in case of an override solution do not attach sap note 1557506
			} else {
				//add 1557506 (pagecache note) to list of notes
				sol[param.Key] = append(sol[param.Key], "1557506")
			}
		}
	}
	switch currentArch {
	case "ArchPPC64LE":
		if system.IsPagecacheAvailable() {
			sols[ArchPPC64LE_PC] = sol
		}
		sols[ArchPPC64LE] = sol
	case "ArchX86":
		if system.IsPagecacheAvailable() {
			sols[ArchX86_PC] = sol
		}
		sols[ArchX86] = sol
	}
	return sols
}

// read solution override definition from file
// build same structure for AllSolutions as before
// can be simplyfied later
func GetOverrideSolution(fileName, noteFiles string) (map[string]map[string]Solution) {
	sols := make(map[string]map[string]Solution)
	sol  := make(map[string]Solution)
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
		for _, noteID := range strings.Split(content.KeyValue[param.Section][param.Key].Value,"\t") {
			if _, err := os.Stat(fmt.Sprintf("%s%s", noteFiles, noteID)); err != nil {
				log.Printf("Definition for note '%s' used for solution '%s' in override file '%s' not found in %s", noteID, param.Key, fileName, noteFiles)
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
		sol[param.Key] = strings.Split(param.Value,"\t")
	}
	switch currentArch {
	case "ArchPPC64LE":
		if system.IsPagecacheAvailable() {
			sols[ArchPPC64LE_PC] = sol
		}
		sols[ArchPPC64LE] = sol
	case "ArchX86":
		if system.IsPagecacheAvailable() {
			sols[ArchX86_PC] = sol
		}
		sols[ArchX86] = sol
	}
	return sols
}

// Return all solution names, sorted alphabetically.
func GetSortedSolutionNames(archName string) (ret []string) {
	ret = make([]string, 0, len(AllSolutions))
	for id := range AllSolutions[archName] {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}
