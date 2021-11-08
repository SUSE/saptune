package solution

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

var SolutionSheetsInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/sols") + "/"
var ExtraFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra") + "/"
var OverTstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/etc/saptune/override") + "/"
var DeprecFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/deprecated") + "/"
var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/")

func TestGetSolutionDefintion(t *testing.T) {
	// prepare custom solution and override
	noteFiles := TstFilesInGOPATH + "/"
	extraNoteFiles := TstFilesInGOPATH + "/extra/"
	CustomSolutions = GetOtherSolution(ExtraFilesInGOPATH, noteFiles, extraNoteFiles)
	OverrideSolutions = GetOtherSolution(OverTstFilesInGOPATH, noteFiles, "")

	nwsols := "941735 1771258 1980196 1984787 2534844"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
		nwsols = "941735 1771258 1980196 1984787 2534844"
	}

	solutions := GetSolutionDefintion(SolutionSheetsInGOPATH, "", "")
	if len(solutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", solutions, len(solutions))
	}
	if strings.Join(solutions[runtime.GOARCH]["NETW"], " ") != nwsols {
		t.Error(solutions)
	}

	sols := GetSolutionDefintion("/saptune_file_not_avail", "", "")
	if len(sols) != 0 {
		t.Error(sols)
	}
}

func TestAvailableShippedSolution(t *testing.T) {
	// BWA, HANA, NETW, MAXDB
	if !IsAvailableSolution("HANA", runtime.GOARCH) {
		t.Error("solution 'HANA' not available")
	}
	if IsAvailableSolution("NoSuchSolution", runtime.GOARCH) {
		t.Error("solution 'NoSuchSolution' reported as available")
	}
	// no file in /usr/share/saptune/sols available in docker container
	// as for now.
	if err := os.MkdirAll(ShippedSolSheets, 0755); err != nil {
		t.Error(err)
	}
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/sols/BWA.sol")
	dst := path.Join(ShippedSolSheets, "BWA.sol")
	err := system.CopyFile(src, dst)
	if err != nil {
		t.Error(err)
	}
	if !IsShippedSolution("BWA") {
		t.Error("shipped solution 'BWA' not available")
	}
	if IsShippedSolution("ANGI") {
		t.Error("solution 'ANGI' reported as shipped solution")
	}
	os.Remove(dst)
	os.RemoveAll(ShippedSolSheets)
}

func TestGetOverrideSolution(t *testing.T) {
	noteFiles := TstFilesInGOPATH + "/"

	hansol := "HANA1 NEWNOTE HANA2"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	ovsolutions := GetOtherSolution(OverTstFilesInGOPATH, noteFiles, "")
	if len(ovsolutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", ovsolutions, len(ovsolutions))
	}
	if strings.Join(ovsolutions[runtime.GOARCH]["HANA"], " ") != hansol {
		t.Error(ovsolutions)
	}

	overSolMissing := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/override-missing") + "/"
	ovsolutions = GetOtherSolution(overSolMissing, noteFiles, "")
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
	noteFiles := TstFilesInGOPATH + "/"
	extraNoteFiles := TstFilesInGOPATH + "/extra/"

	sol1 := "SOL1NOTE1 NEWSOL1NOTE SOL1NOTE2"
	sol2 := "SOL2NOTE1 NEWSOL2NOTE SOL2NOTE2"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	customSolutions := GetOtherSolution(ExtraFilesInGOPATH, noteFiles, extraNoteFiles)
	if len(customSolutions) != solcount {
		t.Errorf("'%+v' has len '%+v'\n", customSolutions, len(customSolutions))
	}
	if strings.Join(customSolutions[runtime.GOARCH]["NEWSOL1"], " ") != sol1 {
		t.Error(customSolutions)
	}
	if strings.Join(customSolutions[runtime.GOARCH]["NEWSOL2"], " ") != sol2 {
		t.Error(customSolutions)
	}

	customSolMissing := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sol/extra-missing") + "/"
	customSolutions = GetOtherSolution(customSolMissing, noteFiles, extraNoteFiles)
	if len(customSolutions) != 0 {
		t.Errorf("'%+v' has len '%+v'\n", customSolutions, len(customSolutions))
	}

	sols := GetOtherSolution("/saptune_file_not_avail", noteFiles, extraNoteFiles)
	if len(sols) != 0 {
		t.Error(sols)
	}
}

func TestGetDeprecatedSolution(t *testing.T) {
	deprec := "deprecated"
	solcount := 2
	if system.IsPagecacheAvailable() {
		solcount = 4
	}

	solutions := GetOtherSolution(DeprecFilesInGOPATH, "", "")
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
