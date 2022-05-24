package system

import (
	"os"
	"regexp"
	"strings"
)

// get saptune arguments and flags
var saptArgs, saptFlags = ParseCliArgs()

// RereadArgs parses the cli parameter again
func RereadArgs() {
	saptArgs, saptFlags = ParseCliArgs()
}

// CliArg returns the i-th command line parameter,
// or empty string if it is not specified.
func CliArg(i int) string {
	if len(saptArgs) >= i+1 {
		return saptArgs[i]
	}
	return ""
}

// CliArgs returns all remaining command line parameters starting with i,
// or empty string if it is not specified.
func CliArgs(i int) []string {
	if len(saptArgs) >= i+1 {
		return saptArgs[i:]
	}
	return []string{}
}

// IsFlagSet returns true, if the flag is available on the command line
// or false, if not
func IsFlagSet(flag string) bool {
	if saptFlags[flag] == "false" || saptFlags[flag] == "" {
		return false
	}
	return true
}

// GetFlagVal returns the value of a saptune commandline flag
func GetFlagVal(flag string) string {
	return saptFlags[flag]
}

// ParseCliArgs parses the command line to identify special flags and the
// 'normal' arguments
// returns a map of Flags (set/not set or value) and a slice containing the
// remaining arguments
// possible Flags - force, dryrun, help, version, format, colorscheme
// on command line - --force, --dry-run or --dryrun, --help, --version, --color-scheme, --format
// Some Flags (like 'format') can have a value (--format=json or --format=csv)
func ParseCliArgs() ([]string, map[string]string) {
	stArgs := []string{os.Args[0]}
	// supported flags
	stFlags := map[string]string{"force": "false", "dryrun": "false", "help": "false", "version": "false", "show-non-compliant": "false", "format": "", "colorscheme": "", "notSupported": ""}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// argument is a flag
			handleFlags(arg, stFlags)
			continue
		}
		// other args
		stArgs = append(stArgs, arg)
	}
	return stArgs, stFlags
}

// handleFlags checks for valid flags in the CLI arg list
func handleFlags(arg string, flags map[string]string) {
	var valueFlag = regexp.MustCompile(`(-[\w-]+)=(.*)`)
	matches := valueFlag.FindStringSubmatch(arg)
	if len(matches) == 3 {
		// flag with value
		handleValueFlags(arg, matches, flags)
		return
	}
	handleSimpleFlags(arg, flags)
}

// handleValueFlags checks for valid flags with value in the CLI arg list
func handleValueFlags(arg string, matches []string, flags map[string]string) {
	if strings.Contains(arg, "--format") {
		// --format=json
		flags["format"] = matches[2]
	}
	if strings.Contains(arg, "-colorscheme") {
		// --colorscheme=zebra
		flags["colorscheme"] = matches[2]
	}
	if _, ok := flags[strings.TrimLeft(matches[1], "-")]; !ok {
		setUnsupportedFlag(matches[1], flags)
	}
}

// handleSimpleFlags checks for valid flags in the CLI arg list
func handleSimpleFlags(arg string, flags map[string]string) {
	// simple flags
	switch arg {
	case "--force", "-force":
		flags["force"] = "true"
	case "--dry-run", "-dry-run", "--dryrun", "-dryrun":
		flags["dryrun"] = "true"
	case "--help", "-help", "-h":
		flags["help"] = "true"
	case "--version", "-version":
		flags["version"] = "true"
	case "--show-non-compliant", "-show-non-compliant":
		flags["show-non-compliant"] = "true"
	default:
		setUnsupportedFlag(arg, flags)
	}
}

// setUnsupportedFlag sets or appends a value to the unsupported Flag
// collection of unsupported Flags
func setUnsupportedFlag(val string, flags map[string]string) {
	if flags["notSupported"] != "" {
		flags["notSupported"] = flags["notSupported"] + val
	} else {
		flags["notSupported"] = flags["notSupported"] + " " + val
	}
}

// ChkCliSyntax checks, if command line parameter are in the right order
// only checking the right position of the 'options' aka 'flags'
// saptune globOpt realm realmOpt cmd cmdOpt param
func ChkCliSyntax() bool {
	ret := true
	sArgs := os.Args
	globOpt := 1
	realm := 1
	realmOpt := 2 // currently not used
	cmd := 2
	cmdOpt := 3

	if IsFlagSet("format") {
		realm = realm + 1
		realmOpt = realmOpt + 1
		cmd = cmd + 1
		cmdOpt = cmdOpt + 1
	}
	// future - if realmOpt is set (one or more options)
	// we need to increase the cmd and cmdOpt accordingly

	// check some universal syntax and the global option
	if !chkGlobalSyntax(sArgs, globOpt, cmdOpt) {
		return false
	}

	// check realm option
	// not yet used, future option

	// check command options
	if len(sArgs) < cmdOpt || (!IsFlagSet("force") && !IsFlagSet("dryrun") && !IsFlagSet("colorscheme") && !IsFlagSet("show-non-compliant")) {
		// no command options set or  too few options
		return true
	}
	// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
	if !chkStagingReleaseSyntax(sArgs, realm, cmd, cmdOpt) {
		ret = false
	}
	// saptune note verify [--coloscheme=<color scheme>] [--show-non-compliant] [NOTEID]
	if !chkNoteVerifySyntax(sArgs, realm, cmd, cmdOpt) {
		ret = false
	}
	return ret
}

// chkGlobalSyntax checks some universal syntax and the global option
func chkGlobalSyntax(stArgs []string, pglob, pcopt int) bool {
	ret := true
	if IsFlagSet("notSupported") {
		// unknown flag in command line found
		ret = false
	}
	if IsFlagSet("force") && IsFlagSet("dryrun") {
		// both together are not supported
		// even that this case is handled correctly
		ret = false
	}
	if IsFlagSet("colorscheme") && IsFlagSet("show-non-compliant") && len(stArgs) < pcopt+1 {
		// too few options for both flags set
		ret = false
	}

	// check global option
	// saptune -out=FORMAT
	// && !strings.HasPrefix(sArgs[1], "-o="
	if IsFlagSet("format") && !strings.Contains(stArgs[pglob], "--format") {
		ret = false
	}
	return ret
}

// chkStagingReleaseSyntax checks the syntax of 'saptune staging release'
// command line regarding command line options
// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
func chkStagingReleaseSyntax(stArgs []string, prealm, pcmd, popt int) bool {
	ret := true
	if IsFlagSet("dryrun") || IsFlagSet("force") {
		if stArgs[prealm] != "staging" && stArgs[pcmd] != "release" {
			ret = false
		}
		if stArgs[popt] != "--dry-run" && stArgs[popt] != "--force" {
			ret = false
		}
	}
	return ret
}

// chkNoteVerifySyntax checks the syntax of 'saptune note verify' command line
// regarding command line options
// saptune note verify [--coloscheme=<color scheme>] [--show-non-compliant] [NOTEID]
func chkNoteVerifySyntax(stArgs []string, prealm, pcmd, popt int) bool {
	ret := true
	if IsFlagSet("colorscheme") {
		if stArgs[prealm] != "note" && stArgs[pcmd] != "verify" {
			ret = false
		}
		if IsFlagSet("show-non-compliant") {
			if !strings.HasPrefix(stArgs[popt], "--colorscheme=") && !strings.HasPrefix(stArgs[popt+1], "--colorscheme=") {
				ret = false
			}
		} else if !strings.HasPrefix(stArgs[popt], "--colorscheme=") {
			ret = false
		}
	}
	if IsFlagSet("show-non-compliant") {
		if stArgs[prealm] != "note" && stArgs[pcmd] != "verify" {
			ret = false
		}
		if IsFlagSet("colorscheme") {
			if stArgs[popt] != "--show-non-compliant" && stArgs[popt+1] != "--show-non-compliant" {
				ret = false
			}
		} else if stArgs[popt] != "--show-non-compliant" {
			ret = false
		}
	}
	return ret
}
