package system

import (
	"os"
	"path"
	"testing"
)

var procCmdline1 = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/cmdline1")
var procCmdline2 = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/cmdline2")

func TestParseCmdline(t *testing.T) {
	actualVal := ParseCmdline(procCmdline1, "intel_idle.max_cstate")
	if actualVal != "NA" {
		t.Fatalf("intel_idle.max_cstate is set to '%s', but shouldn't\n", actualVal)
	}
	actualVal = ParseCmdline(procCmdline1, "processor.max_cstate")
	if actualVal != "NA" {
		t.Fatalf("processor.max_cstate is set to '%s', but shouldn't\n", actualVal)
	}
	actualVal = ParseCmdline(procCmdline1, "numa_balancing")
	if actualVal != "NA" {
		t.Fatalf("numa_balancing is set to '%s', but shouldn't\n", actualVal)
	}

	actualVal = ParseCmdline(procCmdline2, "intel_idle.max_cstate")
	if actualVal != "1" {
		t.Fatalf("intel_idle.max_cstate is not set to 1, but '%s'\n", actualVal)
	}
	actualVal = ParseCmdline(procCmdline2, "processor.max_cstate")
	if actualVal != "1" {
		t.Fatalf("processor.max_cstate is not set to 1, but '%s'\n", actualVal)
	}
	actualVal = ParseCmdline(procCmdline2, "numa_balancing")
	if actualVal != "NA" {
		t.Fatalf("numa_balancing is set to '%s', but shouldn't\n", actualVal)
	}
	actualVal = ParseCmdline(procCmdline2, "showopts")
	if actualVal != "showopts" {
		t.Fatalf("showopts is set, but '%s' is returned\n", actualVal)
	}

	actualVal = ParseCmdline("/saptune_file_not_avail", "intel_idle.max_cstate")
	if actualVal != "NA" {
		t.Fatalf("File '/saptune_file_not_avail' should not be available, so return should 'NA', but is '%s'\n", actualVal)
	}
}
