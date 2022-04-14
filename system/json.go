package system

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var schemaDir = "file:///usr/share/saptune/schemas/1.0/"
var supportedRAC = map[string]bool{"daemon start": false, "daemon status": false, "daemon stop": false, "service start": false, "service status": false, "service stop": false, "service restart": false, "service takeover": false, "service enable": false, "service disable": false, "service enablestart": false, "service disablestop": false, "note list": true, "note revertall": false, "note enabled": true, "note applied": true, "note apply": false, "note simulate": false, "note customise": false, "note create": false, "note edit": false, "note revert": false, "note show": false, "note delete": false, "note verify": false, "note rename": false, "solution list": true, "solution verify": false, "solution enabled": true, "solution applied": true, "solution apply": false, "solution simulate": false, "solution customise": false, "solution create": false, "solution edit": false, "solution revert": false, "solution show": false, "solution delete": false, "solution rename": false, "staging status": false, "staging enable": false, "staging disable": false, "staging is-enabled": false, "staging list": false, "staging diff": false, "staging analysis": false, "staging release": false, "revert all": false, "lock remove": false, "check": false, "status": true, "version": true, "help": false}

// jentry is the json entry to display
var jentry JEntry

// JMsg is a single log message and it's severity/priority
type JMsg struct {
	// Priority of the log messages as defined AT
	// https://confluence.suse.com/display/SAP/Logging+Guide"
	// "CRITICAL", "ERROR" ,"WARNING" ,"NOTICE" ,"INFO", "DEBUG"
	Prio string `json:"priority"`
	// The log message itself
	Txt string `json:"message"`
}

// JMessages contains all log messages normally printed on the screen in the
// order they were created
type JMessages []JMsg

// JResult is the result (output) of the command.
type JResult interface{}

// JEntry defines the global structure of our json object
type JEntry struct {
	// URI to the schema definition
	Schema string `json:"$schema"`
	// saptune timestamp of the time this JSON object was created
	Created string `json:"publish time"`
	// The entire saptune command as it was called
	CmdLine string `json:"argv"`
	// PID of the saptune process creating this object
	Pid int `json:"pid"`
	// The saptune command (classifier), which was execute
	// realm + command, no options, no parameter
	Cmd string `json:"command"`
	// The return code the saptune command terminated with
	CmdRet int `json:"exit code"`
	// The result (output) of the command.
	CmdResult JResult `json:"result"`
	//Contains all log messages normally printed on the screen in the
	// order they were created
	CmdMsg JMessages `json:"messages"`
}

// emptyResult defines an empty result
type emptyResult struct {
}

// notSupportedResult is the result type definition for commands not yet
// supported for json output
type notSupportedResult struct {
	Implemented bool `json:"implemented"`
}

// versResults is the result type definition for 'saptune version'
type versResults struct {
	ConfVers string `json:"configured version"`
}

// configuredSol is for 'saptune solution enabled'
type configuredSol struct {
	ConfiguredSol string `json:"solution configured"`
}

// appliedSol is for 'saptune solution applied'
type appliedSol struct {
	AppliedSol string `json:"solution applied"`
}

// appliedNotes is for 'saptune note applied'
type appliedNotes struct {
	AppliedNotes []string `json:"notes applied"`
}

// notesOrder is for 'saptune note enabled'
type notesOrder struct {
	NotesOrder []string `json:"notes order"`
}

// NoteListEntry is one line of 'saptune note list'
type NoteListEntry struct {
	NoteID       string `json:"Note ID"`
	NoteDesc     string `json:"Note description"`
	ManEnabled   bool   `json:"Note manually enabled"`
	SolEnabled   bool   `json:"Note enabled by solution"`
	ManReverted  bool   `json:"Note manually revertd"`
	NoteOverride bool   `json:"Note override exists"`
	CustomNote   bool   `json:"Custom note"`
}

// NoteList is the whole 'saptune note list'
type NoteList struct {
	NotesList  []NoteListEntry `json:"Note list"`
	NotesOrder []string        `json:"Notes order"`
	Msg        string          `json:"Remember message"`
}

// SolListEntry is one line of 'saptune solution list'
type SolListEntry struct {
	SolName     string   `json:"Solution Name"`
	NotesList   []string `json:"Note list"`
	SolEnabled  bool     `json:"Solution enabled"`
	SolOverride bool     `json:"Note override exists"`
	CustomSol   bool     `json:"Custom note"`
	DepSol      bool     `json:"Solution deprecated"`
}

// SolList is the whole 'saptune solution list'
type SolList struct {
	SolsList []SolListEntry `json:"Solution list"`
	Msg      string         `json:"Remember message"`
}

// jInit creates an initial json entry
// used in system/InitOut
func jInit() {
	created := time.Now().Format("2006-01-02 15:04:05.000")
	rac := realmAndCmd()
	jentry = JEntry{
		Schema:    schemaName(rac),
		Created:   created,
		CmdLine:   strings.Join(os.Args, " "),
		Pid:       os.Getpid(),
		Cmd:       rac,
		CmdResult: emptyResult{},
		CmdMsg:    []JMsg{},
	}
}

// jWriteMsg appends messages from logging to the json entry
// instead of writing to stdout/stderr
// used in system/logging
func jWriteMsg(prio, msg string) {
	var jmsg JMsg
	if GetFlagVal("output") != "json" {
		return
	}
	jmsg.Prio = prio
	jmsg.Txt = msg
	jentry.CmdMsg = append(jentry.CmdMsg, jmsg)
}

// jOut writes the json output to stdout
// used in function system/ErrorExit
func jOut(exit int) error {
	var err error
	if GetFlagVal("output") != "json" {
		return err
	}
	// reset stdout to original setting
	os.Stdout = stdOutOrg
	//fmt.Printf("ANGI: JStore - jentry is '%+v'\n", jentry)
	jentry.CmdRet = exit
	//fmt.Printf("ANGI: JStore 2 - jentry is '%+v'\n", jentry)
	data, err := json.Marshal(jentry)
	if err == nil {
		fmt.Println(string(data))
	}
	return err
}

// JInvalid is the answer of an invalid saptune call
// used in function action/PrintHelpAndExit
func JInvalid(exitStatus int) {
	if exitStatus != 0 {
		jentry.Schema = schemaName("invalid")
		jentry.Cmd = "invalid"
	}
	ErrorExit("", exitStatus)
}

// JnotSupportedYet is the answer of a command without json support yet
// used in function action/SelectAction
func JnotSupportedYet() {
	rac := realmAndCmd()
	if GetFlagVal("output") != "json" || racIsSupported(rac) {
		return
	}
	jentry.CmdResult = notSupportedResult{Implemented: false}
	ErrorExit("", 1)
}

// Jcollect collects the result data
func Jcollect(data interface{}) {
	rac := realmAndCmd()
	if GetFlagVal("output") != "json" || !racIsSupported(rac) {
		return
	}
	switch res := data.(type) {
	case string:
		if rac == "version" {
			jentry.CmdResult = versResults{ConfVers: res}
		}
		if rac == "solution enabled" {
			jentry.CmdResult = configuredSol{ConfiguredSol: res}
		}
		if rac == "solution applied" {
			jentry.CmdResult = appliedSol{AppliedSol: res}
		}
	case []string:
		if rac == "note enabled" {
			jentry.CmdResult = notesOrder{NotesOrder: res}
		}
		if rac == "note applied" {
			jentry.CmdResult = appliedNotes{AppliedNotes: res}
		}
	case SolList, NoteList:
		if rac == "solution list" || rac == "note list" {
			jentry.CmdResult = res
		}
	default:
		WarningLog("Unknown data type '%T' for command '%s' in Jcollect, skipping", data, rac)
	}
}

// schemaName returns the schema string
func schemaName(name string) string {
	return fmt.Sprintf("%ssaptune_%s.schema.json", schemaDir, strings.Replace(name, " ", "_", -1))
}

// realmAndCmd returns the realms name and the command name, if available
func realmAndCmd() string {
	rac := CliArg(1)
	if CliArg(2) != "" {
		rac = rac + " " + CliArg(2)
	}
	return rac
}

// racIsSupported checks, if the combination 'realm command' has json support
func racIsSupported(rac string) bool {
	if _, ok := supportedRAC[rac]; !ok {
		// rac not a valid combination
		// return true to let PrintHelpAndExit later do it's job
		return true
	}
	return supportedRAC[rac]
}
