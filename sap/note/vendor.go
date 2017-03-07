package note

import (
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"log"
	"regexp"
)

const VENDOR_DIR = "/etc/saptune/extra/"

var consecSpaces = regexp.MustCompile("[[:space:]]+")

/*
handle vendor specific sysconfig files
see /etc/saptune/extra/HPE-Recommended_OS_settings.conf
*/
type VendorSettings struct {
	SetValues string
}

func (vend VendorSettings) Name() string {
	return "HPE - Vendor specific optimization"
}

func (vend VendorSettings) Initialise() (Note, error) {
	// TODO: implement this to fully integrate extra vender settings with saptune advantages
	return nil, nil
}

func (vend VendorSettings) Optimise() (Note, error) {
	// TODO: implement this to fully integrate extra vender settings with saptune advantages
	return nil, nil
}

func (vend VendorSettings) Apply() error {
	// Act upon all vendor customisation files
	_, files, err := system.ListDir(VENDOR_DIR)
	if err != nil {
		return err
	}
	for _, iniFile := range files {
		skv, inierr := txtparser.ParseIniFile(iniFile)
		if inierr != nil {
			return err
		}
		for _, entry := range skv.KeyValue {
			section := entry.Section
			tunable := entry.Key
			value   := entry.Value
			if section == "[sysctl]" {
				system.SetSysctlString(tunable, value)
				log.Printf("VendorSettings.Apply: set %s=%s", tunable, value)
			}
			// else for future use
		}
	}
	return nil
}
