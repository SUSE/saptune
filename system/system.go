package system

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode"
)

// SaptuneSectionDir defines saptunes saved state directory
const SaptuneSectionDir = "/run/saptune/sections"

// map to hold the current available systemd services
var services map[string]string

// stdOutOrg contains the origin stdout for resetting, if needed
var stdOutOrg = os.Stdout

// OSExit defines, which exit function should be used
var OSExit = os.Exit

// ErrorExitOut defines, which exit output function should be used
var ErrorExitOut = ErrorLog

// ErrExitOut defines the output function, which should be used in case
// of colored output
var ErrExitOut = errExitOut

// InfoOut defines, which log output function should be used
var InfoOut = InfoLog

// DmiID is the path to the dmidecode representation in the /sys filesystem
var DmiID = "/sys/class/dmi/id"

// IsUserRoot return true only if the current user is root.
func IsUserRoot() bool {
	return os.Getuid() == 0
}

// GetSolutionSelector returns the architecture string
// needed to select the supported set os solutions
func GetSolutionSelector() string {
	solutionSelector := runtime.GOARCH
	if IsPagecacheAvailable() {
		solutionSelector = solutionSelector + "_PC"
	}
	return solutionSelector
}

// CmdIsAvailable returns true, if the cmd is available.
func CmdIsAvailable(cmdName string) bool {
	if _, err := os.Stat(cmdName); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetOsVers returns the OS version
func GetOsVers() string {
	// VERSION="12", VERSION="15"
	// VERSION="12-SP1", VERSION="12-SP2", VERSION="12-SP3"
	var re = regexp.MustCompile(`VERSION="([\w-]+)"`)
	val, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(string(val))
	if len(matches) == 0 {
		return ""
	}
	return matches[1]
}

// GetOsName returns the OS name
func GetOsName() string {
	// NAME="SLES"
	var re = regexp.MustCompile(`NAME="([\w\s]+)"`)
	val, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(string(val))
	if len(matches) == 0 {
		return ""
	}
	return matches[1]
}

// IsSLE15 returns true, if System is running a SLE15 release
func IsSLE15() bool {
	var re = regexp.MustCompile(`15-SP\d+`)
	if GetOsName() == "SLES" && (GetOsVers() == "15" || re.MatchString(GetOsVers())) {
		return true
	}
	return false
}

// IsSLE12 returns true, if System is running a SLE12 release
func IsSLE12() bool {
	var re = regexp.MustCompile(`12-SP\d+`)
	if GetOsName() == "SLES" && (GetOsVers() == "12" || re.MatchString(GetOsVers())) {
		return true
	}
	return false
}

// CheckForPattern returns true, if the file is available and
// contains the expected string
func CheckForPattern(file, pattern string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	//check whether content contains substring pattern
	return strings.Contains(string(content), pattern)
}

// CalledFrom returns the name and the line number of the calling source file
func CalledFrom() string {
	ret := ""
	_, file, no, ok := runtime.Caller(2)
	if ok {
		_, relfile := filepath.Split(file)
		ret = fmt.Sprintf("%s:%d: ", relfile, no)
	}
	return ret
}

func errExitOut(writer io.Writer, template string, stuff ...interface{}) {
	// stuff is: color, bold, text/template, reset bold, reset color
	stuff = stuff[1:]
	fmt.Fprintf(writer, "%s%sERROR: "+template+"%s%s\n", stuff...)
	if len(stuff) >= 4 {
		stuff = stuff[2 : len(stuff)-2]
	}
	ErrorLog(template+"\n", stuff...)
}

// ErrorExit prints the message to stderr and exit 1.
func ErrorExit(template string, stuff ...interface{}) {
	exState := 1
	fieldType := ""
	field := len(stuff) - 1
	if field >= 0 {
		fieldType = reflect.TypeOf(stuff[field]).String()
	}
	if fieldType == "*exec.ExitError" {
		// get return code of failed command, if available
		if exitError, ok := stuff[field].(*exec.ExitError); ok {
			exState = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		}
	}
	if fieldType == "int" {
		exState = reflect.ValueOf(stuff[field]).Interface().(int)
		stuff = stuff[:len(stuff)-1]
	}
	if len(template) != 0 {
		if len(stuff) > 0 && stuff[0] == "colorPrint" {
			ErrExitOut(os.Stderr, template, stuff...)
		} else {
			ErrorExitOut(template+"\n", stuff...)
		}
	}
	if isOwnLock() {
		ReleaseSaptuneLock()
	}
	if jerr := jOut(exState); jerr != nil {
		exState = 130
	}
	InfoOut("saptune terminated with exit code '%v'", exState)
	OSExit(exState)
}

// OutIsTerm returns true, if Stdout is a terminal
func OutIsTerm(writer *os.File) bool {
	fileInfo, _ := writer.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// InitOut initializes the various output methodes
// currently only json and screen are supported
func InitOut(logSwitch map[string]string) {
	if GetFlagVal("format") == "json" {
		// if writing json format, switch off
		// the stdout and stderr output of the log messages
		logSwitch["verbose"] = "off"
		logSwitch["error"] = "off"
		// switch off stdout
		if os.Getenv("SAPTUNE_JDEBUG") != "on" {
			os.Stdout, _ = os.Open(os.DevNull)
		}
		jInit()
	}
}

// SwitchOffOut disables stdout and stderr
func SwitchOffOut() (*os.File, *os.File) {
	oldStdout := os.Stdout
	oldSdterr := os.Stderr
	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)
	return oldStdout, oldSdterr
}

// SwitchOnOut restores stdout and stderr to the settings before SwitchOffOut
// was called
func SwitchOnOut(stdout *os.File, stderr *os.File) {
	os.Stdout = stdout
	os.Stderr = stderr
}

// WrapTxt implements something like 'fold' command
// A given text string will be wrapped at word borders into
// lines of a given width
func WrapTxt(text string, width int) (folded []string) {
	var words []string
	fallback := false

	if strings.Contains(text, " ") {
		words = strings.Split(text, " ")
	} else {
		// fallback (e.g. net.ipv4.ip_local_reserved_ports)
		words = strings.Split(text, ",")
		fallback = true
	}
	if len(words) == 0 {
		return
	}
	foldedTxt := words[0]
	spaceLeft := width - len(foldedTxt)
	noSpace := false
	for _, word := range words[1:] {
		if word == "\n" {
			foldedTxt += word
			spaceLeft = width
			noSpace = true
			continue
		}
		if len(word)+1 > spaceLeft {
			// fold; start next row
			if fallback {
				foldedTxt += ",\n" + word
			} else {
				foldedTxt += "\n" + word
			}
			if strings.HasSuffix(word, "\n") {
				spaceLeft = width
				noSpace = true
			} else {
				spaceLeft = width - len(word)
				noSpace = false
			}
		} else {
			if noSpace {
				foldedTxt += word
				spaceLeft -= len(word)
				noSpace = false
			} else if fallback {
				foldedTxt += "," + word
				spaceLeft -= 1 + len(word)
			} else {
				foldedTxt += " " + word
				spaceLeft -= 1 + len(word)
			}
			if strings.HasSuffix(word, "\n") {
				spaceLeft = width
				noSpace = true
			}
		}
	}
	folded = strings.Split(foldedTxt, "\n")
	return
}

// GetDmiID return the content of /sys/devices/virtual/dmi/id/<file> or
// an empty string
func GetDmiID(file string) (string, error) {
	var err error
	var content []byte
	ret := ""
	fileName := fmt.Sprintf("%s/%s", DmiID, file)
	if content, err = ioutil.ReadFile(fileName); err == nil {
		ret = strings.TrimSpace(string(content))
	} else {
		InfoLog("failed to read %s - %v", fileName, err)
	}
	return ret, err
}

// GetHWIdentity returns the hardwar vendor or model of the system
// needs adaption, if the files to identify the hardware will change or
// if we need to look at different files for different vendors
// but the 'open' API GetDmiID exists for workarounds at customer side
func GetHWIdentity(info string) (string, error) {
	var err error
	var content []byte
	fix := ""
	fileName := ""
	ret := ""

	switch info {
	case "vendor":
		if runtime.GOARCH == "ppc64le" {
			fix = "IBM"
		} else {
			fileName = fmt.Sprintf("%s/board_vendor", DmiID)
		}
	case "model":
		if runtime.GOARCH == "ppc64le" {
			fileName = "/sys/firmware/devicetree/base/model"
		} else {
			fileName = fmt.Sprintf("%s/product_name", DmiID)
		}
	}
	if fileName != "" {
		if content, err = ioutil.ReadFile(fileName); err == nil {
			ret = strings.TrimSpace(string(content))
		} else {
			InfoLog("failed to read %s - %v", fileName, err)
		}
	}
	if fix != "" {
		ret = fix
		err = nil
	}
	return ret, err
}

// StripComment will strip everything right from the given comment character
// (including the comment character) and returns the resulting string
// comment characters can be '#' or ';' or something else
// or a regex like `\s#[^#]|"\s#[^#]`
func StripComment(str, commentChars string) string {
	ret := str
	re := regexp.MustCompile(commentChars)
	if cut := re.FindStringIndex(str); cut != nil {
		ret = strings.TrimRightFunc(str[:cut[0]], unicode.IsSpace)
		// strip masked # (\s## -> \s#) inside the text
		re = regexp.MustCompile(`\s(##)`)
		ret = re.ReplaceAllString(ret, "#")
	}
	return ret
}

// GetVirtStatus gets the status of virtualization environment
func GetVirtStatus() string {
	vtype := ""
	// first check vm (-v)
	virt, vm, _ := SystemdDetectVirt("-v")
	if virt {
		// vm detected
		vtype = vm
	}
	// next check container (-c)
	virt, container, _ := SystemdDetectVirt("-c")
	if virt {
		// container detected
		if vtype == "" {
			vtype = container
		} else {
			vtype = vtype + " " + container
		}
	}
	// last check for chroot (-r)
	// be in mind, that the command will not deliver any output, but only
	// return 0, if it found a chroot env or 1, if not
	virt, _, _ = SystemdDetectVirt("-r")
	if virt {
		// chroot detected
		if vtype == "" {
			vtype = "chroot"
		} else {
			vtype = vtype + " chroot"
		}
	}
	if vtype == "" {
		vtype = "none"
	}
	return vtype
}

// Watch prints the current time
func Watch() string {
	t := time.Now()
	//watch := fmt.Sprintf("%s", t.Format(time.UnixDate))
	watch := t.Format("2006/01/02 15:04:05.99999999")
	return watch
}
