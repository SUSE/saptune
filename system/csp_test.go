package system

import (
	"os"
	"path"
	"regexp"
	"testing"
)

func TestGetCSP(t *testing.T) {
	// initialize test environment
	noCloud := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_no_cloud")

	// amazon-web-services
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws"), map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val := GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws"), map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws"), map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws"), map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws"), map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}

	// GoogleCloud
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google"), map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google"), map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google"), map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}

	// OracleCloud
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_ovm"), map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "ovm" {
		t.Errorf("Test failed, expected 'ovm', but got '%s'", val)
	}

	// microsoft-azure
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure"), map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "azure" {
		t.Errorf("Test failed, expected 'azure', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure_cat"), map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "azure" {
		t.Errorf("Test failed, expected 'azure', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure_false"), map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "" {
		t.Errorf("Test failed, expected empty value, but got '%s'", val)
	}

	// Alibaba Cloud
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_ali"), map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}
	val = GetCSP()
	if val != "alibaba" {
		t.Errorf("Test failed, expected 'alibaba', but got '%s'", val)
	}
	allManufacturerProviders = [...]manufacturerProviders{
		{noCloud, map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
		{noCloud, map[*regexp.Regexp]string{isAzure: CSPAzure, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
		{noCloud, map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
		{noCloud, map[*regexp.Regexp]string{isAWS: CSPAWS}},
	}

	// dmiDir does not exist
	dmiDir = "/path/does/not/exist"
	val = GetCSP()
	if val != "" {
		t.Errorf("Test failed, expected empty, but got '%s'", val)
	}

	// restore original environment
	dmiDir = "/sys/class/dmi"
	allManufacturerProviders = [...]manufacturerProviders{
		// dmidecode key files
		// /usr/sbin/dmidecode -s chassis-asset-tag
		{"/sys/class/dmi/id/chassis_asset_tag", map[*regexp.Regexp]string{isAzureCat: CSPAzure}},
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
	}
}
