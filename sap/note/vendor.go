package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/system"
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"strings"
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
		contentBytes, err := ioutil.ReadFile(path.Join(VENDOR_DIR, iniFile))
		log.Printf("VendorSettings.Apply: applying from file %s", iniFile)
		if err != nil {
			return fmt.Errorf("failed to read vendor file: %v", err)
		}

		var tunable, value string
		var fstart = 0

		for _, line := range strings.Split(string(contentBytes), "\n") {
			fields := consecSpaces.Split(strings.TrimSpace(line), -1)
			if len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#' {
				continue // skip comments and empty lines
			}
			if len(fields) < 3 { // handle tuning lines without spaces
				fields = strings.Split(strings.TrimSpace(line), "=")
				if len(fields) == 1 {
					continue
				}
				fstart = 1
			} else {
				if fields[1] != "=" {
					continue
				}
				fstart = 2
			}
			value = fields[fstart]
			for i := fstart + 1; i < len(fields); i++ { // handle tunables with more than one value
				value = value + " " + fields[i]
			}
			tunable = fields[0]

			system.SetSysctlString(tunable, value)
			log.Printf("VendorSettings.Apply: set %s=%s", tunable, value)
		}
	}
	return nil
}
