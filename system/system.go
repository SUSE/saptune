package system

import (
	"encoding/json"
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
	"strconv"
	"strings"
	"syscall"
)

// SaptuneSectionDir defines saptunes saved state directory
const SaptuneSectionDir = "/var/lib/saptune/sections"

// saptune lock file
var stLockFile = "/var/run/.saptune.lock"

// map to hold the current available systemd services
var services map[string]string

// OSExit defines, which exit function should be used
var OSExit = os.Exit

// ErrorExitOut defines, which exit output function should be used
var ErrorExitOut = ErrorLog

// get saptune arguments and flags
var saptArgs, saptFlags = ParseCliArgs()

// BlockDev contains all key-value pairs of current avaliable
// block devices in /sys/block
type BlockDev struct {
	AllBlockDevs    []string
	BlockAttributes map[string]map[string]string
}

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

// CopyFile from source to destination
func CopyFile(srcFile, destFile string) error {
	var src, dst *os.File
	var err error
	if src, err = os.Open(srcFile); err == nil {
		defer src.Close()
		if dst, err = os.OpenFile(destFile, os.O_RDWR|os.O_CREATE, 0644); err == nil {
			defer dst.Close()
			if _, err = io.Copy(dst, src); err == nil {
				// flush file content from  memory to disk
				err = dst.Sync()
			}
		}
	}
	return err
}

// BlockDeviceIsDisk checks, if a block device is a disk
// /sys/block/*/device/type (TYPE_DISK / 0x00)
// does not work for virtio block devices, needs workaround
func BlockDeviceIsDisk(dev string) bool {
	isVD := regexp.MustCompile(`^vd\w+$`)
	fname := fmt.Sprintf("/sys/block/%s/device/type", dev)
	dtype, err := ioutil.ReadFile(fname)
	if err != nil || strings.TrimSpace(string(dtype)) != "0" {
		if strings.Join(isVD.FindStringSubmatch(dev), "") == "" {
			// unsupported device
			return false
		}
	}
	return true
}

// GetBlockDeviceInfo reads content of stored block device information.
// content stored in SaptuneSectionDir (/var/lib/saptune/sections)
// as blockdev.run
// Return the content as BlockDev
func GetBlockDeviceInfo() (*BlockDev, error) {
	bdevFileName := fmt.Sprintf("%s/blockdev.run", SaptuneSectionDir)
	bdevConf := &BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	content, err := ioutil.ReadFile(bdevFileName)
	if err == nil && len(content) != 0 {
		err = json.Unmarshal(content, &bdevConf)
	}
	return bdevConf, err
}

// CollectBlockDeviceInfo collects all needed information about
// block devices from /sys/block
// write info to /var/lib/saptune/sections/block.run
func CollectBlockDeviceInfo() []string {
	bdevConf := BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}
	blockMap := make(map[string]string)

	// List /sys/block and inspect the needed info of each one
	_, sysDevs := ListDir("/sys/block", "the available block devices of the system")
	for _, bdev := range sysDevs {
		if !BlockDeviceIsDisk(bdev) {
			// skip unsupported devices
			WarningLog("skipping device '%s', unsupported", bdev)
			continue
		}
		// add new block device
		blockMap = make(map[string]string)

		// Remember, GetSysChoice does not accept the leading /sys/
		elev, _ := GetSysChoice(path.Join("block", bdev, "queue", "scheduler"))
		blockMap["IO_SCHEDULER"] = elev
		val, err := ioutil.ReadFile(path.Join("/sys/block/", bdev, "/queue/scheduler"))
		sched := ""
		if err == nil {
			sched = string(val)
		}
		blockMap["VALID_SCHEDS"] = sched

		// Remember, GetSysString does not accept the leading /sys/
		nrreq, _ := GetSysString(path.Join("block", bdev, "queue", "nr_requests"))
		blockMap["NRREQ"] = nrreq

		readahead, _ := GetSysString(path.Join("block", bdev, "queue", "read_ahead_kb"))
		blockMap["READ_AHEAD_KB"] = readahead

		// future use
		// VENDOR, TYPE for FUJITSU udev replacement
		// vend := GetDMIDecode(bdev, "VENDOR")
		// blockMap["VENDOR"] = vendor
		// blckType := GetDMIDecode(bdev, "TYPE")
		// blockMap["TYPE""] = blckType
		// ... more to come

		// end of sys/block loop
		// save block info
		bdevConf.BlockAttributes[bdev] = blockMap
		bdevConf.AllBlockDevs = append(bdevConf.AllBlockDevs, bdev)
	}

	err := storeBlockDeviceInfo(bdevConf)
	if err != nil {
		ErrorLog("could not store block device information - err: %v", err)
	}
	return bdevConf.AllBlockDevs
}

// storeBlockDeviceInfo stores block device information to file blockdev.run
// only used in txtparser
// storeSectionInfo stores INIFile section information to section directory
func storeBlockDeviceInfo(obj BlockDev) error {
	overwriteExisting := true
	bdevFileName := fmt.Sprintf("%s/blockdev.run", SaptuneSectionDir)

	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(SaptuneSectionDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(bdevFileName); os.IsNotExist(err) || overwriteExisting {
		return ioutil.WriteFile(bdevFileName, content, 0644)
	}
	return nil
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

// isOwnLock return true, if lock file is from the current running process
// pid inside the lock file is the pid of current running saptune instance
func isOwnLock() bool {
	if !saptuneIsLocked() {
		// no lock file found, return false
		return false
	}
	p, err := ioutil.ReadFile(stLockFile)
	if err != nil {
		ErrorLog("problems during reading the lock file - '%v'", err)
		ReleaseSaptuneLock()
		OSExit(99)
	}
	// file exists, check if empty or if pid inside is from a dead process
	// if yes, remove file and return false
	pid, _ := strconv.Atoi(string(p))
	if pid == os.Getpid() {
		return true
	}
	return false
}

// SaptuneLock creates the saptune lock file
func SaptuneLock() {
	// check for saptune lock file
	if saptuneIsLocked() {
		ErrorExit("saptune currently in use, try later ...", 11)
	}
	stLock, err := os.OpenFile(stLockFile, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
	if err != nil {
		ErrorExit("problems setting lock", 12)
	} else {
		fmt.Fprintf(stLock, "%d", os.Getpid())
	}
	stLock.Close()
}

// saptuneIsLocked checks, if the lock file for saptune exists
func saptuneIsLocked() bool {
	f, err := os.Stat(stLockFile)
	if os.IsNotExist(err) {
		return false
	}
	// file is empty, remove file and return false
	if f.Size() == 0 {
		ReleaseSaptuneLock()
		return false
	}
	// file exists, read content
	p, err := ioutil.ReadFile(stLockFile)
	if err != nil {
		ErrorLog("problems during reading the lock file - '%v'", err)
		ReleaseSaptuneLock()
		OSExit(99)
	}
	// file contains a pid. Check, if process is still alive
	// if not (dead process) remove file and return false
	// TODO - check, if p is really a pid
	pid, _ := strconv.Atoi(string(p))
	if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
		// process exists, must not be the same process, which
		// created the lock file. Will be checked in ErrorExit
		return true
	}
	// process does not exists
	ReleaseSaptuneLock()
	return false
}

// ReleaseSaptuneLock removes the saptune lock file
func ReleaseSaptuneLock() {
	if err := os.Remove(stLockFile); os.IsNotExist(err) {
		// no lock file available, nothing to do
	} else if err != nil {
		ErrorLog("problems removing lock. Please remove lock file '%s' manually before the next start of saptune.\n", stLockFile)
	}
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
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}
	foldedTxt := words[0]
	spaceLeft := width - len(foldedTxt)
	for _, word := range words[1:] {
		if word == "\n" {
			foldedTxt += word
			continue
		}
		if len(word)+1 > spaceLeft {
			// fold; start next row
			foldedTxt += "\n" + word
			spaceLeft = width - len(word)
		} else {
			foldedTxt += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	folded = strings.Split(foldedTxt, "\n")
	return
}
