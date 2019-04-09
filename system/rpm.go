package system

// wrapper to rpm command

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

var alphanumPattern = regexp.MustCompile("([a-zA-Z]+)|([0-9]+)|(~)")

// GetRpmVers return the version of an installed RPM
func GetRpmVers(rpm string) string {
	// rpm -q --qf '%{VERSION}-%{RELEASE}\n' glibc
	notInstalled := fmt.Sprintf("package %s is not installed", rpm)
	rpmVers := ""
	cmdName := "/bin/rpm"
	cmdArgs := []string{"-q", "--qf", "%{VERSION}-%{RELEASE}\n", rpm}

	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		if len(string(cmdOut)) == 0 || strings.TrimSpace(string(cmdOut)) != notInstalled {
			WarningLog("There was an error running external command 'rpm -q --qf '%%{VERSION}-%%{RELEASE}' %s': %v, output: %s", rpm, err, cmdOut)
		}
		return ""
	}
	for _, line := range strings.Split(strings.TrimSpace(string(cmdOut)), "\n") {
		// ANGI: TODO - was, wenn mehr als eine Paketversion installiert ist
		rpmVers = line
	}
	return rpmVers
}

/* compare rpm versions
func CmpRpmVers, func CheckRpmVers
build on base of information from http://rpm.org/user_doc/dependencies.html
and https://github.com/rpm-software-management/rpm/blob/master/lib/rpmvercmp.c
Equal == 0, GreaterThan > 0, LessThan < 0
vers1 is '228-150.22.1', vers2 is '228-142.1'
*/

// CmpRpmVers compare versions of 2 RPMs (installed version, expected version)
// Return true, if installed package version is equal or higher than expected
// Return false, if installed package version is less than expected
func CmpRpmVers(vers1, vers2 string) bool {
	if vers1 == "" {
		// package not installed
		return false
	}
	if vers1 == vers2 {
		// rpm version and release are equal
		return true
	}
	// actV is 228-150.22.1, expV is 228-142.1
	actV := strings.Split(vers1, "-")
	expV := strings.Split(vers2, "-")
	// check rpm version
	ret := CheckRpmVers(actV[0], expV[0])
	if ret > 0 {
		// installed package version is higher than expected
		return true
	} else if ret < 0 {
		// installed package version is less than expected
		return false
	}
	// rpm version is equal, so check rpm release
	ret = CheckRpmVers(actV[1], expV[1])
	if ret < 0 {
		// installed package release is less than expected
		return false
	}
	// installed package release is equal or higher than expected
	return true
}

// CheckRpmVers compare versions of 2 RPMs (installed version, expected version)
// Return 0 (Equal), 1 (GreaterThan) or -1 (LessThan)
func CheckRpmVers(vers1, vers2 string) int {
	// per definition numbers are greater than alphas
	if vers1 == vers2 {
		return 0
	}
	// get bunches of numbers or characters for comparision
	partsV1 := alphanumPattern.FindAllString(vers1, -1)
	partsV2 := alphanumPattern.FindAllString(vers2, -1)
	nrParts := len(partsV1)
	if len(partsV1) > len(partsV2) {
		nrParts = len(partsV2)
	}
	// compare each bunche of numbers or characters
	for i := 0; i < nrParts; i++ {
		p1 := partsV1[i]
		p2 := partsV2[i]
		r10 := []rune(p1)[0]
		r20 := []rune(p2)[0]
		// searching for 'tildes'
		// first character in bunch - []rune(p1)[0]
		if r10 == '~' || r20 == '~' {
			if r10 != '~' {
				return 1
			}
			if r20 != '~' {
				return -1
			}
		}
		if unicode.IsNumber(r10) {
			// actual vers part is a number
			if !unicode.IsNumber(r20) {
				// actual vers is numeric, expected vers is alpha, so actual vers is higher
				return 1
			}
			// both are numbers, trim leading zeros
			p1 = strings.TrimLeft(p1, "0")
			p2 = strings.TrimLeft(p2, "0")
			// longest string wins, no need for further comparison
			if len(p1) > len(p2) {
				return 1
			} else if len(p2) > len(p1) {
				return -1
			}
		} else if unicode.IsNumber(r20) {
			// actual vers part is a alpha, but expected vers part is a number
			return -1
		}
		// both parts are alpha, so use simple string compare
		if p1 < p2 {
			return -1
		} else if p1 > p2 {
			return 1
		}
	}
	// the bunches were all the same. Check, if separators of bunches have been different
	if len(partsV1) == len(partsV2) {
		return 0
	}
	// look for a tilde in a bunch/part past the minimal number of bunches/parts
	// could not be found in the for loop above because it's outside of the loop range
	if len(partsV1) > nrParts && []rune(partsV1[nrParts])[0] == '~' {
		return -1
	} else if len(partsV2) > nrParts && []rune(partsV2[nrParts])[0] == '~' {
		return 1
	}
	// at least the highest number of bunches/parts wins
	if len(partsV1) > len(partsV2) {
		return 1
	}
	return -1
}
