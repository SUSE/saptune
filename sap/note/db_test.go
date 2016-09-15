package note

import (
	"gitlab.suse.de/guohouzuo/saptune/system"
	"testing"
)

func TestHANARecommendedOSSettings(t *testing.T) {
	if system.IsUserOBS() {
		t.Skip("Transparent huge page settings cannot be read on build service")
	}
	prepare := HANARecommendedOSSettings{}
	if prepare.Name() == "" {
		t.Fatal(prepare.Name())
	}
	initPrepare, err := prepare.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	optimised, err := initPrepare.(HANARecommendedOSSettings).Optimise()
	if err != nil || optimised.(HANARecommendedOSSettings) == initPrepare.(HANARecommendedOSSettings) {
		t.Fatal(err, optimised, initPrepare)
	}
	// Check attributes from each optimised parameter
	o := optimised.(HANARecommendedOSSettings)
	if o.KernelMMTransparentHugepage != "never" || o.KernelNumaBalancing != false || o.KernelMMKsm != false {
		t.Fatal(o)
	}
}

func TestLinuxPagingImprovements(t *testing.T) {
	prepare := LinuxPagingImprovements{SysconfigPrefix: SYSCONFIG_SRC_DIR}
	if prepare.Name() == "" {
		t.Fatal(prepare.Name())
	}
	initPrepare, err := prepare.Initialise()
	if err != nil {
		t.Fatal(err)
	}
	optimised, err := initPrepare.(LinuxPagingImprovements).Optimise()
	if err != nil {
		t.Fatal(err)
	}
	// As written in SYSCONFIG_SRC_DIR, paging improvements are not to be enabled by default, hence it should not change anything
	o := optimised.(LinuxPagingImprovements)
	if o.VMPagecacheLimitMB != 0 || o.VMPagecacheLimitIgnoreDirty != 1 {
		t.Fatal(o)
	}
}
