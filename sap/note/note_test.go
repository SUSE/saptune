package note

import (
	"encoding/json"
	"os"
	"path"
	"testing"
)

var OSNotesInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/usr/share/notes")
var OSPackageInGOPATH = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/ospackage/")

func jsonMarshalAndBack(original interface{}, receiver interface{}, t *testing.T) {
	serialised, err := json.Marshal(original)
	if err != nil {
		t.Fatal(original, err)
	}
	json.Unmarshal(serialised, &receiver)
}

func TestNoteSerialisation(t *testing.T) {
	// All notes must be tested
	paging := LinuxPagingImprovements{VMPagecacheLimitMB: 1000, VMPagecacheLimitIgnoreDirty: 2, UseAlgorithmForHANA: true}
	newPaging := LinuxPagingImprovements{}
	jsonMarshalAndBack(paging, &newPaging, t)
	if eq, diff, valapply := CompareNoteFields(paging, newPaging); !eq {
		t.Fatal(diff, valapply)
	}

	sysctl := INISettings{ConfFilePath: path.Join(OSNotesInGOPATH, "1410736"), ID: "1410736", DescriptiveName: "", SysctlParams: map[string]string{"net.ipv4.tcp_keepalive_time": "300", "net.ipv4.tcp_keepalive_intvl": "75", "reminder": ""}, ValuesToApply: map[string]string{"": ""}}
	newSysctl := INISettings{}
	jsonMarshalAndBack(sysctl, &newSysctl, t)
	if eq, diff, valapply := CompareNoteFields(sysctl, newSysctl); !eq {
		t.Fatal(diff, valapply)
	}
}

func TestGetTuningOptions(t *testing.T) {
	allOpts := GetTuningOptions(OSNotesInGOPATH, "")
	if sorted := allOpts.GetSortedIDs(); len(allOpts) != len(sorted) {
		t.Fatal(sorted, allOpts)
	}
}
