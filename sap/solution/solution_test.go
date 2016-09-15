package solution

import "testing"

func TestGetSortedSolutionIDs(t *testing.T) {
	if len(GetSortedSolutionNames()) != len(AllSolutions) {
		t.Fatal(GetSortedSolutionNames())
	}
}
