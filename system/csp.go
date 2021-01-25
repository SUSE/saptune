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
	// amazon-web-services
	CSPAWS     = "aws"
	CSPAWSLong = "Amazon Web Services"
	// GoogleCloud
	CSPGoogle     = "google"
	CSPGoogleLong = "Google Cloud Platform"
	// OracleCloud
	CSPOVM     = "ovm"
	CSPOVMLong = "Oracle Cloud"
	// Alibaba Cloud
	CSPAlibaba     = "alibaba"
	CSPAlibabaLong = "Alibaba Cloud"
)

// dmidecode key files
// /usr/sbin/dmidecode -s chassis-asset-tag
var dmiChassisAssetTag = "/sys/class/dmi/id/chassis_asset_tag"

// /usr/sbin/dmidecode -s board-vendor
var dmiBoardVendor = "/sys/class/dmi/id/board_vendor"

// /usr/sbin/dmidecode -s bios-vendor
var dmiBiosVendor = "/sys/class/dmi/id/bios_vendor"

// /usr/sbin/dmidecode -s bios-version
var dmiBiosVersion = "/sys/class/dmi/id/bios_version"

// /usr/sbin/dmidecode -s system-version
var dmiSystemVersion = "/sys/class/dmi/id/system_version"

// /usr/sbin/dmidecode -s system-manufacturer
var dmiSystemManufacturer = "/sys/class/dmi/id/system-manufacturer"

// CSP identifier
var isAzureCat = regexp.MustCompile(`.*(7783-7084-3265-9085-8269-3286-77|MSFT AZURE VM).*`)
var isAzure = regexp.MustCompile(`.*[mM]icrosoft [cC]orporation.*`)
var isAWS = regexp.MustCompile(`.*[aA]mazon.*`)
var isGoogle = regexp.MustCompile(`.*[gG]oogle.*`)
var isOVM = regexp.MustCompile(`.*OVM.*`)
var isAlibaba = regexp.MustCompile(`.*[aA]libaba.*`)

// GetDMIDecode
//func GetDMIDecode(key string) string {
// system_version = "dmidecode -s system-version"
//}

// GetCSP returns the identifier of the cloud service provider of the current
// running system or an empty string, if the system does not belong to a CSP
// use files in /sys/class/dmi/id/ instead of dmidecode command
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
			// check for Google
			matches = isGoogle.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPGoogle
			}
			// check for Alibaba
			matches = isAlibaba.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPAlibaba
			}
		}
	}
	if csp == "" {
		// BoardVendor
		if content, err := ioutil.ReadFile(dmiBoardVendor); err == nil {
			// check for AWS
			matches := isAWS.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPAWS
			}
		}
	}
	if csp == "" {
		// BiosVersion
		if content, err := ioutil.ReadFile(dmiBiosVersion); err == nil {
			// check for AWS
			matches := isAWS.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPAWS
			}
			// check for Google
			matches = isGoogle.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPGoogle
			}
			// check for Oracle Cloud
			matches = isOVM.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPOVM
			}
		}
	}
	if csp == "" {
		// BiosVendor
		if content, err := ioutil.ReadFile(dmiBiosVendor); err == nil {
			// check for Google
			matches := isGoogle.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPGoogle
			}
			// check for AWS
			matches = isAWS.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPAWS
			}
		}
	}
	if csp == "" {
		// SystemVersion
		if content, err := ioutil.ReadFile(dmiSystemVersion); err == nil {
			// check for AWS
			matches := isAWS.FindStringSubmatch(string(content))
			if len(matches) != 0 {
				csp = CSPAWS
			}
		}
	}
	return csp
}
