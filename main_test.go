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

func TestCheckWorkingArea(t *testing.T) {
	os.Remove("/usr/share/saptune")
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune")
	target := "/usr/share/saptune"
	if err := os.Symlink(src, target); err != nil {
		t.Errorf("linking '%s' to '%s' failed - '%v'", src, target, err)
	}
	defer os.Remove("/usr/share/saptune")

	// setup ErrorExit handling
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer

	checkWorkingArea()
	if tstRetErrorExit != -1 {
		t.Errorf("error exit should be '-1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	// cleanup - remove link and create directory
	os.Remove("/usr/share/saptune")
	os.MkdirAll("/usr/share/saptune", 0755)
	os.RemoveAll("/var/lib/saptune/working/notes")
}

func TestCallSaptuneCheckScript(t *testing.T) {
	// setup ErrorExit handling
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer
	txt2chk := `ERROR: command '/usr/sbin/saptune_check' failed with error 'fork/exec /usr/sbin/saptune_check: no such file or directory'

`

	callSaptuneCheckScript("check")
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != txt2chk {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of '%v'\n", errExOut, txt2chk)
	}

	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/saptune_check")
	dest := "/usr/sbin/saptune_check"
	err := system.CopyFile(src, dest)
	if err != nil {
		t.Error(err)
	}
	err = os.Chmod(dest, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(dest)
	errExitbuffer.Reset()

	callSaptuneCheckScript("check")
	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}
}

func TestCheckForTuned(t *testing.T) {
	checkForTuned()
}

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
	//buffer := bytes.Buffer{}

	lSwitch := logSwitch
	saptuneVers := checkSaptuneConfigFile(saptuneConf)
	if saptuneVers != "5" {
		t.Errorf("wrong value for 'SAPTUNE_VERSION' - '%+v' instead of ''\n", saptuneVers)
	}
	logSwitchFromConfig(saptuneConf, lSwitch)
	if lSwitch["debug"] != "on" {
		t.Errorf("wrong value for 'DEBUG' - '%+v' instead of 'on'\n", lSwitch["debug"])
	}
	if lSwitch["verbose"] != "on" {
		t.Errorf("wrong value for 'VERBOSE' - '%+v' instead of 'on'\n", lSwitch["debug"])
	}

	//buffer.Reset()

	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer

	// check missing variable
	saptuneConf = fmt.Sprintf("%s/saptune_MissingVar", TstFilesInGOPATH)
	matchTxt := fmt.Sprintf("ERROR: File '%s' is broken. Missing variables 'COLOR_SCHEME'\n", saptuneConf)
	//lSwitch = logSwitch
	_ = checkSaptuneConfigFile(saptuneConf)

	//txt := buffer.String()
	//checkOut(t, txt, matchTxt)
	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != matchTxt {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of '%v'\n", errExOut, matchTxt)
	}

	// initialise next test
	//buffer.Reset()
	errExitbuffer.Reset()

	// check wrong STAGING value
	saptuneConf = fmt.Sprintf("%s/saptune_WrongStaging", TstFilesInGOPATH)
	saptuneVers = ""
	matchTxt = fmt.Sprintf("ERROR: Variable 'STAGING' from file '%s' contains a wrong value 'hugo'. Needs to be 'true' or 'false'\n", saptuneConf)
	//lSwitch = logSwitch
	_ = checkSaptuneConfigFile(saptuneConf)

	//txt = buffer.String()
	//checkOut(t, txt, matchTxt)
	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != matchTxt {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of '%v'\n", errExOut, matchTxt)
	}
}
