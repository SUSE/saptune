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
	// prepare custom solution and override
	customSolutionFile := path.Join(TstFilesInGOPATH, "saptune-test-custom-sols")
	ovsolutionFile := path.Join(TstFilesInGOPATH, "saptune-test-override-sols")
	noteFiles := TstFilesInGOPATH + "/"
	extraNoteFiles := TstFilesInGOPATH + "/extra/"
	CustomSolutions = GetOtherSolution(customSolutionFile, noteFiles, extraNoteFiles)
	OverrideSolutions = GetOtherSolution(ovsolutionFile, noteFiles, "")

	solutionFile := path.Join(TstFilesInGOPATH, "saptune-test-solutions")
	nwsols := "941735 1771258 1980196 1984787 2534844"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
		nwsols = "941735 1771258 1980196 1984787 2534844"
	}

	solutions := GetSolutionDefintion(solutionFile)
	if len(solutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", solutions, len(solutions))
	}
	if strings.Join(solutions[runtime.GOARCH]["NETW"], " ") != nwsols {
		t.Error(solutions)
	}

	sols := GetSolutionDefintion("/saptune_file_not_avail")
	if len(sols) != 0 {
		t.Error(sols)
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

	//ovsolutions := GetOverrideSolution(ovsolutionFile, noteFiles)
	ovsolutions := GetOtherSolution(ovsolutionFile, noteFiles, "")
	if len(ovsolutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", ovsolutions, len(ovsolutions))
	}
	if strings.Join(ovsolutions[runtime.GOARCH]["HANA"], " ") != hansol {
		t.Error(ovsolutions)
	}

	ovsolutionFile = path.Join(TstFilesInGOPATH, "saptune-test-override-sols-missing-note")
	//ovsolutions = GetOverrideSolution(ovsolutionFile, noteFiles)
	ovsolutions = GetOtherSolution(ovsolutionFile, noteFiles, "")
	if len(ovsolutions) != 0 {
		t.Errorf("'%+v' has len '%+v'\n", ovsolutions, len(ovsolutions))
	}

	//sols := GetOverrideSolution("/saptune_file_not_avail", noteFiles)
	sols := GetOtherSolution("/saptune_file_not_avail", noteFiles, "")
	if len(sols) != 0 {
		t.Error(sols)
	}
}

func TestGetCustomSolution(t *testing.T) {
	customSolutionFile := path.Join(TstFilesInGOPATH, "saptune-test-custom-sols")
	noteFiles := TstFilesInGOPATH + "/"
	extraNoteFiles := TstFilesInGOPATH + "/extra/"

	sol1 := "SOL1NOTE1 NEWSOL1NOTE SOL1NOTE2"
	sol2 := "SOL2NOTE1 NEWSOL2NOTE SOL2NOTE2"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	//customSolutions := GetCustomSolution(customSolutionFile, noteFiles, extraNoteFiles)
	customSolutions := GetOtherSolution(customSolutionFile, noteFiles, extraNoteFiles)
	if len(customSolutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", customSolutions, len(customSolutions))
	}
	if strings.Join(customSolutions[runtime.GOARCH]["NEWSOL1"], " ") != sol1 {
		t.Error(customSolutions)
	}
	if strings.Join(customSolutions[runtime.GOARCH]["NEWSOL2"], " ") != sol2 {
		t.Error(customSolutions)
	}

	customSolutionFile = path.Join(TstFilesInGOPATH, "saptune-test-custom-sols-missing-note")
	//customSolutions = GetCustomSolution(customSolutionFile, noteFiles, extraNoteFiles)
	customSolutions = GetOtherSolution(customSolutionFile, noteFiles, extraNoteFiles)
	if len(customSolutions) != 0 {
		t.Errorf("'%+v' has len '%+v'\n", customSolutions, len(customSolutions))
	}

	//sols := GetCustomSolution("/saptune_file_not_avail", noteFiles, extraNoteFiles)
	sols := GetOtherSolution("/saptune_file_not_avail", noteFiles, extraNoteFiles)
	if len(sols) != 0 {
		t.Error(sols)
	}
}

func TestGetDeprecatedSolution(t *testing.T) {
	deprecSolutionFile := path.Join(TstFilesInGOPATH, "saptune-test-deprecated-sols")
	deprec := "deprecated"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	//solutions := GetDeprecatedSolution(deprecSolutionFile)
	solutions := GetOtherSolution(deprecSolutionFile, "", "")
	if len(solutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", solutions, len(solutions))
	}
	//if solutions[runtime.GOARCH]["MAXDB"] != deprec {
	if strings.Join(solutions[runtime.GOARCH]["MAXDB"], " ") != deprec {
		t.Error(solutions)
	}

	//sols := GetDeprecatedSolution("/saptune_file_not_avail")
	sols := GetOtherSolution("/saptune_file_not_avail", "", "")
	if len(sols) != 0 {
		t.Error(sols)
	}
}

func TestGetSortedSolutionIDs(t *testing.T) {
	if len(GetSortedSolutionNames(runtime.GOARCH)) != len(AllSolutions[runtime.GOARCH]) {
		t.Error(GetSortedSolutionNames(runtime.GOARCH))
	}
}
