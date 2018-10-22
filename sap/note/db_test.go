package note

import (
	//"github.com/SUSE/saptune/system"
	"os"
	"path"
	"testing"
)

func TestLinuxPagingImprovements(t *testing.T) {
	if _, err := os.Stat(path.Join(OSPackageInGOPATH, "/etc/sysconfig/saptune-note-1557506")); os.IsNotExist(err) {
		t.Skip("file etc/sysconfig/saptune-note-1557506 not available")
	}
	prepare := LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH}
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
	// As written in OSPackageInGOPATH, paging improvements are not to be enabled by default, hence it should not change anything
	o := optimised.(LinuxPagingImprovements)
	if o.VMPagecacheLimitMB != 0 || o.VMPagecacheLimitIgnoreDirty != 1 {
		t.Fatal(o)
	}
}
