package main

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/actions"
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

func TestCtrlClArgs(t *testing.T) {
	orgArgs := os.Args
	buffer := bytes.Buffer{}
	// setup ErrorExit handling
	oldOSExit := system.OSExit
	defer func() { system.OSExit = oldOSExit }()
	system.OSExit = tstosExit
	oldErrorExitOut := system.ErrorExitOut
	defer func() { system.ErrorExitOut = oldErrorExitOut }()
	system.ErrorExitOut = tstErrorExitOut

	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer

	SaptuneVersion = "5"
	errVersMatchText := actions.TcmdLineSyntax("16")
	os.Args = []string{"saptune", "version"}
	system.RereadArgs()
	ctrlClArgs(&buffer)
	txt := buffer.String()
	if txt != "" {
		// ANGI TODO
		t.Log(errVersMatchText)
		t.Errorf("wrong saptune version, expected '', but got '%s'", txt)
	}
	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	buffer.Reset()
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	// because of the test situation ErrorExit does not stop the execution
	// of the function, so we are running into the PrintHelpAndExit with
	// return 1.
	// ANGI TODO - adapt ErrorExit for testing
	errVersFlagMatchText := actions.TcmdLineSyntax("15")
	os.Args = []string{"saptune", "--version"}
	system.RereadArgs()
	ctrlClArgs(&buffer)
	txt = buffer.String()
	//if txt != "" {
	checkOut(t, txt, errVersFlagMatchText)
	//if tstRetErrorExit != 0 {
	//t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	//}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	SaptuneVersion = ""
	buffer.Reset()
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	os.Args = []string{"saptune", "help"}
	// errExitMatchText := printHelpAndExitMatchText + `No notes or solutions enabled, nothing to verify.
	errExitMatchText := actions.TcmdLineSyntax("15")
	var errExitMatchText2 = errExitMatchText + errExitMatchText

	system.RereadArgs()

	ctrlClArgs(&buffer)
	txt = buffer.String()
	checkOut(t, txt, errExitMatchText)

	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	buffer.Reset()
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	os.Args = []string{"saptune", "--help"}
	system.RereadArgs()
	ctrlClArgs(&buffer)
	txt = buffer.String()
	checkOut(t, txt, errExitMatchText2)

	//if tstRetErrorExit != 0 {
	//t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	//}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	buffer.Reset()
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	os.Args = []string{"saptune", ""}
	system.RereadArgs()
	ctrlClArgs(&buffer)
	txt = buffer.String()
	checkOut(t, txt, errExitMatchText)

	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	buffer.Reset()
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	os.Args = []string{"saptune", "lock", "remove"}
	system.RereadArgs()
	ctrlClArgs(&buffer)
	txt = buffer.String()
	checkOut(t, txt, "")

	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	buffer.Reset()
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	os.Args = []string{"saptune", "lock", "list"}
	system.RereadArgs()
	ctrlClArgs(&buffer)
	txt = buffer.String()
	checkOut(t, txt, errExitMatchText)

	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	// cleanup
	os.Args = orgArgs
	system.RereadArgs()
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

	//t.Log("1. run 'saptune check'")
	callSaptuneCheckScript("check")
	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	errExitbuffer.Reset()
	tstRetErrorExit = -1
	orgArgs := os.Args
	os.Args = []string{"saptune", "check", "--force-color"}
	system.RereadArgs()

	//t.Log("2. run 'saptune check --force-color'")
	callSaptuneCheckScript("check")
	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}

	errExitbuffer.Reset()
	tstRetErrorExit = -1
	os.Args = []string{"saptune", "--format", "json", "check"}
	system.RereadArgs()

	//t.Log("3. run 'saptune --format json check' - will fail")
	callSaptuneCheckScript("check")
	if tstRetErrorExit != 130 {
		t.Errorf("error exit should be '130' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != "" {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
	}
	// cleanup
	os.Args = orgArgs
	system.RereadArgs()
}

func TestCheckForTuned(t *testing.T) {
	checkForTuned()
}

func TestCheckUpdateLeftOvers(t *testing.T) {
	checkUpdateLeftOvers()
	orgArgs := os.Args
	os.Args = []string{"saptune", "configure", "reset"}
	system.RereadArgs()
	checkUpdateLeftOvers()
	// cleanup
	os.Args = orgArgs
	system.RereadArgs()
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

	errExitbuffer := bytes.Buffer{}
	tstwriter = &errExitbuffer

	// check missing variable
	saptuneConf = fmt.Sprintf("%s/saptune_MissingVar", TstFilesInGOPATH)
	matchTxt := fmt.Sprintf("ERROR: File '%s' is broken. Missing variables 'COLOR_SCHEME'\n", saptuneConf)
	_ = checkSaptuneConfigFile(saptuneConf)

	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut := errExitbuffer.String()
	if errExOut != matchTxt {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of '%v'\n", errExOut, matchTxt)
	}

	// initialise next test
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	// check wrong STAGING value
	saptuneConf = fmt.Sprintf("%s/saptune_WrongStaging", TstFilesInGOPATH)
	saptuneVers = ""
	matchTxt = fmt.Sprintf("ERROR: Variable 'STAGING' from file '%s' contains a wrong value 'hugo'. Needs to be 'true' or 'false'\n", saptuneConf)
	_ = checkSaptuneConfigFile(saptuneConf)

	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != matchTxt {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of '%v'\n", errExOut, matchTxt)
	}

	// initialise next test
	errExitbuffer.Reset()
	tstRetErrorExit = -1

	// check logSwitchFromConfig failed
	saptuneConf = fmt.Sprintf("%s/saptune_not_avail", TstFilesInGOPATH)
	saptuneVers = ""
	err := fmt.Errorf("open %s: no such file or directory", saptuneConf)
	matchTxt = fmt.Sprintf("ERROR: Checking saptune configuration file - Unable to read file '%s': %v\n", saptuneConf, err)
	logSwitchFromConfig(saptuneConf, lSwitch)

	if tstRetErrorExit != 128 {
		t.Errorf("error exit should be '128' and NOT '%v'\n", tstRetErrorExit)
	}
	errExOut = errExitbuffer.String()
	if errExOut != matchTxt {
		t.Errorf("wrong text returned by ErrorExit: '%v' instead of '%v'\n", errExOut, matchTxt)
	}
	orgArgs := os.Args
	os.Args = []string{"saptune", "configure", "reset"}
	system.RereadArgs()
	saptuneVers = checkSaptuneConfigFile("angi")
	if saptuneVers != "3" {
		t.Errorf("wrong value for 'SAPTUNE_VERSION' - '%+v' instead of '3'\n", saptuneVers)
	}
	// cleanup
	os.Args = orgArgs
	system.RereadArgs()
}

func TestCheckSaptuneServiceDropIn(t *testing.T) {
	checkSaptuneServiceDropIn()
	system.TCSP = "azure"
	// test create drop in
	checkSaptuneServiceDropIn()
	// test drop in already available
	checkSaptuneServiceDropIn()
	// cleanup
	os.Remove("/etc/systemd/system/saptune.service.d/10-after_cloud-init.conf")
	os.Remove("/etc/systemd/system/saptune.service.d")
	system.TCSP = "skip"
}
