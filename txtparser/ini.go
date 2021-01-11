package txtparser

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
)

// Operator definitions
const (
	OperatorLessThan      = "<"
	OperatorLessThanEqual = "<="
	OperatorMoreThan      = ">"
	OperatorMoreThanEqual = ">="
	OperatorEqual         = "="
)

// Operator is the comparison or assignment operator used in an INI file entry
type Operator string

// RegexKeyOperatorValue breaks up a line into key, operator, value.
var RegexKeyOperatorValue = regexp.MustCompile(`([\w.+_-]+)\s*([<=>]+)\s*["']*(.*?)["']*$`)

// counter to control the [login] section info message
var loginCnt = 0

// counter to control the [block] section detected warning
var blckCnt = 0

var blockDev = make([]string, 0, 10)

// INIEntry contains a single key-value pair in INI file.
type INIEntry struct {
	Section  string
	Key      string
	Operator Operator
	Value    string
}

// INIFile contains all key-value pairs of an INI file.
type INIFile struct {
	AllValues []INIEntry
	KeyValue  map[string]map[string]INIEntry
}

// GetINIFileDescriptiveName return the descriptive name of the Note
func GetINIFileDescriptiveName(fileName string) string {
	var re = regexp.MustCompile(`# .*NOTE=.*VERSION=(\d*)\s*DATE=(.*)\s*NAME="([^"]*)"`)
	rval := ""
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(string(content))
	if len(matches) != 0 {
		rval = fmt.Sprintf("%s\n\t\t\t%sVersion %s from %s", matches[3], "", matches[1], matches[2])
	}
	return rval
}

// GetINIFileVersionSectionEntry returns the field 'entryName' from the version
// section of the Note configuration file
func GetINIFileVersionSectionEntry(fileName, entryName string) string {
	var re = regexp.MustCompile(`# .*NOTE=.*TEST=(\d*)\s*DATE=.*"`)
	switch entryName {
	case "version":
		re = regexp.MustCompile(`# .*NOTE=.*VERSION=(\d*)\s*DATE=.*"`)
	case "category":
		re = regexp.MustCompile(`# .*NOTE=.*CATEGORY=(\w*)\s*VERSION=.*"`)
	case "date":
		re = regexp.MustCompile(`# .*NOTE=.*VERSION=\d*\s*DATE=(.*)\s*NAME=.*"`)
	case "name":
		re = regexp.MustCompile(`# .*NOTE=.*VERSION=\d*\s*DATE=.*\s*NAME="([^"]*)"`)
	default:
		return ""
	}
	rval := ""
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(string(content))
	if len(matches) != 0 {
		rval = fmt.Sprintf("%s", strings.TrimSpace(matches[1]))
	}
	return rval
}

// splitLineIntoKOV break apart a line into key, operator, value.
func splitLineIntoKOV(curSection, line string) []string {
	kov := make([]string, 0)
	if curSection == "rpm" {
		kov = splitRPM(line)
	} else {
		kov = RegexKeyOperatorValue.FindStringSubmatch(line)
		if curSection == "grub" {
			kov = splitGrub(line, kov)
		} else if curSection == "service" {
			kov = splitService(line, kov)
		}
	}
	return kov
}

// splitService split line of section service into the needed syntax
func splitService(line string, kov []string) []string {
	if len(kov) == 0 {
		// seams to be a single option and not
		// a key=value pair
		kov = []string{line, "systemd:" + line, "=", "unsupported"}
	} else {
		kov[1] = "systemd:" + kov[1]
	}
	return kov
}

// splitRPM split line of section rpm into the needed syntax
func splitRPM(line string) []string {
	fields := strings.Fields(line)
	kov := make([]string, 0)
	kov = nil
	if len(fields) == 3 {
		// old syntax - rpm to check | os version | expected package version
		// kov needs 3 fields (parameter, operator, value)
		// to not get confused let operator empty, it's not needed for rpm check
		// to be compatible to old section definitions without 'tags' we need to check fields[1] for os matching
		if fields[1] == "all" || fields[1] == system.GetOsVers() {
			kov = []string{"rpm", "rpm:" + fields[0], "", fields[2]}
		} else {
			system.WarningLog("in section 'rpm' the line '%v' contains a non-matching os version '%s'. Skipping line", fields, fields[1])
		}
	} else if len(fields) == 2 {
		// new syntax - rpm to check | expected package version
		// os and/or arch are set with section tags
		// kov needs 3 fields (parameter, operator, value)
		// to not get confused let operator empty, it's not needed for rpm check
		kov = []string{"rpm", "rpm:" + fields[0], "", fields[1]}
	} else {
		// wrong syntax
		system.WarningLog("[rpm] section contains a line with wrong syntax - '%v', skipping entry. Please check", fields)
	}
	return kov
}

// splitGrub split line of section grub into the needed syntax
func splitGrub(line string, kov []string) []string {
	if len(kov) == 0 {
		// seams to be a single option and not
		// a key=value pair
		kov = []string{line, "grub:" + line, "=", line}
	} else {
		kov[1] = "grub:" + kov[1]
	}
	return kov
}

// chkSecTags checks, if the tags of a section are valid
func chkSecTags(secFields, blkDev []string) (bool, []string) {
	ret := true
	cnt := 0
	for _, secTag := range secFields {
		if cnt == 0 {
			// skip section name
			cnt = cnt + 1
			continue
		}
		if secTag == "" {
			// support empty tags
			continue
		}
		tagField := strings.Split(secTag, "=")
		if len(tagField) != 2 {
			system.WarningLog("wrong syntax of section tag '%s', skipping whole section '%v'. Please check. ", secTag, secFields)
			return false, blkDev
		}
		switch tagField[0] {
		case "os":
			ret = chkOsTags(tagField[1], secFields)
		case "arch":
			ret = chkArchTags(tagField[1], secFields)
		case "csp":
			ret = chkCspTags(tagField[1], secFields)
		case "vendor", "model":
			ret, blkDev = chkBlkTags(tagField[0], tagField[1], secFields, blkDev)
		default:
			system.WarningLog("skip unkown section tag '%v'.", secTag)
			ret = false
		}
	}
	return ret, blkDev
}

// chkOsTags checks if the os section tag is valid or not
func chkOsTags(tagField string, secFields []string) bool {
	ret := true
	osWild := regexp.MustCompile(`(.*)-(\*)`)
	osw := osWild.FindStringSubmatch(tagField)
	if len(osw) != 3 {
		if tagField != system.GetOsVers() {
			// os version does not match
			system.WarningLog("os version '%s' in section definition '%v' does not match running os version '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, system.GetOsVers())
			ret = false
		}
	} else if osw[2] == "*" {
		// wildcard
		switch osw[1] {
		case "15":
			if !system.IsSLE15() {
				system.WarningLog("os version '%s' in section definition '%v' does not match running os version '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, system.GetOsVers())
				ret = false
			}
		case "12":
			if !system.IsSLE12() {
				system.WarningLog("os version '%s' in section definition '%v' does not match running os version '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, system.GetOsVers())
				ret = false
			}
		default:
			system.WarningLog("unsupported os version '%s' in section definition '%v'. Skipping whole section with all lines till next valid section definition", tagField, secFields)
			ret = false
		}
	}
	return ret
}

// chkArchTags checks if the arch section tag is valid or not
func chkArchTags(tagField string, secFields []string) bool {
	ret := true
	chkArch := runtime.GOARCH
	if chkArch == "amd64" {
		// map architecture to 'uname -i' output
		chkArch = "x86_64"
	}
	if tagField != chkArch {
		// arch does not match
		system.WarningLog("system architecture '%s' in section definition '%v' does not match the architecture of the running system '%s'. Skipping whole section with all lines till next valid section definition", tagField, secFields, chkArch)
		ret = false
	}
	return ret
}

// chkCsp checks if the csp section tag is valid or not
func chkCspTags(tagField string, secFields []string) bool {
	ret := true
	chkCsp := system.GetCSP()
	if tagField != chkCsp {
		// csp does not match
		if chkCsp == "" {
			chkCsp = "not a cloud"
		}
		system.WarningLog("cloud service provider '%s' in section definition '%v' does not match the cloud service provider of the running system ('%s'). Skipping whole section with all lines till next valid section definition", tagField, secFields, chkCsp)
		ret = false
	}
	return ret
}

// chkBlkTags checks if the vendor or model section tag is valid or not
// and returns a list of valid block devices
func chkBlkTags(info, tagField string, secFields, actbdev []string) (bool, []string) {
	ret := false
	blkInfo := strings.ToUpper(info)
	tagExpr := fmt.Sprintf(".*%s.*", tagField)
	bdev := system.GetAvailBlockInfo(blkInfo, tagExpr)
	if len(bdev) == 0 {
		// vendor or model does not match
		system.WarningLog("%s '%s' in section definition '%v' does not match any available block device %s of the running system. Skipping whole section with all lines till next valid section definition", info, tagField, secFields, info)
	} else {
		if len(actbdev) == 0 {
			bdev = actbdev
		} else {
			// as it is possible to have more than one tag in a
			// section (vendor and module) we need the overlap for
			// a valid result
			// an former call has returned a list of valid block
			// devices and this call also returned a list of valid
			// block devices - get overlap or empty
			newbdev := []string{}
			for _, a := range actbdev {
				for _, b := range bdev {
					if a == b {
						newbdev = append(newbdev, a)
					}
				}
			}
			bdev = newbdev
			if len(newbdev) != 0 {
				ret = true
			}
		}
	}
	return ret, bdev
}

// ParseINIFile read the content of the configuration file
func ParseINIFile(fileName string, autoCreate bool) (*INIFile, error) {
	content, err := system.ReadConfigFile(fileName, autoCreate)
	if err != nil {
		return nil, err
	}
	return ParseINI(string(content)), nil
}

// ParseINI parse the content of the configuration file
func ParseINI(input string) *INIFile {
	ret := &INIFile{
		AllValues: make([]INIEntry, 0, 64),
		KeyValue:  make(map[string]map[string]INIEntry),
	}

	reminder := ""
	bdevs := []string{}
	skipSection := false
	currentSection := ""
	currentEntriesArray := make([]INIEntry, 0, 8)
	currentEntriesMap := make(map[string]INIEntry)
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// skip empty lines
			continue
		}
		if line[0] != '[' && skipSection {
			// skip all lines from a non-valid section
			continue
		}
		if line[0] == '[' {
			// Save previous section, if valid
			if currentSection != "" && !skipSection {
				ret.KeyValue[currentSection] = currentEntriesMap
				ret.AllValues = append(ret.AllValues, currentEntriesArray...)
			}

			// Start a new section
			chkOk := true
			if skipSection {
				skipSection = false
			}
			currentSection = line[1 : len(line)-1]
			if currentSection == "" {
				// empty section line [], skip whole section
				system.WarningLog("found empty section definition []. Skipping whole section with all lines till next valid section definition")
				skipSection = true
				continue
			}
			sectionFields := strings.Split(currentSection, ":")

			// moved block device colletion so that the info can be
			// used inside the 'tag' checks
			if sectionFields[0] == "block" && blckCnt == 0 {
				system.WarningLog("[block] section detected: Traversing all block devices can take a considerable amount of time.")
				blckCnt = blckCnt + 1
				// blockDev all valid block devices of the
				// system regardless of any section tag
				blockDev = system.CollectBlockDeviceInfo()
			}
			bdevs = blockDev
			// len(sectionFields) == 1 - standard syntax [section], no os or arch check needed, chkOk = true
			if len(sectionFields) > 1 {
				// check of section tags needed
				chkOk, bdevs = chkSecTags(sectionFields, bdevs)
			}
			if chkOk {
				currentSection = sectionFields[0]
				currentEntriesArray = make([]INIEntry, 0, 8)
				currentEntriesMap = make(map[string]INIEntry)
			} else {
				// skip non-valid section with all lines
				skipSection = true
			}
			continue
		}
		if strings.HasPrefix(line, "#") {
			// Skip comments. Need to be done before
			// 'break apart the line into key, operator, value'
			// to support comments like # something (default = 60)
			// without side effects
			if currentSection == "reminder" {
				reminder = reminder + line + "\n"
			}
			continue
		}
		// Break apart a line into key, operator, value.
		kov := splitLineIntoKOV(currentSection, line)
		if kov == nil {
			// Skip comments, empty, and irregular lines.
			continue
		}
		if kov[1] == "UserTasksMax" && system.IsSLE15() {
			if loginCnt == 0 {
				system.InfoLog("UserTasksMax setting no longer supported on SLE15 releases. Leaving system's default unchanged.")
			}
			loginCnt = loginCnt + 1
			continue
		}
		if currentSection == "limits" {
			for _, limits := range strings.Split(kov[3], ",") {
				limits = strings.TrimSpace(limits)
				lim := strings.Fields(limits)
				key := ""
				if len(lim) == 0 {
					// empty LIMITS parameter means
					// override file is setting all limits to 'untouched'
					// or a wrong limits entry in an 'extra' file
					key = fmt.Sprintf("%s_NA", kov[1])
					limits = "NA"
				} else {
					key = fmt.Sprintf("LIMIT_%s_%s_%s", lim[0], lim[1], lim[2])
				}
				entry := INIEntry{
					Section:  currentSection,
					Key:      key,
					Operator: Operator(kov[2]),
					Value:    limits,
				}
				currentEntriesArray = append(currentEntriesArray, entry)
				currentEntriesMap[entry.Key] = entry
			}
		} else if currentSection == "block" {
			// bdevs contains all block devices valid for the
			// current block section regarding to the used tags
			for _, bdev := range bdevs {
				entry := INIEntry{
					Section:  currentSection,
					Key:      fmt.Sprintf("%s_%s", kov[1], bdev),
					Operator: Operator(kov[2]),
					Value:    kov[3],
				}
				currentEntriesArray = append(currentEntriesArray, entry)
				currentEntriesMap[entry.Key] = entry
			}
		} else {
			// handle tunables with more than one value
			value := strings.Replace(kov[3], " ", "\t", -1)
			entry := INIEntry{
				Section:  currentSection,
				Key:      kov[1],
				Operator: Operator(kov[2]),
				Value:    value,
			}
			currentEntriesArray = append(currentEntriesArray, entry)
			currentEntriesMap[entry.Key] = entry
		}
	}
	if reminder != "" {
		// save reminder section
		// Save previous section
		if currentSection != "" {
			ret.KeyValue[currentSection] = currentEntriesMap
			ret.AllValues = append(ret.AllValues, currentEntriesArray...)
		}
		// Start the reminder section
		currentEntriesArray = make([]INIEntry, 0, 8)
		currentEntriesMap = make(map[string]INIEntry)
		currentSection = "reminder"

		entry := INIEntry{
			Section:  "reminder",
			Key:      "reminder",
			Operator: "",
			Value:    reminder,
		}
		currentEntriesArray = append(currentEntriesArray, entry)
		currentEntriesMap[entry.Key] = entry
	}

	// Save last section
	if currentSection != "" {
		ret.KeyValue[currentSection] = currentEntriesMap
		ret.AllValues = append(ret.AllValues, currentEntriesArray...)
	}
	return ret
}
