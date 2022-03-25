package system

import (
	"fmt"
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

// OSExit defines, which exit function should be used
var OSExit = os.Exit

// ErrorExitOut defines, which exit output function should be used
var ErrorExitOut = ErrorLog

// InfoOut defines, which log output function should be used
var InfoOut = InfoLog

// get saptune arguments and flags
var saptArgs, saptFlags = ParseCliArgs()

// DmiID is the path to the dmidecode representation in the /sys filesystem
var DmiID = "/sys/class/dmi/id"

// IsUserRoot return true only if the current user is root.
func IsUserRoot() bool {
	return os.Getuid() == 0
}

// RereadArgs parses the cli parameter again
func RereadArgs() {
	saptArgs, saptFlags = ParseCliArgs()
	return
}

// CliArg returns the i-th command line parameter,
// or empty string if it is not specified.
func CliArg(i int) string {
	if len(saptArgs) >= i+1 {
		return saptArgs[i]
	}
	return ""
}

// CliArgs returns all remaining command line parameters starting with i,
// or empty string if it is not specified.
func CliArgs(i int) []string {
	if len(saptArgs) >= i+1 {
		return saptArgs[i:]
	}
	return []string{}
}

// IsFlagSet returns true, if the flag is available on the command line
// or false, if not
func IsFlagSet(flag string) bool {
	if saptFlags[flag] == "true" {
		return true
	}
	return false
}

// GetFlagVal returns the value of a saptune commandline flag
func GetFlagVal(flag string) string {
	return saptFlags[flag]
}

// ParseCliArgs parses the command line to identify special flags and the
// 'normal' arguments
// returns a map of Flags (set/not set or value) and a slice containing the
// remaining arguments
// possible Flags - force, dryrun, help, version, output, colorscheme
// on command line - --force, --dry-run or --dryrun, --help, --version, --color-scheme, --out or --output
// Some Flags (like 'output') can have a value (--out=json or --output=csv)
func ParseCliArgs() ([]string, map[string]string) {
	stArgs := []string{os.Args[0]}
	// supported flags
	stFlags := map[string]string{"force": "false", "dryrun": "false", "help": "false", "version": "false", "show-non-compliant": "false", "output": "screen", "colorscheme": "", "notSupported": ""}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// argument is a flag
			handleFlags(arg, stFlags)
			continue
		}
		// other args
		stArgs = append(stArgs, arg)
	}
	return stArgs, stFlags
}

// handleFlags checks for valid flags in the CLI arg list
func handleFlags(arg string, flags map[string]string) {
	var valueFlag = regexp.MustCompile(`(-[\w-]+)=(.*)`)
	matches := valueFlag.FindStringSubmatch(arg)
	if len(matches) == 3 {
		// flag with value
		handleValueFlags(arg, matches, flags)
		return
	}
	handleSimpleFlags(arg, flags)
	return
}

// handleValueFlags checks for valid flags with value in the CLI arg list
func handleValueFlags(arg string, matches []string, flags map[string]string) {
	if strings.Contains(arg, "-out") {
		// --out=screen or --output=json
		matches[1] = "--output"
		flags["output"] = matches[2]
	}
	if strings.Contains(arg, "-colorscheme") {
		// --colorscheme=zebra
		flags["colorscheme"] = matches[2]
	}
	if _, ok := flags[strings.TrimLeft(matches[1], "-")]; !ok {
		setUnsupportedFlag(matches[1], flags)
	}
	return
}

// handleSimpleFlags checks for valid flags in the CLI arg list
func handleSimpleFlags(arg string, flags map[string]string) {
	// simple flags
	switch arg {
	case "--force", "-force":
		flags["force"] = "true"
	case "--dry-run", "-dry-run", "--dryrun", "-dryrun":
		flags["dryrun"] = "true"
	case "--help", "-help", "-h":
		flags["help"] = "true"
	case "--version", "-version":
		flags["version"] = "true"
	case "--show-non-compliant", "-show-non-compliant":
		flags["show-non-compliant"] = "true"
	default:
		setUnsupportedFlag(arg, flags)
	}
	return
}

// setUnsupportedFlag sets or appends a value to the unsupported Flag
// collection of unsupported Flags
func setUnsupportedFlag(val string, flags map[string]string) {
	if flags["notSupported"] != "" {
		flags["notSupported"] = flags["notSupported"] + val
	} else {
		flags["notSupported"] = flags["notSupported"] + " " + val
	}
	return
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
			// color, bold, text/template, reset bold, reset color
			stuff = stuff[1:]
			fmt.Printf("%s%sERROR: "+template+"%s%s\n", stuff...)
			if len(stuff) >= 4 {
				stuff = stuff[2 : len(stuff)-2]
			}
			ErrLog(template+"\n", stuff...)
		} else {
			ErrorExitOut(template+"\n", stuff...)
		}
	}
	if isOwnLock() {
		ReleaseSaptuneLock()
	}
	InfoOut("saptune terminated with exit code '%v'", exState)
	OSExit(exState)
}

// OutIsTerm returns true, if Stdout is a terminal
func OutIsTerm(writer *os.File) bool {
	fileInfo, _ := writer.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

// WrapTxt implements something like 'fold' command
// A given text string will be wrapped at word borders into
// lines of a given width
func WrapTxt(text string, width int) (folded []string) {
	words := strings.Split(text, " ")
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
			foldedTxt += "\n" + word
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
	fileName := ""
	ret := ""

	switch info {
	case "vendor":
		fileName = fmt.Sprintf("%s/board_vendor", DmiID)
	case "model":
		fileName = fmt.Sprintf("%s/product_name", DmiID)
	}
	if content, err = ioutil.ReadFile(fileName); err == nil {
		ret = strings.TrimSpace(string(content))
	} else {
		InfoLog("failed to read %s - %v", fileName, err)
	}
	return ret, err
}

// StripComment will strip everything right from the given comment character
// (including the comment character) and returns the resulting string
// comment characters can be '#' or ';' or something else
func StripComment(str, commentChars string) string {
	if cut := strings.IndexAny(str, commentChars); cut >= 0 {
		return strings.TrimRightFunc(str[:cut], unicode.IsSpace)
	}
	return str
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
	watch := fmt.Sprintf("%s", t.Format("2006/01/02 15:04:05.99999999"))
	return watch
}
