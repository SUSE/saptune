/*
Solutions are collections of relevant SAP notes, all of which are applicable to specific SAP products.

A system can be tuned for more than one solutions at a time.
*/
package solution

import (
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"log"
	"sort"
	"strings"
)

const (
	SolutionSheet         = "/usr/share/saptune/solutions"
	ArchX86     = "amd64"   // ArchX86 is the GOARCH value for x86 platform.
	ArchPPC64LE = "ppc64le" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform.
	ArchX86_PC     = "amd64_PC"   // ArchX86 is the GOARCH value for x86 platform. _PC indicates PageCache is available
	ArchPPC64LE_PC = "ppc64le_PC" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform. _PC indicates PageCache is available
)

type Solution []string // Solution is identified by set of note numbers.

// Architecture VS solution ID VS note numbers
// AllSolutions = map[string]map[string]Solution
var AllSolutions = GetSolutionDefintion()

// read solution definition from file
// build same structure for AllSolutions as before
// can be simplyfied later
func GetSolutionDefintion() (map[string]map[string]Solution) {
	sols := make(map[string]map[string]Solution)
	sol  := make(map[string]Solution)
	currentArch := ""
	content, err := txtparser.ParseINIFile(SolutionSheet, false)
	if err != nil {
		log.Printf("Failed to read solution definition from file '%s'", SolutionSheet)
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
			}
			currentArch = param.Section
			sol = make(map[string]Solution)
		}

		sol[param.Key] = strings.Split(param.Value,"\t")
		if system.IsPagecacheAvailable() {
			//add 1557506 (pagecache note) to list of notes
			sol[param.Key] = append(sol[param.Key], "1557506")
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

// Return all solution names, sorted alphabetically.
func GetSortedSolutionNames(archName string) (ret []string) {
	ret = make([]string, 0, len(AllSolutions))
	for id := range AllSolutions[archName] {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}
