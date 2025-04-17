package system

import (
	"os"
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
	// IBM Cloud VPC (Not IBM Cloud Classic)
	CSPIBMVPC     = "ibmVPC"
	CSPIBMVPCLong = "IBM Cloud Virtual Server for VPC"
)

var dmiDir = "/sys/class/dmi"

// CSP identifier
var isAzureCat = regexp.MustCompile(`.*(7783-7084-3265-9085-8269-3286-77|MSFT AZURE VM).*`)
var isAzure = regexp.MustCompile(`.*[mM]icrosoft [cC]orporation.*`)
var isAWS = regexp.MustCompile(`.*[aA]mazon.*`)
var isGoogle = regexp.MustCompile(`.*[gG]oogle.*`)
var isOVM = regexp.MustCompile(`.*OVM.*`)
var isAlibaba = regexp.MustCompile(`.*[aA]libaba.*`)
var isIBMVPCCat = regexp.MustCompile(`.*ibmcloud.*`)
var isIBMVPC = regexp.MustCompile(`.*IBM:Cloud Compute Server 1.0:.*`)

type manufacturerProviders struct {
	Manufacturer string
	Providers    map[*regexp.Regexp]string
}

var allManufacturerProviders = [...]manufacturerProviders{
	// dmidecode key files
	// /usr/sbin/dmidecode -s chassis-asset-tag
	{"/sys/class/dmi/id/chassis_asset_tag", map[*regexp.Regexp]string{isAzureCat: CSPAzure, isIBMVPCCat: CSPIBMVPC}},
	// /usr/sbin/dmidecode -s system-manufacturer
	{"/sys/class/dmi/id/system-manufacturer", map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
	// /usr/sbin/dmidecode -s board-vendor
	{"/sys/class/dmi/id/board_vendor", map[*regexp.Regexp]string{isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s bios-version
	{"/sys/class/dmi/id/bios_version", map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
	// /usr/sbin/dmidecode -s bios-vendor
	{"/sys/class/dmi/id/bios_vendor", map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s system-version
	{"/sys/class/dmi/id/system_version", map[*regexp.Regexp]string{isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s sys-vendor
	{"/sys/class/dmi/id/sys_vendor", map[*regexp.Regexp]string{isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s chassis-manufacturer
	{"/sys/class/dmi/id/chassis_vendor", map[*regexp.Regexp]string{isIBMVPC: CSPIBMVPC}},
}

// GetDMIDecode
//func GetDMIDecode(key string) string {
// system_version = "dmidecode -s system-version"
//}

// GetCSP returns the identifier of the cloud service provider of the current
// running system or an empty string, if the system does not belong to a CSP
// use files in /sys/class/dmi/id/ instead of dmidecode command
func GetCSP() string {
	cloudServiceProvider := ""
	getCloudServiceProvider := func(manufacturer string, providers map[*regexp.Regexp]string) string {
		if content, err := os.ReadFile(manufacturer); err == nil {
			for providerRegex, provider := range providers {
				matches := providerRegex.FindStringSubmatch(string(content))
				if len(matches) != 0 {
					return provider
				}
			}
		}
		return ""
	}

	if _, err := os.Stat(dmiDir); os.IsNotExist(err) {
		InfoLog("directory '%s' does not exist", dmiDir)
		return cloudServiceProvider
	}

	for _, mp := range allManufacturerProviders {
		if cloudServiceProvider == "" {
			cloudServiceProvider = getCloudServiceProvider(mp.Manufacturer, mp.Providers)
		} else {
			break
		}
	}
	return cloudServiceProvider
}
