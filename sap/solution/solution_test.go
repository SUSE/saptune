package solution

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/")

func TestGetSolutionDefintion(t *testing.T) {
	solutionFile := path.Join(TstFilesInGOPATH, "saptune-test-solutions")
	nwsols := "941735 1771258 1980196 1984787 2534844"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
		nwsols = "941735 1771258 1980196 1984787 2534844"
	}

	solutions := GetSolutionDefintion(solutionFile)
	if len(solutions) != solcount {
		t.Fatalf("'%+v' has len '%+v'\n", solutions, len(solutions))
	}
	if strings.Join(solutions[runtime.GOARCH]["NETW"], " ") != nwsols {
		t.Fatal(solutions)
	}

	sols := GetSolutionDefintion("/saptune_file_not_avail")
	if len(sols) != 0 {
		t.Fatal(sols)
	}
}

func TestGetOverrideSolution(t *testing.T) {
	ovsolutionFile := path.Join(TstFilesInGOPATH, "saptune-test-override-sols")
	noteFiles := TstFilesInGOPATH + "/"

	hansol := "HANA1 NEWNOTE HANA2"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	t.Log(TstFilesInGOPATH)
	ovsolutions := GetOverrideSolution(ovsolutionFile, noteFiles)
	if len(ovsolutions) != solcount {
		t.Fatalf("'%+v' has len '%+v'\n", ovsolutions, len(ovsolutions))
	}
	if strings.Join(ovsolutions[runtime.GOARCH]["HANA"], " ") != hansol {
		t.Fatal(ovsolutions)
	}

	ovsolutionFile = path.Join(TstFilesInGOPATH, "saptune-test-override-sols-missing-note")
	ovsolutions = GetOverrideSolution(ovsolutionFile, noteFiles)
	if len(ovsolutions) != 0 {
		t.Fatalf("'%+v' has len '%+v'\n", ovsolutions, len(ovsolutions))
	}

	sols := GetOverrideSolution("/saptune_file_not_avail", noteFiles)
	if len(sols) != 0 {
		t.Fatal(sols)
	}
}

func TestGetDeprecatedSolution(t *testing.T) {
	deprecSolutionFile := path.Join(TstFilesInGOPATH, "saptune-test-deprecated-sols")
	deprec := "deprecated"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	solutions := GetDeprecatedSolution(deprecSolutionFile)
	if len(solutions) != solcount {
		t.Fatalf("'%+v' has len '%+v'\n", solutions, len(solutions))
	}
	if solutions[runtime.GOARCH]["MAXDB"] != deprec {
		t.Fatal(solutions)
	}

	sols := GetDeprecatedSolution("/saptune_file_not_avail")
	if len(sols) != 0 {
		t.Fatal(sols)
	}
}

func TestGetSortedSolutionIDs(t *testing.T) {
	if len(GetSortedSolutionNames(runtime.GOARCH)) != len(AllSolutions[runtime.GOARCH]) {
		t.Fatal(GetSortedSolutionNames(runtime.GOARCH))
	}
}
