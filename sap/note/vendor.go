package note

import (
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"fmt"
	"log"
	"regexp"
	"path"
	"encoding/json"
	"io/ioutil"
	"os"
)

const VENDOR_DIR = "/etc/saptune/extra/"
const STATE_DIR = "/var/lib/saptune/saved_state"

var consecSpaces = regexp.MustCompile("[[:space:]]+")

/*
handle vendor specific sysconfig files
see /etc/saptune/extra/HPE-Recommended_OS_settings.conf
*/
type VendorSettings struct {
	SetValues string
}

func (vend VendorSettings) Name() string {
	return "Vendor specific optimization"
}

func (vend VendorSettings) Initialise() (Note, error) {
	// TODO: implement this to fully integrate extra vender settings with saptune advantages
	return nil, nil
}

func (vend VendorSettings) Optimise() (Note, error) {
	// TODO: implement this to fully integrate extra vender settings with saptune advantages
	return nil, nil
}

func (vend VendorSettings) Apply() (savedVendorFiles []string, err error) {
	// Act upon all vendor customisation files
	savedVendorFiles = make([]string, 0, 0)
	_, files, err := system.ListDir(VENDOR_DIR)
	if err != nil {
		return savedVendorFiles, err
	}
	for _, iniFile := range files {
		skv, inierr := txtparser.ParseIniFile(iniFile)
		if inierr != nil {
			return savedVendorFiles, err
		}
		act_cont := &txtparser.Iniconf{
			AllValues: make([]*txtparser.IniEntry, 0, 0),
			KeyValue:  make(map[string]*txtparser.IniEntry),
		}

		for _, entry := range skv.KeyValue {
			section := entry.Section
			tunable := entry.Key
			value   := entry.Value
			if section == "[sysctl]" {
				act_value := system.GetSysctlString(tunable, "")
				act_skv := &txtparser.IniEntry{
					Section: section,
					Key:     tunable,
					Value:   act_value,
				}

				act_cont.AllValues = append(act_cont.AllValues, act_skv)
				act_cont.KeyValue[tunable] = act_skv

				system.SetSysctlString(tunable, value)
				log.Printf("VendorSettings.Apply: set %s=%s", tunable, value)
			}
			// else for future use
		}

		savedVendorFiles = append(savedVendorFiles, iniFile)
		// store previouse values for later revert
		// overwrite existing state file to have every time the correct values for revert
		if err = vend.StoreVendorTunes(iniFile, act_cont, true); err != nil {
			return savedVendorFiles, fmt.Errorf("Failed to save current state of vendor file %s - %v", iniFile, err)
		}
	}
	return
}

func (vend VendorSettings) StoreVendorTunes(fname string, obj *txtparser.Iniconf, overwriteExisting bool) error {
	content, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	path := path.Join(STATE_DIR, fname)

	if _, err := os.Stat(path); os.IsNotExist(err) || overwriteExisting {
		return ioutil.WriteFile(path, content, 0644)
	}
	return nil
}

func (vend VendorSettings) RevertVendorSettings (savedVendorFiles []string) error {
	for _, savedVFiles := range savedVendorFiles {
		path := path.Join(STATE_DIR, savedVFiles)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("Failed to revert vendor settings because of problems with state file '%s'", path)
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Failed to revert vendor settings. Can't read state file '%s'", path)
		}

		act_cont := &txtparser.Iniconf{
			AllValues: make([]*txtparser.IniEntry, 0, 0),
			KeyValue:  make(map[string]*txtparser.IniEntry),
		}

		err = json.Unmarshal(content, &act_cont)
		if err != nil {
			return fmt.Errorf("Failed to revert vendor settings. Can't read content of state file '%s'", path)
		}

		for _, entry := range act_cont.KeyValue {
			section := entry.Section
			tunable := entry.Key
			value   := entry.Value
			if section == "[sysctl]" {
				system.SetSysctlString(tunable, value)
			}
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("Failed to remove state file '%s' of vendor settings", path)
		}
	}
	return nil
}
