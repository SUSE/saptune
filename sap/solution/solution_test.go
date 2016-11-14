package solution

import (
	"runtime"
	"testing"
)

func TestGetSortedSolutionIDs(t *testing.T) {
	if len(GetSortedSolutionNames(runtime.GOARCH)) != len(AllSolutions[runtime.GOARCH]) {
		t.Fatal(GetSortedSolutionNames(runtime.GOARCH))
	}
}
