package system

import (
	"os"
	"path"
	"testing"
)

func TestGetCSP(t *testing.T) {
	// amazon-web-services
	dmiSystemVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val := GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiSystemVersion = "/sys/devices/virtual/dmi/id/system_version"
	dmiBiosVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiBiosVendor = "/sys/devices/virtual/dmi/id/bios_vendor"
	dmiBiosVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiBiosVersion = "/sys/devices/virtual/dmi/id/bios_version"
	dmiBoardVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiBoardVendor = "/sys/devices/virtual/dmi/id/board_vendor"

	// GoogleCloud
	dmiBiosVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google")
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	dmiBiosVendor = "/sys/devices/virtual/dmi/id/bios_vendor"
	dmiBiosVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google")
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	dmiBiosVersion = "/sys/devices/virtual/dmi/id/bios_version"
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google")
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	dmiSystemManufacturer = "/sys/devices/virtual/dmi/id/system-manufacturer"

	// OracleCloud
	dmiBiosVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_ovm")
	val = GetCSP()
	if val != "ovm" {
		t.Errorf("Test failed, expected 'ovm', but got '%s'", val)
	}
	dmiBiosVersion = "/sys/devices/virtual/dmi/id/bios_version"

	// microsoft-azure
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure")
	val = GetCSP()
	if val != "azure" {
		t.Errorf("Test failed, expected 'azure', but got '%s'", val)
	}
	dmiSystemManufacturer = "/sys/devices/virtual/dmi/id/system-manufacturer"
	dmiChassisAssetTag = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure_cat")
	val = GetCSP()
	if val != "azure" {
		t.Errorf("Test failed, expected 'azure', but got '%s'", val)
	}
	dmiChassisAssetTag = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure_false")
	val = GetCSP()
	if val != "" {
		t.Errorf("Test failed, expected empty value, but got '%s'", val)
	}
	dmiChassisAssetTag = "/sys/devices/virtual/dmi/id/chassis_asset_tag"

	// Alibaba Cloud
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_ali")
	val = GetCSP()
	if val != "alibaba" {
		t.Errorf("Test failed, expected 'alibaba', but got '%s'", val)
	}
	dmiSystemManufacturer = "/sys/devices/virtual/dmi/id/system-manufacturer"
}
