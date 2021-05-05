package system

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// SaptuneSectionDir defines saptunes saved state directory
const SaptuneSectionDir = "/var/lib/saptune/sections"

// map to hold the current available systemd services
var services map[string]string

// OSExit defines, which exit function should be used
var OSExit = os.Exit

// ErrorExitOut defines, which exit output function should be used
var ErrorExitOut = ErrorLog

// get saptune arguments and flags
var saptArgs, saptFlags = ParseCliArgs()

// DmiID is the path to the dmidecode representation in the /sys filesystem
var DmiID = "/sys/class/dmi/id"

// IsUserRoot return true only if the current user is root.
func IsUserRoot() bool {
	return os.Getuid() == 0
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

// GetOutTarget returns the target for the saptune command output
// default is 'screen'
func GetOutTarget() string {
	return saptFlags["output"]
}

// ParseCliArgs parses the command line to identify special flags and the
// 'normal' arguments
// returns a map of Flags (set or not) and a slice containing the remaining
// arguments
// possible Flags - force, dryrun, help, version, output
// on command line - --force, --dry-run or --dryrun, --help, --version, --out or --output
// Only the Flag 'output' can have an argument (--out=json or --output=csv)
func ParseCliArgs() ([]string, map[string]string) {
	var isOutFlag = regexp.MustCompile(`-([\w-]+)=.*`)
	var isOutArg = regexp.MustCompile(`--out.*=(\w+)`)
	stArgs := []string{os.Args[0]}
	stFlags := map[string]string{"force": "false", "dryrun": "false", "help": "false", "version": "false", "output": "screen", "notSupported": ""}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// flag handling
			if isOutFlag.MatchString(arg) {
				// --out=screen // --output=json
				matches := isOutArg.FindStringSubmatch(arg)
				if len(matches) > 0 {
					stFlags["output"] = matches[1]
				}
				continue
			}
			switch arg {
			case "--force", "-force":
				stFlags["force"] = "true"
			case "--dry-run", "-dry-run", "--dryrun", "-dryrun":
				stFlags["dryrun"] = "true"
			case "--help", "-help", "-h":
				stFlags["help"] = "true"
			case "--version", "-version":
				stFlags["version"] = "true"
			default:
				stFlags["notSupported"] = "arg"
			}
			continue
		}
		// other args
		stArgs = append(stArgs, arg)
	}
	return stArgs, stFlags
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

// GetAvailServices returns a map of the available services of the system
func GetAvailServices() map[string]string {
	allServices := make(map[string]string)
	cmdArgs := []string{"--no-pager", "list-unit-files"}
	cmdOut, err := exec.Command(systemctlCmd, cmdArgs...).CombinedOutput()
	if err != nil {
		WarningLog("There was an error running external command %s %s: %v, output: %s", systemctlCmd, cmdArgs, err, cmdOut)
		return allServices
	}
	for _, line := range strings.Split(string(cmdOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		serv := strings.TrimSpace(fields[0])
		allServices[serv] = serv
	}
	return allServices
}

// GetServiceName returns the systemd service name for supported services
func GetServiceName(service string) string {
	serviceName := ""
	if services == nil || len(services) == 0 {
		services = GetAvailServices()
	}
	if _, ok := services[service]; ok {
		serviceName = service
	} else {
		serv := fmt.Sprintf("%s.service", service)
		if _, ok := services[serv]; ok {
			serviceName = serv
		}
	}
	if serviceName == "" {
		WarningLog("skipping unkown service '%s'", service)
	}
	return serviceName
}

// ReadConfigFile read content of config file
func ReadConfigFile(fileName string, autoCreate bool) ([]byte, error) {
	content, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) && autoCreate {
		content = []byte{}
		err = os.MkdirAll(path.Dir(fileName), 0755)
		if err == nil {
			err = ioutil.WriteFile(fileName, []byte{}, 0644)
		}
	}
	return content, err
}

// EditFile copies a source file to another name and opens this copy in an
// editor defined by environment variable "EDITOR" or in 'vim'
func EditFile(srcFile, destFile string) error {
	editor := os.Getenv("EDITOR")
	// copy source to destintion
	if err := CopyFile(srcFile, destFile); err != nil {
		ErrorLog("Problems while copying '%s' to '%s' - %v", srcFile, destFile, err)
		return err
	}
	if editor == "" {
		editor = "/usr/bin/vim" // launch vim by default
	}
	cmd := exec.Command(editor, destFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ErrorLog("Failed to launch editor %s: %v", editor, err)
	}
	return err
}

// ChkMD5Pair checks, if the md5sum of 2 files are equal
func ChkMD5Pair(srcFile, destFile string) bool {
	ret := false
	chkSumSrc, err := GetMD5Hash(srcFile)
	if err != nil {
		ErrorLog("Failed to get md5 checksum of file '%s': %v", srcFile, err)
	}
	chkSumDest, err := GetMD5Hash(destFile)
	if err != nil {
		ErrorLog("Failed to get md5 checksum of file '%s': %v", destFile, err)
	}
	if chkSumSrc == chkSumDest && chkSumSrc != "" {
		ret = true
	}
	return ret
}

// GetMD5Hash generate the md5sum of a file
func GetMD5Hash(file string) (string, error) {
	md5Sum := ""
	// open file for reading
	f, err := os.Open(file)
	if err != nil {
		return md5Sum, err
	}
	defer f.Close()

	// create a new hash, which is a writer interface
	hash := md5.New()

	// copy the file in the hash interface
	if _, err := io.Copy(hash, f); err != nil {
		return md5Sum, err
	}
	// hash and print as string. Pass nil since the data is not coming
	// in as a slice argument but is coming through the writer interface
	md5Sum = fmt.Sprintf("%x", hash.Sum(nil))
	return md5Sum, nil
}

// CopyFile from source to destination
func CopyFile(srcFile, destFile string) error {
	var src, dst *os.File
	var err error
	if src, err = os.Open(srcFile); err == nil {
		defer src.Close()
		if dst, err = os.OpenFile(destFile, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644); err == nil {
			defer dst.Close()
			if _, err = io.Copy(dst, src); err == nil {
				// flush file content from  memory to disk
				err = dst.Sync()
			}
		}
	}
	return err
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
		ErrorExitOut(template+"\n", stuff...)
	}
	if isOwnLock() {
		ReleaseSaptuneLock()
	}
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
	}
	return ret, err
}

// Watch prints the current time
func Watch() string {
	t := time.Now()
	//watch := fmt.Sprintf("%s", t.Format(time.UnixDate))
	watch := fmt.Sprintf("%s", t.Format("2006/01/02 15:04:05.99999999"))
	return watch
}
