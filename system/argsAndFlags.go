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

// IsFlagSet returns true if the flag is available on the command line
// or false if not
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
	stFlags := map[string]string{"force": "false", "dryrun": "false", "help": "false", "version": "false", "show-non-compliant": "false", "format": "", "colorscheme": "", "non-compliance-check": "false", "notSupported": "", "force-color": "false", "fun": "false"}
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
			// skip next command line parameter if it is the value
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
	case "--force-color", "-force-color":
		flags["force-color"] = "true"
	case "--fun", "-fun":
		flags["fun"] = "true"
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

// ChkCliSyntax checks if command line parameter are in the right order
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
// saptune --format FORMAT [--version|--help]
// saptune --force-color
// saptune --fun
// saptune --version or saptune --help
func chkGlobalOpts(cmdLinePos map[string]int) bool {
	globalFlags := []string{"format", "version", "help", "force-color", "fun"}
	stArgs := os.Args
	ret := true
	if len(stArgs) < 1 {
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
	}

	ret, globOpt, posOffset := chkGlobalFlag(globalFlags, ret)
	if globOpt {
		cmdLinePos["realm"] = cmdLinePos["realm"] + posOffset
		cmdLinePos["realmOpt"] = cmdLinePos["realmOpt"] + posOffset
		cmdLinePos["cmd"] = cmdLinePos["cmd"] + posOffset
		cmdLinePos["cmdOpt"] = cmdLinePos["cmdOpt"] + posOffset
	}
	return ret
}

// chkGlobalFlag checks if the global flags are on the right position in the
// command line
func chkGlobalFlag(globalFlags []string, result bool) (bool, bool, int) {
	stArgs := os.Args
	globOpt := false
	posOffset := 1
	setglobFlags := []string{}
	for _, gflag := range globalFlags {
		if IsFlagSet(gflag) {
			setglobFlags = append(setglobFlags, gflag)
			if gflag == "format" {
				// the flag '--format' has a value (e.g. json),
				// so we have '2' positions to skip in the
				// command line
				posOffset = posOffset + 2
			} else {
				posOffset = posOffset + 1
			}
		}
	}
	if posOffset < 1 {
		posOffset = 1
	}
	for _, sflag := range setglobFlags {
		fval := "-" + sflag
		found := false
		for i := 1; i < posOffset; i++ {
			if strings.Contains(stArgs[i], fval) {
				found = true
				globOpt = true
				break
			}
		}
		if !found {
			DebugLog("chkGlobalFlag failed - '%v' flag on wrong position in command line", sflag)
			result = false
		}
	}
	return result, globOpt, posOffset
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

	flagToCheck := []string{
		// saptune solution change [--force] SOLUTIONNAME
		// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
		"chkForceFlag",
		// saptune staging release [--force|--dry-run] [NOTE...|SOLUTION...|all]
		"chkDryrunFlag",
		// saptune note verify [--colorscheme <color scheme>] [--show-non-compliant] [NOTEID]
		// saptune solution verify [--colorscheme <color scheme>] [--show-non-compliant] [SOLUTIONNAME]
		"chkVerifySyntax",
		// saptune (service) status  [--non-compliance-check]
		"chkServiceStatusSyntax",
	}

	for _, flag := range flagToCheck {
		if !checkFlag(cmdLinePos, flag) {
			debugString := "chkCmdOpts - " + flag + " failed"
			DebugLog(debugString)
			ret = false
		}
	}

	return ret
}

// checkFlag checks if the command flags are on the right position in the
// command line
func checkFlag(cmdLinePos map[string]int, flagValue string) bool {
	stArgs := os.Args
	result := true

	runChecks := func(flagCase string, flagValue string, cliArg string, notInRealm bool, isWrongPosition bool) bool {
		result := true
		if IsFlagSet(cliArg) {
			if notInRealm {
				DebugLog("%v failed - '%v' flag used with wrong realm '%+v' or wrong command '%+v'", flagCase, flagValue, stArgs[cmdLinePos["realm"]], stArgs[cmdLinePos["cmd"]])
				result = false
			}
			if isWrongPosition {
				DebugLog("%v failed - '%v' flag on wrong position in command line", flagCase, flagCase)
				result = false
			}
		}
		return result
	}
	syntaxCheckNotRealm := func(realmCommand [][]string) bool {
		var result bool = true
		for _, k := range realmCommand {
			result = result && !(stArgs[cmdLinePos["realm"]] == k[0] && stArgs[cmdLinePos["cmd"]] == k[1])
		}
		return result
	}

	// os.Args = []string{"saptune", "staging", "release", "--force"}
	switch flagValue {
	case "chkForceFlag":
		// Checks the syntax of 'saptune solution change' and 'saptune staging release' regarding the 'force' flag
		notInRealm := syntaxCheckNotRealm([][]string{{"solution", "change"}, {"staging", "release"}})
		isWrongPosition := stArgs[cmdLinePos["cmdOpt"]] != "--force"
		result = runChecks("chkForceFlag", "force", "force", notInRealm, isWrongPosition)

	case "chkServiceStatusSyntax":
		// Checks the syntax of 'saptune service status' or 'saptune daemon status'
		notInRealm := syntaxCheckNotRealm([][]string{{"service", "status"}, {"daemon", "status"}})

		isWrongPosition := stArgs[cmdLinePos["cmdOpt"]] != "--non-compliance-check"
		result = runChecks("chkServiceStatusSyntax", "non-compliance-check", "non-compliance-check", notInRealm, isWrongPosition)

	case "chkDryrunFlag":
		// Checks the syntax of 'saptune staging release' regarding the use of the 'dry-run' flag
		notInRealm := syntaxCheckNotRealm([][]string{{"staging", "release"}})
		isWrongPosition := stArgs[cmdLinePos["cmdOpt"]] != "--dry-run"
		result = runChecks("chkDryrunFlag", "dry-run", "dryrun", notInRealm, isWrongPosition)

	case "chkVerifySyntax":
		result = chkVerifySyntax(stArgs, cmdLinePos, result)
	}

	return result
}

// Checks the syntax of 'saptune note|solution verify' regarding options both flags set, check order flag at wrong place in arg list
func chkVerifySyntax(stArgs []string, cmdLinePos map[string]int, ret bool) bool {
	if IsFlagSet("colorscheme") || IsFlagSet("show-non-compliant") {
		if !((stArgs[cmdLinePos["realm"]] == "note" || stArgs[cmdLinePos["realm"]] == "solution") && stArgs[cmdLinePos["cmd"]] == "verify") {
			DebugLog("chkVerifySyntax failed - 'colorscheme' or 'show-non-compliant' flag used with wrong realm '%+v' or wrong command '%+v'", stArgs[cmdLinePos["realm"]], stArgs[cmdLinePos["cmd"]])
			ret = false
		}
	}
	DebugLog("chkVerifySyntax - colorscheme is '%+v', show-non-compliant is '%+v'", GetFlagVal("colorscheme"), IsFlagSet("show-non-compliant"))

	if IsFlagSet("colorscheme") && IsFlagSet("show-non-compliant") {

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

			DebugLog("chkVerifySyntax failed - 'colorscheme' flag on wrong position in command line")
			ret = false
		}
	} else if IsFlagSet("show-non-compliant") && stArgs[cmdLinePos["cmdOpt"]] != "--show-non-compliant" {

		DebugLog("chkVerifySyntax failed - 'show-non-compliant' flag on wrong position in command line")
		ret = false
	}
	return ret
}
