package system

import (
	"os"
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
// Some Flags (like 'format') can have a value (--format json or --format csv)
func ParseCliArgs() ([]string, map[string]string) {
	stArgs := []string{}
	// supported flags
	stFlags := map[string]string{"force": "false", "dryrun": "false", "help": "false", "version": "false", "show-non-compliant": "false", "format": "", "colorscheme": "", "non-compliance-check": "false", "notSupported": ""}
	skip := false
	for i, arg := range os.Args {
		if skip {
			// skip this command line parameter, because it's the value
			// belonging to a flag
			skip = false
			continue
		}
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// argument is a flag
			// skip next command line parameter, if it is the value
			// belonging to a flag
			skip = handleFlags(arg, i, stFlags)
			continue
		}
		// other args
		stArgs = append(stArgs, arg)
	}
	return stArgs, stFlags
}

// handleFlags checks for valid flags in the CLI arg list
func handleFlags(arg string, idx int, flags map[string]string) bool {
	fval := idx + 1
	farg := "flag_value"
	if fval < len(os.Args) {
		farg = os.Args[fval]
	}

	skip := handleValueFlags(arg, farg, flags)
	if !skip {
		handleSimpleFlags(arg, flags)
	}
	return skip
}

// handleValueFlags checks for valid flags with value in the CLI arg list
func handleValueFlags(arg, farg string, flags map[string]string) bool {
	skip := false
	if strings.Contains(arg, "--format") {
		// --format json
		flags["format"] = farg
		skip = true
	}
	if strings.Contains(arg, "-colorscheme") {
		// --colorscheme zebra
		flags["colorscheme"] = farg
		skip = true
	}
	return skip
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
		DebugLog("ChkCliSyntax - chkGlobalSyntax failed")
		return false
	}
	// check for realm options
	// manipulating cmdLinePos values
	if !chkRealmOpts(cmdLinePos) {
		DebugLog("ChkCliSyntax - chkRealmOpts failed")
		return false
	}
	// check for command options
	if !chkCmdOpts(cmdLinePos) {
		DebugLog("ChkCliSyntax - chkCmdOpts failed")
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
		DebugLog("chkGlobalSyntax failed - too many arguments")
		ret = false
	}
	if !IsFlagSet("version") && !IsFlagSet("help") && len(saptArgs) < 2 {
		// too few arguments
		DebugLog("chkGlobalSyntax failed - too few arguments")
		ret = false
	}
	if IsFlagSet("notSupported") {
		// unknown flag in command line found
		DebugLog("chkGlobalSyntax failed - unknown flag '%+v' found in command line", GetFlagVal("notSupported"))
		ret = false
	}
	if IsFlagSet("force") && IsFlagSet("dryrun") {
		// both together are not supported
		// even that this case is handled correctly
		DebugLog("chkGlobalSyntax failed - both 'force' and 'dryrun' set - unsupported")
		ret = false
	}
	if IsFlagSet("colorscheme") && IsFlagSet("show-non-compliant") && len(os.Args) < cmdLinePos["cmdOpt"]+2 {
		// too few options for both flags set
		DebugLog("chkGlobalSyntax failed - both 'colorscheme' and 'show-non-compliant' set, but too few options")
		ret = false
	}
	if ret {
		// check global option
		ret = chkGlobalOpts(cmdLinePos)
	}
	return ret
}

// chkGlobalOpts checks for global options
// saptune -format FORMAT [--version|--help]
// saptune --version or saptune --help
func chkGlobalOpts(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	globOpt := false
	globPos := 1
	if len(stArgs) < globPos {
		// too few arguments
		DebugLog("chkGlobalOpts failed - too few arguments")
		return false
	}
	if IsFlagSet("version") && IsFlagSet("help") {
		// both together are not supported
		DebugLog("chkGlobalOpts failed - both 'version' and 'help' set - unsupported")
		ret = false
	}
	if IsFlagSet("format") {
		if GetFlagVal("format") != "json" {
			DebugLog("chkGlobalOpts failed - wrong 'format' value '%+v'", GetFlagVal("format"))
			ret = false
		}
		if !strings.Contains(stArgs[1], "--format") {
			DebugLog("chkGlobalOpts failed - 'format' flag on wrong position in command line")
			ret = false
		} else {
			globPos = globPos + 2
			// the flag '--format' has a value (e.g. json), so we
			// have '2' positions to skip in the command line
			cmdLinePos["realm"] = cmdLinePos["realm"] + 2
			cmdLinePos["realmOpt"] = cmdLinePos["realmOpt"] + 2
			cmdLinePos["cmd"] = cmdLinePos["cmd"] + 2
			cmdLinePos["cmdOpt"] = cmdLinePos["cmdOpt"] + 2
		}
	}
	if IsFlagSet("version") {
		// support '--version' and '-version'
		if !strings.Contains(stArgs[globPos], "-version") {
			DebugLog("chkGlobalOpts failed - 'version' flag on wrong position in command line")
			ret = false
		} else {
			globOpt = true
		}
	}
	if IsFlagSet("help") {
		// support '--help' and '-help'
		if !strings.Contains(stArgs[globPos], "-help") {
			DebugLog("chkGlobalOpts failed - 'help' flag on wrong position in command line")
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
			DebugLog("chkRealmOpts failed - 'non-compliance-check' flag used with wrong realm '%+v'", stArgs[cmdLinePos["realm"]])
			ret = false
		}
		if stArgs[cmdLinePos["realm"]] == "status" {
			// realm option set
			// check minimum of values items in cmd line
			// (saptune + realm + option)
			if len(stArgs) < cmdLinePos["realmOpt"]+1 {
				// too few arguments
				DebugLog("chkRealmOpts failed - too few arguments for realm 'status'")
				ret = false
			} else if stArgs[cmdLinePos["realmOpt"]] != "--non-compliance-check" {
				DebugLog("chkRealmOpts failed - 'non-compliance-check' flag on wrong position in command line")
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
		DebugLog("chkCmdOpts failed - too few arguments for flags 'force' or 'dryrun' or 'colorscheme' or 'show-non-compliant'")
		return false
	}
	if len(os.Args) < cmdLinePos["cmdOpt"]+1 || (!IsFlagSet("force") && !IsFlagSet("dryrun") && !IsFlagSet("colorscheme") && !IsFlagSet("show-non-compliant") && !IsFlagSet("non-compliance-check")) {
		// no command options set or too few options
		// and/or non of the flags set, which need further checks
		// so let the 'old' default checks (in main and/or actions) set
		// the appropriate result
		DebugLog("chkCmdOpts nok - no command options set or too few options and/or non of the flags set, which need further checks, so let the 'old' default checks set the appropriate result")
		return true
	}
	// saptune solution change [--force] SOLUTIONNAME
	// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
	if !chkForceFlag(cmdLinePos) {
		DebugLog("chkCmdOpts - chkForceFlag failed")
		ret = false
	}
	// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
	if !chkDryrunFlag(cmdLinePos) {
		DebugLog("chkCmdOpts - chkDryrunFlag failed")
		ret = false
	}
	// saptune note verify [--colorscheme <color scheme>] [--show-non-compliant] [NOTEID]
	// saptune solution verify [--colorscheme <color scheme>] [--show-non-compliant] [SOLUTIONNAME]
	if !chkVerifySyntax(cmdLinePos) {
		DebugLog("chkCmdOpts - chkVerifySyntax failed")
		ret = false
	}
	// saptune (service) status  [--non-compliance-check]
	if !chkServiceStatusSyntax(cmdLinePos) {
		DebugLog("chkCmdOpts - chkServiceStatusSyntax failed")
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
			DebugLog("chkForceFlag failed - 'force' flag used with wrong realm '%+v' or command '%+v'", stArgs[cmdLinePos["realm"]], stArgs[cmdLinePos["cmd"]])
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--force" {
			DebugLog("chkForceFlag failed - 'force' flag on wrong position in command line")
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
			DebugLog("chkDryrunFlag failed - 'dryrun' flag used with wrong realm '%+v' or wrong command '%+v'", stArgs[cmdLinePos["realm"]], stArgs[cmdLinePos["cmd"]])
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--dry-run" {
			DebugLog("chkDryrunFlag failed - 'dryrun' flag on wrong position in command line")
			ret = false
		}
	}
	return ret
}

// chkVerifySyntax checks the syntax of 'saptune note|solution verify' command
// line regarding command line options
// saptune note verify [--colorscheme <color scheme>] [--show-non-compliant] [NOTEID]
// saptune solution verify [--colorscheme <color scheme>] [--show-non-compliant] [SOLUTIONNAME]
func chkVerifySyntax(cmdLinePos map[string]int) bool {
	stArgs := os.Args
	ret := true
	if IsFlagSet("colorscheme") || IsFlagSet("show-non-compliant") {
		if !((stArgs[cmdLinePos["realm"]] == "note" || stArgs[cmdLinePos["realm"]] == "solution") && stArgs[cmdLinePos["cmd"]] == "verify") {
			DebugLog("chkVerifySyntax failed - 'colorscheme' or 'show-non-compliant' flag used with wrong realm '%+v' or wrong command '%+v'", stArgs[cmdLinePos["realm"]], stArgs[cmdLinePos["cmd"]])
			ret = false
		}
	}
	DebugLog("chkVerifySyntax - colorscheme is '%+v', show-non-compliant is '%+v'", GetFlagVal("colorscheme"), IsFlagSet("show-non-compliant"))

	if IsFlagSet("colorscheme") && IsFlagSet("show-non-compliant") {
		// both flags set, check order
		poscor := 2
		if GetFlagVal("colorscheme") == "flag_value" {
			poscor = 1
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--colorscheme" && stArgs[cmdLinePos["cmdOpt"]+poscor] != "--show-non-compliant" {
			DebugLog("chkVerifySyntax failed - wrong order of flags 'colorscheme' and 'show-non-compliant'")
			ret = false
		}
	} else if IsFlagSet("colorscheme") {
		if GetFlagVal("colorscheme") == "--show-non-compliant" {
			DebugLog("chkVerifySyntax failed - missing colorscheme leads to wrong position of 'show-non-compliant' flag in command line")
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--colorscheme" {
			// flag at wrong place in arg list
			DebugLog("chkVerifySyntax failed - 'colorscheme' flag on wrong position in command line")
			ret = false
		}
	} else if IsFlagSet("show-non-compliant") && stArgs[cmdLinePos["cmdOpt"]] != "--show-non-compliant" {
		// flag at wrong place in arg list
		DebugLog("chkVerifySyntax failed - 'show-non-compliant' flag on wrong position in command line")
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
			DebugLog("chkServiceStatusSyntax failed - 'non-compliance-check' flag used with wrong realm '%+v' or wrong command '%+v'", stArgs[cmdLinePos["realm"]], stArgs[cmdLinePos["cmd"]])
			ret = false
		}
		if stArgs[cmdLinePos["cmdOpt"]] != "--non-compliance-check" {
			DebugLog("chkServiceStatusSyntax failed - 'non-compliance-check' flag on wrong position in command line")
			ret = false
		}
	}
	return ret
}
