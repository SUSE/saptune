package system

import (
	"os"
	"path"
	"testing"
)

func TestGetCSP(t *testing.T) {
	// microsoft-azure
	dmiSystemManufacturer = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/csp_azure")
	val := GetCSP()
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
}
