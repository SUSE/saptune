package note

import (
	"encoding/json"
	"github.com/HouzuoGuo/saptune/sap/param"
	"os"
	"path"
	"testing"
)

var SYSCONFIG_SRC_DIR = path.Join(os.Getenv("GOPATH"), "/src/github.com/HouzuoGuo/saptune/ospackage/")

func jsonMarshalAndBack(original interface{}, receiver interface{}, t *testing.T) {
	serialised, err := json.Marshal(original)
	if err != nil {
		t.Fatal(original, err)
	}
	json.Unmarshal(serialised, &receiver)
}

func TestNoteSerialisation(t *testing.T) {
	// All notes must be tested
	hana := HANARecommendedOSSettings{
		KernelMMTransparentHugepage: "always",
		KernelMMKsm:                 true,
		KernelNumaBalancing:         true,
	}
	newHANA := HANARecommendedOSSettings{}
	jsonMarshalAndBack(hana, &newHANA, t)
	if eq, diff := CompareNoteFields(hana, newHANA); !eq {
		t.Fatal(diff)
	}

	paging := LinuxPagingImprovements{VMPagecacheLimitMB: 1000, VMPagecacheLimitIgnoreDirty: 2, UseAlgorithmForHANA: true}
	newPaging := LinuxPagingImprovements{}
	jsonMarshalAndBack(paging, &newPaging, t)
	if eq, diff := CompareNoteFields(paging, newPaging); !eq {
		t.Fatal(diff)
	}

	prepare := PrepareForSAPEnvironments{
		ShmFileSystemSizeMB: 1000,
		KernelShmMax:        1001, KernelShmAll: 1002, KernelShmMni: 1003,
		KernelSemMsl: 1004, KernelSemMns: 1005, KernelSemOpm: 1006, KernelSemMni: 1007,
		LimitNofileSapsysSoft: 1,
		LimitNofileSapsysHard: 2,
		LimitNofileSdbaSoft:   3,
		LimitNofileSdbaHard:   4,
		LimitNofileDbaSoft:    5,
		LimitNofileDbaHard:    6,
	}
	newPrepare := PrepareForSAPEnvironments{}
	jsonMarshalAndBack(prepare, &newPrepare, t)
	if eq, diff := CompareNoteFields(prepare, newPrepare); !eq {
		t.Fatal(diff)
	}

	afterInst := AfterInstallation{UuiddSocket: true}
	newAfterInst := AfterInstallation{}
	jsonMarshalAndBack(afterInst, &newAfterInst, t)
	if eq, diff := CompareNoteFields(afterInst, newAfterInst); !eq {
		t.Fatal(diff)
	}

	ioel := VmwareGuestIOElevator{
		BlockDeviceSchedulers: param.BlockDeviceSchedulers{
			SchedulerChoice: map[string]string{"sda": "noop", "sdb": "deadline", "sdc": "cfq"},
		},
	}
	newIoel := VmwareGuestIOElevator{}
	jsonMarshalAndBack(ioel, &newIoel, t)
	if eq, diff := CompareNoteFields(ioel, newIoel); !eq {
		t.Fatal(diff)
	}

	systune1 := SUSESysOptimisation{
		SysconfigPrefix:        "abc",
		VMNumberHugePages:      1,
		VMSwappiness:           2,
		VMVfsCachePressure:     3,
		VMOvercommitMemory:     4,
		VMOvercommitRatio:      5,
		VMDirtyRatio:           6,
		VMDirtyBackgroundRatio: 7,
		BlockDeviceSchedulers: param.BlockDeviceSchedulers{
			SchedulerChoice: map[string]string{"sda": "noop", "sdb": "deadline", "sdc": "cfq"},
		},
	}
	newSystune1 := SUSESysOptimisation{}
	jsonMarshalAndBack(systune1, &newSystune1, t)
	if eq, diff := CompareNoteFields(systune1, newSystune1); !eq {
		t.Fatal(diff)
	}

	systune2 := SUSENetCPUOptimisation{
		SysconfigPrefix: "a",
		NetCoreWmemMax:  1, NetCoreRmemMax: 2,
		NetCoreNetdevMaxBacklog: 3, NetCoreSoMaxConn: 4,
		NetIpv4TcpRmem: 5, NetIpv4TcpWmem: 6,
		NetIpv4TcpTimestamps: 7, NetIpv4TcpSack: 8, NetIpv4TcpFack: 9, NetIpv4TcpDsack: 10,
		NetIpv4IpfragLowThres: 11, NetIpv4IpfragHighThres: 12,
		NetIpv4TcpMaxSynBacklog: 14, NetIpv4TcpSynackRetries: 15, NetIpv4TcpRetries2: 16,
		NetTcpKeepaliveTime: 17, NetTcpKeepaliveProbes: 18, NetTcpKeepaliveIntvl: 19,
		NetTcpTwRecycle: 20, NetTcpTwReuse: 21, NetTcpFinTimeout: 22,
		NetTcpMtuProbing:     23,
		NetIpv4TcpSyncookies: 24, NetIpv4ConfAllAcceptSourceRoute: 25,
		NetIpv4ConfAllAcceptRedirects: 36, NetIpv4ConfAllRPFilter: 27,
		NetIpv4IcmpEchoIgnoreBroadcasts: 28, NetIpv4IcmpIgnoreBogusErrorResponses: 29, NetIpv4ConfAllLogMartians: 30,
		KernelRandomizeVASpace: 31, KernelKptrRestrict: 32, FSProtectedHardlinks: 33, FSProtectedSymlinks: 34,
		KernelSchedChildRunsFirst: 35,
	}
	newSystune2 := SUSENetCPUOptimisation{}
	jsonMarshalAndBack(systune2, &newSystune2, t)
	if eq, diff := CompareNoteFields(systune2, newSystune2); !eq {
		t.Fatal(diff)
	}
}

func TestGetTuningOptions(t *testing.T) {
	allOpts := GetTuningOptions("")
	if sorted := allOpts.GetSortedIDs(); len(allOpts) != len(sorted) {
		t.Fatal(sorted, allOpts)
	}
}

func TestCompareNoteFields(t *testing.T) {
	// SUSESysOptimisation has a good mix of data types among its fields, hence it is chosen for this test.
	systune := SUSESysOptimisation{
		SysconfigPrefix:        "abc",
		VMNumberHugePages:      1,
		VMSwappiness:           2,
		VMVfsCachePressure:     3,
		VMOvercommitMemory:     4,
		VMOvercommitRatio:      5,
		VMDirtyRatio:           6,
		VMDirtyBackgroundRatio: 7,
		BlockDeviceSchedulers: param.BlockDeviceSchedulers{
			SchedulerChoice: map[string]string{"sda": "noop", "sdb": "deadline", "sdc": "cfq"},
		},
	}
	allMatch, comparisons := CompareNoteFields(systune, systune)
	if !allMatch || len(comparisons) != 9 {
		t.Fatal(allMatch, comparisons)
	}
	for _, comparison := range comparisons {
		if comparison.ReflectFieldName == "" || comparison.ActualValueJS == "" || comparison.ExpectedValueJS == "" || !comparison.MatchExpectation {
			t.Fatal(comparison)
		}
	}
	// Make three mismatches in string, integer, and map
	newSystune := SUSESysOptimisation{
		SysconfigPrefix:        "MISMATCH",
		VMNumberHugePages:      99999999,
		VMSwappiness:           2,
		VMVfsCachePressure:     3,
		VMOvercommitMemory:     4,
		VMOvercommitRatio:      5,
		VMDirtyRatio:           6,
		VMDirtyBackgroundRatio: 7,
		BlockDeviceSchedulers: param.BlockDeviceSchedulers{
			SchedulerChoice: map[string]string{"MISMATCH": "MISMATCH"},
		},
	}
	allMatch, comparisons = CompareNoteFields(newSystune, systune)
	if allMatch || len(comparisons) != 9 {
		t.Fatal(allMatch, comparisons)
	}
	for _, comparison := range comparisons {
		switch comparison.ReflectFieldName {
		case "SysconfigPrefix":
			if comparison.ExpectedValueJS != `abc` || comparison.ActualValueJS != `MISMATCH` || comparison.MatchExpectation {
				t.Fatalf("%+v", comparison)
			}
		case "VMNumberHugePages":
			if comparison.ExpectedValueJS != "1" || comparison.ActualValueJS != "99999999" || comparison.MatchExpectation {
				t.Fatalf("%+v", comparison)
			}
		case "BlockDeviceSchedulers":
			if comparison.ExpectedValueJS != `{"SchedulerChoice":{"sda":"noop","sdb":"deadline","sdc":"cfq"}}` ||
				comparison.ActualValueJS != `{"SchedulerChoice":{"MISMATCH":"MISMATCH"}}` || comparison.MatchExpectation {
				t.Fatalf("%+v", comparison)
			}
		default:
			// Other fields should match
			if comparison.ReflectFieldName == "" || comparison.ActualValueJS == "" || comparison.ExpectedValueJS == "" || !comparison.MatchExpectation {
				t.Fatal(comparison)
			}
		}

	}
}
