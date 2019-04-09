package system

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// IsUserRoot return true only if the current user is root.
func IsUserRoot() bool {
	return os.Getuid() == 0
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
