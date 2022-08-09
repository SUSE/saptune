package system

import (
	"os"
	"path"
	"testing"
)

func TestGetCSP(t *testing.T) {
	// initialize test environment
	noCloud := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_no_cloud")
	dmiChassisAssetTag = noCloud
	dmiBoardVendor = noCloud
	dmiBiosVendor = noCloud
	dmiBiosVersion = noCloud
	dmiSystemVersion = noCloud
	dmiSystemManufacturer = noCloud

	// amazon-web-services
	dmiSystemVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val := GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiSystemVersion = noCloud
	dmiBiosVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiBiosVendor = noCloud
	dmiBiosVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiBiosVersion = noCloud
	dmiBoardVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiBoardVendor = noCloud
	dmiSysVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_aws")
	val = GetCSP()
	if val != "aws" {
		t.Errorf("Test failed, expected 'aws', but got '%s'", val)
	}
	dmiSysVendor = noCloud

	// GoogleCloud
	dmiBiosVendor = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google")
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	dmiBiosVendor = noCloud
	dmiBiosVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google")
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	dmiBiosVersion = noCloud
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_google")
	val = GetCSP()
	if val != "google" {
		t.Errorf("Test failed, expected 'google', but got '%s'", val)
	}
	dmiSystemManufacturer = noCloud

	// OracleCloud
	dmiBiosVersion = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_ovm")
	val = GetCSP()
	if val != "ovm" {
		t.Errorf("Test failed, expected 'ovm', but got '%s'", val)
	}
	dmiBiosVersion = noCloud

	// microsoft-azure
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure")
	val = GetCSP()
	if val != "azure" {
		t.Errorf("Test failed, expected 'azure', but got '%s'", val)
	}
	dmiSystemManufacturer = noCloud
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
	dmiChassisAssetTag = noCloud

	// Alibaba Cloud
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_ali")
	val = GetCSP()
	if val != "alibaba" {
		t.Errorf("Test failed, expected 'alibaba', but got '%s'", val)
	}
	dmiSystemManufacturer = noCloud

	// dmiDir does not exist
	dmiDir = "/path/does/not/exist"
	val = GetCSP()
	if val != "" {
		t.Errorf("Test failed, expected empty, but got '%s'", val)
	}

	// restore original environment
	dmiDir = "/sys/class/dmi"
	dmiChassisAssetTag = "/sys/class/dmi/id/chassis_asset_tag"
	dmiBoardVendor = "/sys/class/dmi/id/board_vendor"
	dmiBiosVendor = "/sys/class/dmi/id/bios_vendor"
	dmiBiosVersion = "/sys/class/dmi/id/bios_version"
	dmiSystemVersion = "/sys/class/dmi/id/system_version"
	dmiSystemManufacturer = "/sys/class/dmi/id/system-manufacturer"
	dmiSysVendor = "/sys/class/dmi/id/sys_vendor"
}
