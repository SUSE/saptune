/*
Solutions are collections of relevant SAP notes, all of which are applicable to specific SAP products.

A system can be tuned for more than one solutions at a time.
*/
package solution

import (
	"sort"
)

type Solution []string // Solution is identified by set of note numbers.

var AllSolutions = map[string]Solution{
	"HANA":             Solution{"1275776", "1984787", "1557506", "2205917"},
	"NETWEAVER":        Solution{"1275776", "1984787", "1557506"},
	"MAXDB":            Solution{"1275776", "1984787", "1557506"},
	"S4HANA-APPSERVER": Solution{"1275776", "1984787", "1557506"},            // identical to Netweaver
	"S4HANA-DBSERVER":  Solution{"1275776", "1984787", "1557506", "2205917"}, // identical to HANA
} // Solution ID VS note numbers

// Return all solution names, sorted alphabetically.
func GetSortedSolutionNames() (ret []string) {
	ret = make([]string, 0, len(AllSolutions))
	for id := range AllSolutions {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return
}
