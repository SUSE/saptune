/*
Solutions are collections of relevant SAP notes, all of which are applicable to specific SAP products.

A system can be tuned for more than one solutions at a time.
*/
package solution

import (
	"sort"
)

const (
	ArchX86     = "amd64"   // ArchX86 is the GOARCH value for x86 platform.
	ArchPPC64LE = "ppc64le" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform.
	ArchX86_PC     = "amd64_PC"   // ArchX86 is the GOARCH value for x86 platform. _PC indicates PageCache is available
	ArchPPC64LE_PC = "ppc64le_PC" // ArchPPC64LE is the GOARCH for 64-bit PowerPC little endian platform. _PC indicates PageCache is available
)

type Solution []string // Solution is identified by set of note numbers.

var AllSolutions = map[string]map[string]Solution{
	ArchX86: {
		"BOBJ":             {"1275776", "1984787", "SAP_BOBJ"},
		"SAP-ASE":          {"1275776", "1984787", "SAP_ASE"},
		"HANA":             {"1275776", "1984787", "2205917"},
		"NETWEAVER":        {"1275776", "1984787"},
		"MAXDB":            {"1275776", "1984787"},
		"S4HANA-APPSERVER": {"1275776", "1984787"},            // identical to Netweaver
		"S4HANA-DBSERVER":  {"1275776", "1984787", "2205917"}, // identical to HANA
	},
	ArchPPC64LE: {
		"HANA":             {"1275776", "1984787", "2205917"},
		"NETWEAVER":        {"1275776", "1984787"},
		"MAXDB":            {"1275776", "1984787"},
		"S4HANA-APPSERVER": {"1275776", "1984787"},            // identical to Netweaver
		"S4HANA-DBSERVER":  {"1275776", "1984787", "2205917"}, // identical to HANA
	},
	ArchX86_PC: {
		"BOBJ":             {"1275776", "1984787", "1557506", "SAP_BOBJ"},
		"SAP-ASE":          {"1275776", "1984787", "1557506", "SAP_ASE"},
		"HANA":             {"1275776", "1984787", "1557506", "2205917"},
		"NETWEAVER":        {"1275776", "1984787", "1557506"},
		"MAXDB":            {"1275776", "1984787", "1557506"},
		"S4HANA-APPSERVER": {"1275776", "1984787", "1557506"},            // identical to Netweaver
		"S4HANA-DBSERVER":  {"1275776", "1984787", "1557506", "2205917"}, // identical to HANA
	},
	ArchPPC64LE_PC: {
		"HANA":             {"1275776", "1984787", "1557506", "2205917"},
		"NETWEAVER":        {"1275776", "1984787", "1557506"},
		"MAXDB":            {"1275776", "1984787", "1557506"},
		"S4HANA-APPSERVER": {"1275776", "1984787", "1557506"},            // identical to Netweaver
		"S4HANA-DBSERVER":  {"1275776", "1984787", "1557506", "2205917"}, // identical to HANA
	},
} // Architecture VS solution ID VS note numbers

// Return all solution names, sorted alphabetically.
func GetSortedSolutionNames(archName string) (ret []string) {
	ret = make([]string, 0, len(AllSolutions))
	for id := range AllSolutions[archName] {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}
