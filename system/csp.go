package system

import (
	"io/ioutil"
	"regexp"
)

// constant definitions
const (
	// Cloud Service Provider short and long names
	// microsoft-azure
	CSPAzure     = "azure"
	CSPAzureLong = "Microsoft Azure"
)

// dmidecode key files
// /usr/sbin/dmidecode -s chassis-asset-tag
var dmiChassisAssetTag = "/sys/devices/virtual/dmi/id/chassis_asset_tag"

// /usr/sbin/dmidecode -s system-manufacturer
var dmiSystemManufacturer = "/sys/devices/virtual/dmi/id/system-manufacturer"

// CSP identifier
var isAzureCat = regexp.MustCompile(`.*(7783-7084-3265-9085-8269-3286-77|MSFT AZURE VM).*`)
var isAzure = regexp.MustCompile(`.*[mM]icrosoft [cC]orporation.*`)

// GetCSP returns the identifier of the cloud service provider of the current
// running system or an empty string, if the system does not belong to a CSP
// use files in /sys/devices/virtual/id/ instead of dmidecode command
func GetCSP() string {
	csp := ""

	// check for Azure
	if content, err := ioutil.ReadFile(dmiChassisAssetTag); err == nil {
		matches := isAzureCat.FindStringSubmatch(string(content))
		if len(matches) != 0 {
			csp = CSPAzure
		}
	}
	if csp == "" {
		// SystemManufacturer
		if content, err := ioutil.ReadFile(dmiSystemManufacturer); err == nil {
			// check for Azure
			matches := isAzure.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPAzure
			}
		}
	}
	return csp
}
