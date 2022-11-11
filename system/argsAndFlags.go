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
// possible Flags - force, dryrun, help, version, show-non-compliant, format,
// colorscheme, non-compliance-check
// on command line - --force, --dry-run or --dryrun, --help, --version, --color-scheme, --format
// Some Flags (like 'format') can have a value (--format=json or --format=csv)
func ParseCliArgs() ([]string, map[string]string) {
	stArgs := []string{os.Args[0]}
	// supported flags
	stFlags := map[string]string{"force": "false", "dryrun": "false", "help": "false", "version": "false", "show-non-compliant": "false", "format": "", "colorscheme": "", "non-compliance-check": "false", "notSupported": ""}
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
	case "--non-compliance-check", "-non-compliance-check":
		flags["non-compliance-check"] = "true"
	default:
		setUnsupportedFlag(arg, flags)
	}
}

// setUnsupportedFlag sets or appends a value to the unsupported Flag
// collection of unsupported Flags
func setUnsupportedFlag(val string, flags map[string]string) {
	if flags["notSupported"] == "" {
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
	cmdLinePos := map[string]int{"globOpt": 1, "realm": 1, "realmOpt": 2, "cmd": 2, "cmdOpt": 3}

	// check some universal syntax and the global option
	// manipulating cmdLinePos values
	if !chkGlobalSyntax(cmdLinePos) {
		return false
	}
	// check for realm options
	// manipulating cmdLinePos values
	if !chkRealmOpts(cmdLinePos) {
		return false
	}
	// check for command options
	if !chkCmdOpts(cmdLinePos) {
		return false
	}
	return ret
}

// chkGlobalSyntax checks some universal syntax and the global options
func chkGlobalSyntax(cmdLinePos map[string]int) bool {
	ret := true
	// check minimum of arguments without flags (saptune + realm)
	if (IsFlagSet("version") || IsFlagSet("help")) && len(saptArgs) > 1 {
		// too many arguments
		ret = false
	}
	if !IsFlagSet("version") && !IsFlagSet("help") && len(saptArgs) < 2 {
		// too few arguments
		ret = false
	}
	if IsFlagSet("notSupported") {
		// unknown flag in command line found
		ret = false
	}
	if IsFlagSet("force") && IsFlagSet("dryrun") {
		// both together are not supported
		// even that this case is handled correctly
		ret = false
	}
	if IsFlagSet("colorscheme") && IsFlagSet("show-non-compliant") && len(os.Args) < cmdLinePos["cmdOpt"]+1 {
		// too few options for both flags set
		ret = false
	}
	if ret {
		// check global option
		ret = chkGlobalOpts(cmdLinePos)
	}
	return ret
}

// chkGlobalOpts checks for global options
// saptune -format=FORMAT [--version|--help]
// saptune --version or saptune --help
func chkGlobalOpts(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	globOpt := false
	globPos := 1
	if IsFlagSet("version") && IsFlagSet("help") {
		// both together are not supported
		ret = false
	}
	if IsFlagSet("format") {
		if !strings.Contains(stArgs[1], "--format") {
			ret = false
		} else {
			globPos++
			cmdLinePos["realm"] = cmdLinePos["realm"] + 1
			cmdLinePos["realmOpt"] = cmdLinePos["realmOpt"] + 1
			cmdLinePos["cmd"] = cmdLinePos["cmd"] + 1
			cmdLinePos["cmdOpt"] = cmdLinePos["cmdOpt"] + 1
		}
	}
	if IsFlagSet("version") {
		// support '--version' and '-version'
		if !strings.Contains(stArgs[globPos], "-version") {
			ret = false
		} else {
			globOpt = true
		}
	}
	if IsFlagSet("help") {
		// support '--help' and '-help'
		if !strings.Contains(stArgs[globPos], "-help") {
			ret = false
		} else {
			globOpt = true
		}
	}
	if globOpt {
		cmdLinePos["realm"] = cmdLinePos["realm"] + 1
		cmdLinePos["realmOpt"] = cmdLinePos["realmOpt"] + 1
		cmdLinePos["cmd"] = cmdLinePos["cmd"] + 1
		cmdLinePos["cmdOpt"] = cmdLinePos["cmdOpt"] + 1
	}
	return ret
}

// chkRealmOpts checks for realm options
// at the moment only 'saptune status' has an option (--non-compliance-check)
func chkRealmOpts(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	if IsFlagSet("non-compliance-check") {
		// check for valid realm
		if !(stArgs[cmdLinePos["realm"]] == "status" || stArgs[cmdLinePos["realm"]] == "service" || stArgs[cmdLinePos["realm"]] == "daemon") {
			ret = false
		}
		if stArgs[cmdLinePos["realm"]] == "status" {
			// realm option set
			// check minimum of values items in cmd line
			// (saptune + realm + option)
			if len(stArgs) < cmdLinePos["realmOpt"]+1 {
				// too few arguments
				ret = false
			} else if stArgs[cmdLinePos["realmOpt"]] != "--non-compliance-check" {
				ret = false
			} else {
				cmdLinePos["cmd"] = cmdLinePos["cmd"] + 1
				cmdLinePos["cmdOpt"] = cmdLinePos["cmdOpt"] + 1
			}
		}
	}
	return ret
}

// chkCmdOpts checks for command options
func chkCmdOpts(cmdLinePos map[string]int) bool {
	ret := true
	// check minimum of arguments for command options
	// saptune realm cmd
	if len(saptArgs) < 3 && (IsFlagSet("force") || IsFlagSet("dryrun") || IsFlagSet("colorscheme") || IsFlagSet("show-non-compliant")) {
		// too few arguments for the active flags
		return false
	}
	if len(os.Args) < cmdLinePos["cmdOpt"]+1 || (!IsFlagSet("force") && !IsFlagSet("dryrun") && !IsFlagSet("colorscheme") && !IsFlagSet("show-non-compliant") && !IsFlagSet("non-compliance-check")) {
		// no command options set or too few options
		// and/or non of the flags set, which need further checks
		// so let the 'old' default checks (in main and/or actions) set
		// the appropriate result
		return true
	}
	// saptune solution change [--force] SOLUTIONNAME
	// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
	if !chkForceFlag(cmdLinePos) {
		ret = false
	}
	// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
	if !chkDryrunFlag(cmdLinePos) {
		ret = false
	}
	// saptune note verify [--colorscheme=<color scheme>] [--show-non-compliant] [NOTEID]
	if !chkNoteVerifySyntax(cmdLinePos) {
		ret = false
	}
	// saptune (service) status  [--non-compliance-check]
	if !chkServiceStatusSyntax(cmdLinePos) {
		ret = false
	}
	return ret
}

// chkForceFlag checks the syntax of 'saptune solution change'
// and 'saptune staging release' command line regarding the use
// of the 'force' flag
// saptune solution change [--force] SOLUTIONNAME
// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
func chkForceFlag(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	if IsFlagSet("force") {
		if !(stArgs[cmdLinePos["realm"]] == "solution" && stArgs[cmdLinePos["cmd"]] == "change") && !(stArgs[cmdLinePos["realm"]] == "staging" && stArgs[cmdLinePos["cmd"]] == "release") {
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--force" {
			ret = false
		}
	}
	return ret
}

// chkDryrunFlag checks the syntax of 'saptune staging release'
// command line regarding command line the use of the 'dry-run' flag
// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
func chkDryrunFlag(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	if IsFlagSet("dryrun") {
		if !(stArgs[cmdLinePos["realm"]] == "staging" && stArgs[cmdLinePos["cmd"]] == "release") {
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--dry-run" {
			ret = false
		}
	}
	return ret
}

// chkNoteVerifySyntax checks the syntax of 'saptune note verify' command line
// regarding command line options
// saptune note verify [--colorscheme=<color scheme>] [--show-non-compliant] [NOTEID]
func chkNoteVerifySyntax(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	if IsFlagSet("colorscheme") || IsFlagSet("show-non-compliant") {
		if !(stArgs[cmdLinePos["realm"]] == "note" && stArgs[cmdLinePos["cmd"]] == "verify") {
			ret = false
		}
	}
	if IsFlagSet("colorscheme") && IsFlagSet("show-non-compliant") {
		// both flags set, check order
		if !(strings.HasPrefix(stArgs[cmdLinePos["cmdOpt"]], "--colorscheme=") && stArgs[cmdLinePos["cmdOpt"]+1] == "--show-non-compliant") {
			ret = false
		}
	} else if IsFlagSet("colorscheme") && !strings.HasPrefix(stArgs[cmdLinePos["cmdOpt"]], "--colorscheme=") {
		// flag at wrong place in arg list
		ret = false
	} else if IsFlagSet("show-non-compliant") && stArgs[cmdLinePos["cmdOpt"]] != "--show-non-compliant" {
		// flag at wrong place in arg list
		ret = false
	}
	return ret
}

// chkServiceStatusSyntax checks the syntax of 'saptune service status' or
// 'saptune daemon status' command line regarding command line options
// saptune service status [--non-compliance-check]
// saptune status [--non-compliance-check] is checked earlier (as 'realm')
func chkServiceStatusSyntax(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	if IsFlagSet("non-compliance-check") {
		// saptune service status --non-compliance-check
		// saptune daemon status --non-compliance-check
		if !(stArgs[cmdLinePos["realm"]] == "service" && stArgs[cmdLinePos["cmd"]] == "status") && !(stArgs[cmdLinePos["realm"]] == "daemon" && stArgs[cmdLinePos["cmd"]] == "status") {
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--non-compliance-check" {
			ret = false
		}
	}
	return ret
}
