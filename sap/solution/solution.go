/*
Solutions are collections of relevant SAP notes, all of which are applicable to specific SAP products.

A system can be tuned for more than one solutions at a time.
*/
package solution

import (
	"github.com/HouzuoGuo/saptune/sap/note"
	"sort"
)

type Solution []string // Solution is identified by set of note numbers.

var AllSolutions = map[string]map[string]Solution{
	note.ARCH_X86: {
		"BOBJ":             {"1275776", "1984787", "1557506", "SAP_BOBJ"},
		"SAP-ASE":          {"1275776", "1984787", "1557506", "Block", "SAP_ASE"},
		"HANA":             {"1275776", "1984787", "1557506", "2205917"},
		"NETWEAVER":        {"1275776", "1984787", "1557506"},
		"MAXDB":            {"1275776", "1984787", "1557506"},
		"S4HANA-APPSERVER": {"1275776", "1984787", "1557506"},            // identical to Netweaver
		"S4HANA-DBSERVER":  {"1275776", "1984787", "1557506", "2205917"}, // identical to HANA
	},
	note.ARCH_PPC: {
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
