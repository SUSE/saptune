package system

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
)

var tstRetErrorExit = -1
var tstosExit = func(val int) {
	tstRetErrorExit = val
}
var tstwriter io.Writer
var errwriter io.Writer
var tstErrorExitOut = func(str string, out ...interface{}) error {
	fmt.Fprintf(tstwriter, "ERROR: "+str, out...)
	return fmt.Errorf(str+"\n", out...)
}
var tstErrExitOut = func(errw io.Writer, str string, out ...interface{}) {
	out = out[1:]
	fmt.Printf("%v\n", errw)
	fmt.Fprintf(errwriter, "%s%sERROR: "+str+"%s%s\n", out...)
	if len(out) >= 4 {
		out = out[2 : len(out)-2]
	}
	fmt.Fprintf(tstwriter, "ERROR: "+str, out...)
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

func TestIsUserRoot(t *testing.T) {
	if !IsUserRoot() {
		t.Log("the test requires root access")
	}
}

func TestGetSolutionSelector(t *testing.T) {
	solSelector := GetSolutionSelector()
	t.Logf("architecture is '%s'\n", solSelector)
	//if solSelector != "amd64" && solSelector != "amd64_PC" && solSelector != "ppc64le" && solSelector != "ppc64le_PC" && solSelector != "TRAVIS_TODO" {
	if solSelector != "amd64" && solSelector != "amd64_PC" && solSelector != "ppc64le" && solSelector != "ppc64le_PC" {
		t.Errorf("Test failed, solSelector '%s'", solSelector)
	}
}

func TestGetOsName(t *testing.T) {
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
	actualVal := GetOsName()
	//if actualVal != "SLES" && actualVal != "openSUSE Leap" {
	if actualVal != "SLES" {
		t.Logf("OS is '%s' and not 'SLES'\n", actualVal)
		_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr15"), "/etc/os-release")
		actualVal = GetOsName()
		if actualVal != "SLES" {
			t.Errorf("OS is '%s' and not 'SLES'\n", actualVal)
		}
	}
	// test with non existing file
	os.Remove("/etc/os-release")
	actualVal = GetOsName()
	if actualVal != "" {
		t.Errorf("/etc/os-release should not exist, but returns value '%v'\n", actualVal)
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestGetOsVers(t *testing.T) {
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
	actualVal := GetOsVers()
	if actualVal == "" {
		_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr15"), "/etc/os-release")
		actualVal = GetOsVers()
		if actualVal != "15-SP2" {
			t.Errorf("unexpected OS version '%s'\n", actualVal)
		}
		_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr12"), "/etc/os-release")
		actualVal = GetOsVers()
		if actualVal != "12-SP5" {
			t.Errorf("unexpected OS version '%s'\n", actualVal)
		}
	} else {
		switch actualVal {
		case "12", "12-SP1", "12-SP2", "12-SP3", "12-SP4", "12-SP5", "15", "15-SP1", "15-SP2", "15-SP3":
			t.Logf("expected OS version '%s' found\n", actualVal)
		default:
			t.Logf("unexpected OS version '%s'\n", actualVal)
		}
	}

	// test with non existing file
	os.Remove("/etc/os-release")
	actualVal = GetOsVers()
	if actualVal != "" {
		t.Errorf("/etc/os-release should not exist, but returns value '%v'\n", actualVal)
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestIsSLE15(t *testing.T) {
	_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr15"), "/etc/os-release")
	if IsSLE15() {
		t.Logf("found SLE15 OS version\n")
		_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
		if IsSLE15() {
			t.Errorf("expected a non SLE15 os version, but OS version is '%s'\n", GetOsVers())
		}
	} else {
		t.Errorf("expected '15-SP2', but OS version is '%s'\n", GetOsVers())
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestIsSLE12(t *testing.T) {
	_ = CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr12"), "/etc/os-release")
	if IsSLE12() {
		t.Logf("found SLE12 OS version\n")
		_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
		if IsSLE12() {
			t.Errorf("expected a non SLE12 os version, but OS version is '%s'\n", GetOsVers())
		}
	} else {
		t.Errorf("expected '12-SP5', but OS version is '%s'\n", GetOsVers())
	}
	_ = CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestCmdIsAvailable(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/go") {
		t.Error("'/usr/bin/go' not found")
	}
	if CmdIsAvailable("/cmd_not_avail.CnA") {
		t.Error("found '/cmd_not_avail.CnA'")
	}
}

func TestCheckForPattern(t *testing.T) {
	if CheckForPattern("/file_not_available", "Test_Text") {
		t.Error("found '/file_not_available'")
	}
	if CheckForPattern("/sys/module/video/uevent", "Test_Text") {
		t.Error("could read '/sys/module/video/uevent'")
	}
}

func TestCalledFrom(t *testing.T) {
	val := CalledFrom()
	if !strings.Contains(val, "testing.go") {
		t.Fatalf("called from '%s' instead of 'testing.go'\n", val)
	}
}

func TestErrorExit(t *testing.T) {
	var setRedText = "\033[31m"
	var setBoldText = "\033[1m"
	var resetBoldText = "\033[22m"
	var resetTextColor = "\033[0m"

	oldOSExit := OSExit
	defer func() { OSExit = oldOSExit }()
	OSExit = tstosExit
	oldErrorExitOut := ErrorExitOut
	defer func() { ErrorExitOut = oldErrorExitOut }()
	ErrorExitOut = tstErrorExitOut
	oldErrExitOut := ErrExitOut
	defer func() { ErrExitOut = oldErrExitOut }()
	ErrExitOut = tstErrExitOut
	buffer := bytes.Buffer{}
	tstwriter = &buffer

	ErrorExit("Hallo")
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}
	txt := buffer.String()
	checkOut(t, txt, "ERROR: Hallo\n")

	buffer.Reset()
	errbuf := bytes.Buffer{}
	errwriter = &errbuf
	ErrorExit("Colored Hallo", "colorPrint", setRedText, setBoldText, resetBoldText, resetTextColor)
	txt = buffer.String()
	checkOut(t, txt, "ERROR: Colored Hallo")
	errtxt := errbuf.String()
	//lint:ignore ST1018 Unicode control characters are expected here
	checkOut(t, errtxt, "[31m[1mERROR: Colored Hallo[22m[0m\n")

	// check errExitOut function
	outbuf := bytes.Buffer{}
	errExitOut(&outbuf, "Colored Hallo direct", "colorPrint", setRedText, setBoldText, resetBoldText, resetTextColor)
	txt = outbuf.String()
	//lint:ignore ST1018 Unicode control characters are expected here
	checkOut(t, txt, "[31m[1mERROR: Colored Hallo direct[22m[0m\n")

	SaptuneLock()
	// to reach ErrorExit("saptune currently in use, try later ...", 11)
	SaptuneLock()
	ErrorExit("", 0)
	if tstRetErrorExit != 0 {
		t.Errorf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
	}
	// error is '*exec.ExitError'
	cmd := exec.Command("/usr/bin/false")
	err := cmd.Run()
	t.Logf("%s: command failed with error '%v'\n", Watch(), err)
	if err != nil {
		ErrorExit("command failed with error '%v'\n", err)
	}
	if tstRetErrorExit != 1 {
		t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
	}

	ErrorExit("", 5)
	if tstRetErrorExit != 5 {
		t.Errorf("error exit should be '5' and NOT '%v'\n", tstRetErrorExit)
	}
	// error is '*os.PathError'
	_, err = os.Stat("/not_avail")
	if err != nil {
		ErrorExit("problems with file '/not_avail': %v", err)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
	}
}

func TestOutIsTerm(t *testing.T) {
	pipeName := "/tmp/saptune_pipe_tst"
	syscall.Mkfifo(pipeName, 0666)
	pipeFile, _ := os.OpenFile(pipeName, os.O_CREATE|syscall.O_NONBLOCK, os.ModeNamedPipe)
	pipeInfo, _ := pipeFile.Stat()
	if OutIsTerm(pipeFile) {
		t.Errorf("file is a pipe, but reported as terminal - %+v\n", pipeInfo.Mode())
	}
	pipeFile.Close()
	os.Remove(pipeName)
	termFile := os.Stdin
	termInfo, _ := termFile.Stat()
	if !OutIsTerm(termFile) {
		t.Errorf("file is a terminal, but reported as NOT a terminal - %+v\n", termInfo.Mode())
	}
}

func TestWrapTxt(t *testing.T) {
	testString := "This is a really long text, which does not make any sense, except that I need something for testing my new function.\n need some line feeds\n and a second one\n 12345 \n 678910\n"
	expected := []string{"This is a really", "long text, which", "does not make any", "sense, except", "that I need", "something for", "testing my new", "function.", "need some line", "feeds", "and a second one", "12345", "678910", ""}
	actual := WrapTxt(testString, 17)
	if len(actual) != len(expected) {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	} else {
		for i, line := range actual {
			expectedLine := expected[i]
			if line != expectedLine {
				t.Errorf("Test failed, expected: '%s', got: '%s'", expectedLine, line)
			}
		}
	}

	testString = "ONLY_ON_WORD"
	expected = []string{"ONLY_ON_WORD"}
	actual = WrapTxt(testString, 17)
	if len(actual) != len(expected) {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	} else {
		for i, line := range actual {
			expectedLine := expected[i]
			if line != expectedLine {
				t.Errorf("Test failed, expected: '%s', got: '%s'", expectedLine, line)
			}
		}
	}

	testString = " "
	expected = []string{" "}
	actual = WrapTxt(testString, 17)
	if len(actual) != len(expected) {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	} else {
		for i, line := range actual {
			expectedLine := expected[i]
			if line != expectedLine {
				t.Errorf("Test failed, expected: '%s', got: '%s'", expectedLine, line)
			}
		}
	}

	testString = ""
	expected = []string{""}
	actual = WrapTxt(testString, 17)
	if len(actual) != len(expected) {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	} else {
		for i, line := range actual {
			expectedLine := expected[i]
			if line != expectedLine {
				t.Errorf("Test failed, expected: '%s', got: '%s'", expectedLine, line)
			}
		}
	}

	testString = "\n"
	expected = []string{"", ""}
	actual = WrapTxt(testString, 17)
	if len(actual) != len(expected) {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	} else {
		for i, line := range actual {
			expectedLine := expected[i]
			if line != expectedLine {
				t.Errorf("Test failed, expected: '%s', got: '%s'", expectedLine, line)
			}
		}
	}
}

func TestGetDmiID(t *testing.T) {
	DmiID = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")
	expected := "SUSE saptune"
	file := "product_name"
	dmi, _ := GetDmiID(file)
	if dmi != "SUSE saptune" {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, dmi)
	}
	file = "product_hugo"
	expected = ""
	dmi, _ = GetDmiID(file)
	if dmi != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, dmi)
	}
	file = "no_dmi_file_found"
	_, err := GetDmiID(file)
	if err == nil {
		t.Errorf("file '%s' exists, but shouldn't", file)
	}
	DmiID = "/sys/class/dmi/id"
}

func TestGetHWIdentity(t *testing.T) {
	DmiID = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")
	expected := "SUSE HW"
	info := "vendor"
	hwvend, _ := GetHWIdentity(info)
	if hwvend != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, hwvend)
	}

	info = "model"
	expected = "SUSE saptune"
	hwvend, _ = GetHWIdentity(info)
	if hwvend != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, hwvend)
	}

	info = "hugo"
	expected = ""
	hwvend, _ = GetHWIdentity(info)
	if hwvend != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, hwvend)
	}
	DmiID = "/sys/class/dmi/id"
}

func TestStripComments(t *testing.T) {
	str := "Test string with # comment to strip"
	exp := "Test string with"
	res := StripComment(str, "#")
	if res != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, res)
	}
	str = "Test string without comment to strip"
	exp = "Test string without comment to strip"
	res = StripComment(str, "#")
	if res != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, res)
	}
	str = "Test string with another; comment to strip"
	exp = "Test string with another"
	res = StripComment(str, ";")
	if res != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, res)
	}
}

func TestGetVirtStatus(t *testing.T) {
	oldSystemddvCmd := systemddvCmd
	// test: virtualization found
	systemddvCmd = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/systemdDVOK")
	exp := "kvm lxc chroot"
	vtype := GetVirtStatus()
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}

	// test: virtualization NOT available
	systemddvCmd = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/systemdDVNOK")
	exp = "none"
	vtype = GetVirtStatus()
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	systemddvCmd = oldSystemddvCmd
}
