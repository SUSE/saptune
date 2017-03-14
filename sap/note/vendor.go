package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"log"
)

// Parse 3rd party tuning option file and return their parameter name vs recommended value.
func Get3rdPartyParameters(confFilePath string) (params map[string]string, err error) {
	params = make(map[string]string)
	// Parse the configuration file
	skv, err := txtparser.ParseIniFile(confFilePath)
	if err != nil {
		return
	}
	for _, entry := range skv.KeyValue {
		if entry.Section != "[sysctl]" {
			// saptune does not yet understand settings outside of [sysctl] section
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", confFilePath, entry.Section)
			continue
		}
		params[entry.Key] = entry.Value
	}
	return
}

// Tuning options composed by a third party vendor.
type VendorSettings struct {
	ConfFilePath    string            // Full path to the 3rd party vendor's tuning configuration file
	ID              string            // ID portion of the tuning configuration
	DescriptiveName string            // Descriptive name portion of the tuning configuration
	SysctlParams    map[string]string // Sysctl parameter values from the computer system
}

func (vend VendorSettings) Name() string {
	return vend.DescriptiveName
}

func (vend VendorSettings) Initialise() (Note, error) {
	params, err := Get3rdPartyParameters(vend.ConfFilePath)
	if err != nil {
		return vend, err
	}
	// Read current parameter values
	vend.SysctlParams = make(map[string]string)
	for param := range params {
		currValue := system.GetSysctlString(param, "")
		if currValue == "" {
			return vend, fmt.Errorf("VendorSettings %s: cannot find parameter \"%s\" in system", vend.ID, param)
		}
		vend.SysctlParams[param] = currValue
	}
	return vend, nil
}

func (vend VendorSettings) Optimise() (Note, error) {
	// To optimise the parameters, simply copy vendor's proposed settings into the system
	params, err := Get3rdPartyParameters(vend.ConfFilePath)
	if err != nil {
		return vend, err
	}
	vend.SysctlParams = params
	return vend, nil
}

func (vend VendorSettings) Apply() error {
	// Apply sysctl parameters
	for key, value := range vend.SysctlParams {
		system.SetSysctlString(key, value)
	}
	return nil
}
