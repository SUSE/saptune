package system

import (
	"os"
	"testing"
)

func TestCliArg(t *testing.T) {
	os.Args = []string{"saptune", "note", "list"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()

	expected := "note"
	actual := CliArg(1)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	expected = "list"
	actual = CliArg(2)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	expected = ""
	actual = CliArg(4)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	expectedSlice := []string{"note", "list"}
	actualSlice := CliArgs(1)
	for i, arg := range actualSlice {
		if arg != expectedSlice[i] {
			t.Errorf("Test failed, expected: '%s', got: '%s'", expectedSlice[i], arg)
		}
	}
	expectedSlice = []string{}
	actualSlice = CliArgs(4)
	if len(actualSlice) != 0 {
		t.Errorf("Test failed, expected: '%v', got: '%v'", expectedSlice, actualSlice)
	}

	if IsFlagSet("force") {
		t.Errorf("Test failed, expected 'force' flag as 'false', but got 'true'")
	}
	if IsFlagSet("dryrun") {
		t.Errorf("Test failed, expected 'dryrun' flag as 'false', but got 'true'")
	}
	if IsFlagSet("help") {
		t.Errorf("Test failed, expected 'help' flag as 'false', but got 'true'")
	}
	if IsFlagSet("version") {
		t.Errorf("Test failed, expected 'version' flag as 'false', but got 'true'")
	}
	if IsFlagSet("format") {
		t.Errorf("Test failed, expected 'format' flag as 'false', but got 'true'")
	}
	expected = ""
	actual = GetFlagVal("format")
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	if IsFlagSet("notsupported") {
		t.Errorf("Test failed, expected 'notsupported' flag as 'false', but got 'true'")
	}
	if IsFlagSet("") {
		t.Errorf("Test failed, expected 'notsupported' flag as 'false', but got 'true'")
	}
	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
}

func TestCliFlags(t *testing.T) {
	os.Args = []string{"saptune", "note", "list", "--format=json", "--force", "--dryrun", "--help", "--version"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()

	if !IsFlagSet("force") {
		t.Errorf("Test failed, expected 'force' flag as 'true', but got 'false'")
	}
	if !IsFlagSet("dryrun") {
		t.Errorf("Test failed, expected 'dryrun' flag as 'true', but got 'false'")
	}
	if !IsFlagSet("help") {
		t.Errorf("Test failed, expected 'help' flag as 'true', but got 'false'")
	}
	if !IsFlagSet("version") {
		t.Errorf("Test failed, expected 'version' flag as 'true', but got 'false'")
	}
	if !IsFlagSet("format") {
		t.Errorf("Test failed, expected 'format' flag as 'true', but got 'false'")
	}
	expected := "json"
	actual := GetFlagVal("format")
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
	os.Args = []string{"saptune", "-force"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if !IsFlagSet("force") {
		t.Errorf("Test failed, expected 'force' flag as 'true', but got 'false'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
	os.Args = []string{"saptune", "-dry-run"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if !IsFlagSet("dryrun") {
		t.Errorf("Test failed, expected 'dryrun' flag as 'true', but got 'false'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
	os.Args = []string{"saptune", "-help"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if !IsFlagSet("help") {
		t.Errorf("Test failed, expected 'help' flag as 'true', but got 'false'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
	os.Args = []string{"saptune", "-h"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if !IsFlagSet("help") {
		t.Errorf("Test failed, expected 'help' flag as 'true', but got 'false'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
	os.Args = []string{"saptune", "-version"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if !IsFlagSet("version") {
		t.Errorf("Test failed, expected 'version' flag as 'true', but got 'false'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
}
