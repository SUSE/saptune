package txtparser

import (
	"encoding/json"
	"fmt"
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// OverrideTuningSheets defines saptunes override directory
const OverrideTuningSheets = "/etc/saptune/override/"

var saptuneSectionDir = system.SaptuneSectionDir

// counter to control the warning message for the use of old style
// version section
var oldStyleCnt = map[string]int{"file": 0}

// counter to control the error message of missing or wrong version section
var missVersionCnt = map[string]int{"file": 0}

// StoreSectionInfo stores INIFile section information to section directory
func StoreSectionInfo(obj *INIFile, file, ID string, overwriteExisting bool) error {
	iniFileName := ""
	if file == "run" {
		iniFileName = fmt.Sprintf("%s/%s.run", saptuneSectionDir, ID)
	} else if file == "ovw" {
		iniFileName = fmt.Sprintf("%s/over_%s.run", saptuneSectionDir, ID)
	} else {
		iniFileName = fmt.Sprintf("%s/%s.sections", saptuneSectionDir, ID)
	}
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(saptuneSectionDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(iniFileName); os.IsNotExist(err) || overwriteExisting {
		return os.WriteFile(iniFileName, content, 0644)
	}
	return nil
}

// GetSectionInfo reads content of stored INIFile information.
// Return the content as INIFile
func GetSectionInfo(initype, ID string, fileSelect bool) (*INIFile, error) {
	iniFileName := ""
	if fileSelect {
		iniFileName = fmt.Sprintf("%s/%s.sections", saptuneSectionDir, ID)
	} else if initype == "ovw" {
		iniFileName = fmt.Sprintf("%s/over_%s.run", saptuneSectionDir, ID)
	} else {
		iniFileName = fmt.Sprintf("%s/%s.run", saptuneSectionDir, ID)
	}
	iniConf := &INIFile{
		AllValues: make([]INIEntry, 0, 64),
		KeyValue:  make(map[string]map[string]INIEntry),
	}

	content, err := os.ReadFile(iniFileName)
	if err == nil {
		// do not remove section runtime file, but remove section
		// saved state file after reading
		if fileSelect {
			// remove section saved state file after reading
			err = os.Remove(iniFileName)
		}
		if len(content) != 0 {
			err = json.Unmarshal(content, &iniConf)
		}
	}
	return iniConf, err
}

// GetOverrides is looking for an override file and parse the content
func GetOverrides(filetype, ID string) (bool, *INIFile) {
	override := false
	ow, err := GetSectionInfo(filetype, ID, false)
	if err != nil {
		// Parse the override file
		ow, err = ParseINIFile(path.Join(OverrideTuningSheets, ID), false)
		if err == nil {
			// write section data to section runtime file
			_ = StoreSectionInfo(ow, filetype, ID, true)
			override = true
		}
	} else {
		override = true
	}
	return override, ow
}

// readVersionSection read content of [version] section from config file
func readVersionSection(fileName string) ([]string, bool, error) {
	skipSection := false
	staging := false
	chkVersEntries := map[string]bool{"missing": false, "found": false, "isNew": false, "isOld": false, "skip": false, "mandVers": false, "mandDate": false, "mandDesc": false, "mandRefs": false}
	vsection := []string{}
	fName := filepath.Base(fileName)
	if strings.Contains(filepath.Dir(fileName), "/staging/") {
		staging = true
	}
	versRun := fmt.Sprintf("%s/version_%s.run", saptuneSectionDir, fName)
	// if processing a note from the staging area, read from staging file
	// and NOT from the stored 'run' file
	if _, err := os.Stat(versRun); err == nil && !staging {
		return getVersionRunInfo(versRun)
	}
	content, err := os.ReadFile(fileName)
	if err != nil {
		return vsection, chkVersEntries["isNew"], err
	}
	for _, line := range strings.Split(string(content), "\n") {
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
			if chkVersEntries["found"] {
				// stop reading beyond section [version]
				break
			}
			if skipSection {
				skipSection = false
			}
			if line[1:len(line)-1] != "version" {
				// skip whole section
				skipSection = true
				continue
			}
			// found section [version]
			chkVersEntries["found"] = true
			continue
		}
		chkVersEntriesSyntax(line, chkVersEntries)
		if chkVersEntries["skip"] {
			// Skip comments. But include old style version header
			chkVersEntries["skip"] = false
			continue
		}
		vsection = append(vsection, line)
	}

	err = chkVersEntriesResult(fileName, chkVersEntries)
	// if processing a note from the staging area do NOT store the version
	// info in the 'run' file to not override the section info from the
	// working area
	if !chkVersEntries["missing"] && !staging {
		err = storeVersionRunInfo(versRun, vsection, chkVersEntries["isNew"])
	}
	return vsection, chkVersEntries["isNew"], err
}

// getVersionRunInfo reads content of stored version section info from
// saptuneSectionDir (/run/saptune/sections)
func getVersionRunInfo(versRun string) ([]string, bool, error) {
	var dest []string
	var vsection []string
	isNew := true
	content, err := os.ReadFile(versRun)
	if err == nil && len(content) != 0 {
		err = json.Unmarshal(content, &dest)
		vsection = dest[0 : len(dest)-1]
		if dest[len(dest)-1] == "ISNEW=false" {
			isNew = false
		}
	}
	return vsection, isNew, err
}

// storeVersionRunInfo stores the version section info for re-use
// in saptuneSectionDir (/run/saptune/sections)
func storeVersionRunInfo(versRun string, vsection []string, isNew bool) error {
	var obj []string
	overwriteExisting := true
	if isNew {
		obj = append(vsection, "ISNEW=true")
	} else {
		obj = append(vsection, "ISNEW=false")
	}
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(saptuneSectionDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(versRun); os.IsNotExist(err) || overwriteExisting {
		return os.WriteFile(versRun, content, 0644)
	}
	return nil
}

// chkVersEntriesSyntax checks the version section for syntax errors like
// missing mandatory fields or completely missing version section
func chkVersEntriesSyntax(line string, chkVents map[string]bool) {
	var old = regexp.MustCompile(`# .*NOTE=.*VERSION=([\w.+-_]+)\s*DATE=(.*)\s*NAME="([^"]*)"`)
	matches := old.FindStringSubmatch(line)
	if len(matches) != 0 {
		chkVents["isOld"] = true
	}
	if !chkVents["isOld"] && strings.HasPrefix(line, "#") {
		// Skip comments. But include old style version header
		chkVents["skip"] = true
		return
	}
	// check for the mandatory fields in the new style version section
	ents := map[string]string{"version": "mandVers", "date": "mandDate", "reference": "mandRefs", "description": "mandDesc"}
	for key, ent := range ents {
		re := regexp.MustCompile(newStyleVersionSectionEntry(key))
		matches = re.FindStringSubmatch(line)
		if len(matches) > 1 {
			chkVents[ent] = true
		}
	}
}

// chkVersEntriesResult checks the result of the version section entries check.
// print and return error message
func chkVersEntriesResult(fileName string, chkVents map[string]bool) error {
	var err error
	object := "Note"
	if strings.HasSuffix(fileName, ".sol") {
		object = "Solution"
	}
	if !chkVents["mandVers"] && !chkVents["mandDate"] && !chkVents["mandRefs"] && !chkVents["mandDesc"] {
		chkVents["isNew"] = false
	} else {
		chkVents["isNew"] = true
	}
	if (!chkVents["isNew"] && !chkVents["isOld"]) || !chkVents["found"] {
		// missing version section
		chkVents["missing"] = true
		if missVersionCnt[fileName] < 1 {
			err = system.ErrorLog("missing version section in %s definition file '%s'. Please check", object, fileName)
			missVersionCnt[fileName] = missVersionCnt[fileName] + 1
		} else {
			err = fmt.Errorf("1")
		}
	}
	if chkVents["isNew"] && (!chkVents["mandVers"] || !chkVents["mandDate"] || !chkVents["mandRefs"] || !chkVents["mandDesc"]) {
		// wrong version section
		chkVents["missing"] = true
		if missVersionCnt[fileName] < 1 {
			if chkVents["isOld"] {
				// version section mismatch
				system.ErrorLog("version section mismatch in %s definition file '%s' - old and (partial) new style version header found. Please check", object, fileName)
			}
			system.ErrorLog("wrong version section found in %s definition file '%s'. At least one of the mandatory fields is missing. Please check", object, fileName)
			missVersionCnt[fileName] = missVersionCnt[fileName] + 1
		}
	}
	return err
}

// GetINIFileVersionSectionEntry returns the field 'entryName' from the version
// section of the Note configuration file
func GetINIFileVersionSectionEntry(fileName, entryName string) string {
	var re = regexp.MustCompile(`.*(ID\s*=).*`)
	rval := ""
	content, isNewStyle, err := readVersionSection(fileName)
	if err != nil {
		return ""
	}
	regex := selectVersionExpression(isNewStyle, entryName, fileName)
	re = regexp.MustCompile(regex)
	for _, entryLine := range content {
		matches := re.FindStringSubmatch(entryLine)
		if len(matches) > 1 {
			val := matches[1]
			val = system.StripComment(val, `\s#[^#]|"\s#[^#]`)
			rval = strings.TrimSpace(val)
			break
		}
	}
	return rval
}

// GetINIFileVersionSectionRefs returns the reference field from the version
// section of the Note configuration file
func GetINIFileVersionSectionRefs(fileName string) []string {
	refs := GetINIFileVersionSectionEntry(fileName, "reference")
	rval := strings.Fields(refs)
	return rval
}

// selectVersionExpression returns the regular expression needed to
// identify a specific version section entry
func selectVersionExpression(newStyle bool, entry, file string) string {
	regex := ""
	if newStyle {
		regex = newStyleVersionSectionEntry(entry)
	} else {
		if oldStyleCnt[file] < 1 {
			object := "Note"
			if strings.HasSuffix(file, ".sol") {
				object = "Solution"
			}
			system.WarningLog("You are still using the old style version section syntax in %s definition file '%s' which is deprecated. Please adapt.", object, file)
			oldStyleCnt[file] = oldStyleCnt[file] + 1
		}
		regex = oldStyleVersionSectionEntry(entry)
	}
	return regex
}

// newStyleVersionSectionEntry returns the regular expression to retrieve
// the field 'entryName' from the new style version section of the Note
// configuration file
func newStyleVersionSectionEntry(entryName string) string {
	var re = `^\s*TESTID\s*=\s*"?([^\s].*?)"?$`
	switch entryName {
	//case "id":
	//	re = `^\s*ID\s*=\s*"?([^\s].*?)"?$`
	case "version":
		re = `^\s*VERSION\s*=\s*"?([\w.+-_]+)"?.*$`
	case "category":
		re = `^\s*CATEGORY\s*=\s*"?(\w*)"?$`
	case "reference":
		re = `^\s*REFERENCES\s*=\s*"?(.*)"?$`
	case "date":
		re = `^\s*DATE\s*=\s*"?(\d{2}[-./]{1}\d{2}[-./]{1}\d{4}|\d{4}[-./]{1}\d{2}[-./]{1}\d{2})"?.*$`
	case "name", "description":
		re = `^\s*DESCRIPTION\s*=\s*"?(.*)"?$`
	}
	return re
}

// oldStyleVersionSectionEntry returns the regular expression to retrieve
// the field 'entryName' from the old style version section of the Note
// configuration file
// needed for compatibility reason
func oldStyleVersionSectionEntry(entryName string) string {
	var re = `# .*NOTE=.*TEST=([\w.+-_]+)\s*DATE=.*"`
	switch entryName {
	case "version":
		re = `# .*NOTE=.*VERSION=([\w.+-_]+)\s*DATE=.*"`
	case "category":
		re = `# .*NOTE=.*CATEGORY=(\w*)\s*VERSION=.*"`
	case "reference":
		re = `# .*NOTE=.*REFERENCES="([^"]*)"\s*VERSION=.*"`
	case "date":
		re = `# .*NOTE=.*VERSION=[\w.+-_]+\s*DATE=(.*)\s*NAME=.*"`
	case "name", "description":
		re = `# .*NOTE=.*VERSION=[\w.+-_]+\s*DATE=.*\s*NAME="([^"]*)"`
	}
	return re
}

// ResetVersionSectCnts resets the counters to control
// the error message of missing or wrong version section
// the warning message for the use of old style version section
// needed for staging to get the Warnings and Errors displayed
func ResetVersionSectCnts(tag string) {
	for key := range oldStyleCnt {
		if strings.Contains(filepath.Dir(key), tag) {
			oldStyleCnt[key] = 0
		}
	}
	for key := range missVersionCnt {
		if strings.Contains(filepath.Dir(key), tag) {
			missVersionCnt[key] = 0
		}
	}
}
