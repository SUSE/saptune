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
	os.Args = []string{"saptune", "note", "list", "--format", "json", "--force", "--dryrun", "--help", "--version", "--colorscheme", "full-green-zebra", "--show-non-compliant", "--wrongflag", "--unknownflag=none"}
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
	if !IsFlagSet("colorscheme") {
		t.Errorf("Test failed, expected 'colorscheme' flag as 'true', but got 'false'")
	}
	if !IsFlagSet("show-non-compliant") {
		t.Errorf("Test failed, expected 'show-non-compliant' flag as 'true', but got 'false'")
	}
	if !IsFlagSet("notSupported") {
		t.Errorf("Test failed, expected 'notSupported' flag as 'true', but got 'false'")
	}

	expected := "json"
	actual := GetFlagVal("format")
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}

	expected = ""
	actual = GetFlagVal("unknownflag")
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

func TestRereadArgs(t *testing.T) {
	os.Args = []string{"saptune", "note", "list"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	os.Args = []string{"saptune", "--format", "json", "solution", "enabled"}
	RereadArgs()

	expected := "solution"
	actual := CliArg(1)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	expected = "enabled"
	actual = CliArg(2)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	expected = ""
	actual = CliArg(4)
	if actual != expected {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expected, actual)
	}
	expectedSlice := []string{"solution", "enabled"}
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

	if IsFlagSet("version") {
		t.Errorf("Test failed, expected 'version' flag as 'false', but got 'true'")
	}
	if !IsFlagSet("format") {
		t.Errorf("Test failed, expected 'format' flag as 'true', but got 'false'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
}

func TestChkCliSyntax(t *testing.T) {
	// {"saptune", "note", "list", "--format json"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--format", "json"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "--format hugo", "note", "list"} -> wrong
	os.Args = []string{"saptune", "--format", "hugo", "note", "list"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// to few arguments
	// {"saptune", "staging"} -> ok
	// the check of this situation will be postponed to our 'old' default
	// checks (in 'main' and/or 'actions'
	os.Args = []string{"saptune", "staging"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// line with unknown flag
	// {"saptune", "--unknown", "note", "list"} -> wrong
	os.Args = []string{"saptune", "--unknown", "note", "list"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "note", "list", "--unknown"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--unknown"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "--out=json", "note", "list"} -> wrong
	os.Args = []string{"saptune", "--out=json", "note", "list"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "--format json", "note", "list"} -> ok
	os.Args = []string{"saptune", "--format", "json", "note", "list"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "note", "list", "--dry-run"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--dry-run"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "note", "list", "--force"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--force"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
	// {"saptune", "staging", "list", "--force"} -> wrong
	os.Args = []string{"saptune", "staging", "list", "--force"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "staging", "release"} -> ok
	os.Args = []string{"saptune", "staging", "release"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "staging", "release", "--force"} -> ok
	os.Args = []string{"saptune", "staging", "release", "--force"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "staging", "release", "--dry-run"} -> ok
	os.Args = []string{"saptune", "staging", "release", "--dry-run"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// line with force AND dry-run
	// {"saptune", "staging", "release", "--force", "--dry-run"} -> wrong
	os.Args = []string{"saptune", "staging", "release", "--force", "--dry-run"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "staging", "release", "--show-non-compliant"} -> wrong
	os.Args = []string{"saptune", "staging", "release", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "staging", "release", "--force", "--show-non-compliant"} -> wrong
	os.Args = []string{"saptune", "staging", "release", "--force", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "--force", "staging", "release"} -> wrong
	os.Args = []string{"saptune", "--force", "staging", "release"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "--force", "staging", "release", "--dry-run"} -> wrong
	os.Args = []string{"saptune", "--force", "staging", "release", "--dry-run"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "staging", "release", "--hugo", "--dry-run"} -> wrong
	os.Args = []string{"saptune", "--force", "staging", "release", "--hugo", "--dry-run"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// saptune note verify [--colorscheme <color scheme>] [--show-non-compliant] [NOTEID]
	// {"saptune", "note", "list", "--colorscheme full-green-zebra"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--colorscheme", "full-green-zebra"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "note", "list", "--show-non-compliant"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "note", "list", "--colorscheme full-green-zebra", "--show-non-compliant"} -> wrong
	os.Args = []string{"saptune", "note", "list", "--colorscheme", "full-green-zebra", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "staging", "list", "--colorscheme full-green-zebra"} -> wrong
	os.Args = []string{"saptune", "staging", "list", "--colorscheme", "full-green-zebra"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "note", "verify"} -> ok
	os.Args = []string{"saptune", "note", "verify"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "note", "verify", "--colorscheme full-green-zebra"} -> ok
	os.Args = []string{"saptune", "note", "verify", "--colorscheme", "full-green-zebra"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "note", "verify", "--colorscheme full-green-zebra", "--show-non-compliant"} -> ok
	os.Args = []string{"saptune", "note", "verify", "--colorscheme", "full-green-zebra", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "note", "verify", "--show-non-compliant"} -> ok
	os.Args = []string{"saptune", "note", "verify", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "note", "verify", "--colorscheme zebra", "--show-non-compliant"} -> ok
	os.Args = []string{"saptune", "note", "verify", "--colorscheme", "zebra", "--show-non-compliant"}
	saptArgs, saptFlags = ParseCliArgs()
	if !ChkCliSyntax() {
		t.Errorf("Test failed, expected good syntax, but got 'wrong'")
	}

	// {"saptune", "note", "verify", "--show-non-compliant", "--colorscheme full-green-zebra"} -> wrong
	os.Args = []string{"saptune", "note", "verify", "--show-non-compliant", "--colorscheme", "full-green-zebra"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "--show-non-compliant", "note", "verify"} -> wrong
	os.Args = []string{"saptune", "--show-non-compliant", "note", "verify"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// {"saptune", "note", "--colorscheme full-green-zebra", "verify"} -> wrong
	os.Args = []string{"saptune", "note", "--colorscheme", "full-green-zebra", "verify"}
	saptArgs, saptFlags = ParseCliArgs()
	if ChkCliSyntax() {
		t.Errorf("Test failed, expected wrong syntax, but got 'good'")
	}

	// reset CLI flags and args
	saptArgs = []string{}
	saptFlags = map[string]string{}
}
