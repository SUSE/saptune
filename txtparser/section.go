package txtparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// OverrideTuningSheets defines saptunes override directory
const OverrideTuningSheets = "/etc/saptune/override/"

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
		return ioutil.WriteFile(iniFileName, content, 0644)
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

	content, err := ioutil.ReadFile(iniFileName)
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
