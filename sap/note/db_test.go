package note

import (
	"os"
	"path"
	"testing"
)

func TestLinuxPagingImprovements(t *testing.T) {

	PCTestBaseConf := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/saptune/notes/1557506")
	if _, err := os.Stat(PCTestBaseConf); os.IsNotExist(err) {
		t.Skipf("file %s not available", PCTestBaseConf)
	}
	prepare := LinuxPagingImprovements{PagingConfig: PCTestBaseConf}
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
