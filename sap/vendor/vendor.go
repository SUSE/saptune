package vendor

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/system"
	"io/ioutil"
	"regexp"
	"strings"
)

const VENDOR_DIR = "/etc/saptune/extra/"
const VENDOR_FILE = "HPE-Recommended_OS_settings.conf"

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

/* will be worked out later
func (vend VendorSettings) Initialise() (Note, error) {
	newV := VendorSettings{}
	return newV, nil
}

func (vend VendorSettings) Optimise() (Note, error) {
	newV := VendorSettings{}
	return newV, nil
}

func (vend VendorSettings) Apply() error {
	return nil
}
*/

func (vend VendorSettings) SetVendorTunes() error {

	// read tunables from vendor file and activate them using SetSysctlString
	vendor_file, err := ioutil.ReadFile(VENDOR_DIR + VENDOR_FILE)
	if err != nil {
		panic(fmt.Errorf("failed to read vendor file: %v", err))
	}

	var consecSpaces = regexp.MustCompile("[[:space:]]+")
	var tunable, value string
	var fstart = 0

	vendor_content := string(vendor_file)
	for _, line := range strings.Split(vendor_content, "\n") {
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
	}
	return nil
}
