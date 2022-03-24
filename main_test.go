package main

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"path"
	"testing"
)

var TstFilesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")

// setup for ErroExit catches
var tstRetErrorExit = -1
var tstosExit = func(val int) {
	tstRetErrorExit = val
}
var tstwriter io.Writer
var tstErrorExitOut = func(str string, out ...interface{}) error {
	fmt.Fprintf(tstwriter, "ERROR: "+str, out...)
	return fmt.Errorf(str+"\n", out...)
}

var checkOut = func(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		fmt.Println("==============")
		fmt.Println(got)
		fmt.Println("==============")
		fmt.Println(want)
		fmt.Println("==============")
		t.Errorf("Output differs from expected one")
	}
}

// func TestcheckForTuned()

func TestCheckUpdateLeftOvers(t *testing.T) {
	checkUpdateLeftOvers()
}

func TestCheckSaptuneConfigFile(t *testing.T) {
	// setup ErrorExit handling
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	logSwitch := map[string]string{"verbose": "", "debug": ""}
	// check saptune version and debug
	saptuneConf := fmt.Sprintf("%s/saptune_VersAndDebug", TstFilesInGOPATH)
	buffer := bytes.Buffer{}

	lSwitch := logSwitch
	saptuneVers := checkSaptuneConfigFile(&buffer, saptuneConf, lSwitch)
	if saptuneVers != "5" {
		t.Errorf("wrong value for 'SAPTUNE_VERSION' - '%+v' instead of ''\n", saptuneVers)
	}
	if lSwitch["debug"] != "1" {
		t.Errorf("wrong value for 'DEBUG' - '%+v' instead of '1'\n", lSwitch["debug"])
	}
	if lSwitch["verbose"] != "on" {
		t.Errorf("wrong value for 'VERBOSE' - '%+v' instead of 'on'\n", lSwitch["debug"])
	}

	buffer.Reset()
	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer

	// check missing variable
	saptuneConf = fmt.Sprintf("%s/saptune_NoVersion", TstFilesInGOPATH)
	matchTxt := fmt.Sprintf("Error: File '%s' is broken. Missing variables 'SAPTUNE_VERSION'\n", saptuneConf)
	lSwitch = logSwitch
	saptuneVers = checkSaptuneConfigFile(&buffer, saptuneConf, lSwitch)

	txt := buffer.String()
	checkOut(t, txt, matchTxt)
	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	// initialise next test
	buffer.Reset()
	errExitbuffer.Reset()

	// check wrong STAGING value
	saptuneConf = fmt.Sprintf("%s/saptune_WrongStaging", TstFilesInGOPATH)
	saptuneVers = ""
	matchTxt = fmt.Sprintf("Error: Variable 'STAGING' from file '%s' contains a wrong value 'hugo'. Needs to be 'true' or 'false'\n", saptuneConf)
	lSwitch = logSwitch
	saptuneVers = checkSaptuneConfigFile(&buffer, saptuneConf, lSwitch)

	txt = buffer.String()
	checkOut(t, txt, matchTxt)
	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}
}
