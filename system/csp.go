package system

import (
	"io/ioutil"
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
)

var dmiDir = "/sys/class/dmi"

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

// /usr/sbin/dmidecode -s sys-vendor
var dmiSysVendor = "/sys/class/dmi/id/sys_vendor"

// /usr/sbin/dmidecode -s system-manufacturer
var dmiSystemManufacturer = "/sys/class/dmi/id/system-manufacturer"

// CSP identifier
var isAzureCat = regexp.MustCompile(`.*(7783-7084-3265-9085-8269-3286-77|MSFT AZURE VM).*`)
var isAzure = regexp.MustCompile(`.*[mM]icrosoft [cC]orporation.*`)
var isAWS = regexp.MustCompile(`.*[aA]mazon.*`)
var isGoogle = regexp.MustCompile(`.*[gG]oogle.*`)
var isOVM = regexp.MustCompile(`.*OVM.*`)
var isAlibaba = regexp.MustCompile(`.*[aA]libaba.*`)

type manufacturerProviders struct {
	Manufacturer string
	Providers    map[*regexp.Regexp]string
}

var allManufacturerProviders = [...]manufacturerProviders{
	{dmiChassisAssetTag, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
	{dmiSystemManufacturer, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
	{dmiBoardVendor, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	{dmiBiosVersion, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
	{dmiBiosVendor, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
	{dmiSystemVersion, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	{dmiSysVendor, map[*regexp.Regexp]string{isAWS: CSPAWS}},
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
		if content, err := ioutil.ReadFile(manufacturer); err == nil {
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

	for {
		if cloudServiceProvider == "" {
			for _, mp := range allManufacturerProviders {
				cloudServiceProvider = getCloudServiceProvider(mp.Manufacturer, mp.Providers)
			}
		} else {
			break
		}
	}
	return cloudServiceProvider
}
