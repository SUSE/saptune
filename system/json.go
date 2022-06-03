package system

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var schemaDir = "file:///usr/share/saptune/schemas/1.0/"
var supportedRAC = map[string]bool{"daemon start": false, "daemon status": true, "daemon stop": false, "service start": false, "service status": true, "service stop": false, "service restart": false, "service takeover": false, "service enable": false, "service disable": false, "service enablestart": false, "service disablestop": false, "note list": true, "note revertall": false, "note enabled": true, "note applied": true, "note apply": false, "note simulate": false, "note customise": false, "note create": false, "note edit": false, "note revert": false, "note show": false, "note delete": false, "note verify": false, "note rename": false, "solution list": true, "solution verify": false, "solution enabled": true, "solution applied": true, "solution apply": false, "solution simulate": false, "solution customise": false, "solution create": false, "solution edit": false, "solution revert": false, "solution show": false, "solution delete": false, "solution rename": false, "staging status": false, "staging enable": false, "staging disable": false, "staging is-enabled": false, "staging list": false, "staging diff": false, "staging analysis": false, "staging release": false, "revert all": false, "lock remove": false, "check": false, "status": true, "version": true, "help": false}

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

// JObj is a string, a string pointer or a bool for some parts in the result
// of the command.
type JObj interface{}

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
	ConfiguredSol []string `json:"enabled Solution"`
}

//type configuredSol struct {
	//ConfiguredSol JObj `json:"enabled Solution"`
//}

// appliedSol is for 'saptune solution applied'
type appliedSol struct {
	AppliedSol []string `json:"applied Solution"`
}

// appliedNotes is for 'saptune note applied'
type appliedNotes struct {
	AppliedNotes []string `json:"applied Notes"`
}

// notesOrder is for 'saptune note enabled'
type notesOrder struct {
	NotesOrder []string `json:"enabled Notes"`
}

// JNoteListEntry is one line of 'saptune note list'
type JNoteListEntry struct {
	NoteID       string `json:"Note ID"`
	NoteDesc     string `json:"Note description"`
	NoteRef      JObj   `json:"Note reference"`
	NoteVers     string `json:"Note version"`
	NoteRdate    string `json:"Note release date"`
	ManEnabled   bool   `json:"Note enabled manually"`
	SolEnabled   bool   `json:"Note enabled by Solution"`
	ManReverted  bool   `json:"Note reverted manually"`
	NoteOverride bool   `json:"Note override exists"`
	CustomNote   bool   `json:"custom Note"`
}

// JNoteList is the whole 'saptune note list'
type JNoteList struct {
	NotesList  []JNoteListEntry `json:"available Notes"`
	NotesOrder []string         `json:"enabled Notes"`
	Msg        string           `json:"remember message"`
}

// JStatus is the whole 'saptune status'
type JStatus struct {
	Services        JStatusServs   `json:"services"`
	SystemState     string         `json:"system state"`
	VirtEnv         string         `json:"virtualization"`
	SaptuneVersion  string         `json:"configured version"`
	RPMVersion      string         `json:"package version"`
	ConfiguredSol   []string       `json:"enabled Solution"`
	ConfiguredNotes []string       `json:"Notes configured"`
	EnabledNotes    []string       `json:"Notes enabled"`
	AppliedNotes    []string       `json:"Notes applied"`
	Staging         JStatusStaging `json:"staging"`
	Msg             string         `json:"remember message"`
}

// JStatusStaging contains the staging infos for 'saptune status'
type JStatusStaging struct {
	StagingEnabled bool     `json:"enabled"`
	StagedNotes    []string `json:"staged Notes"`
	StagedSols     []string `json:"staged Solutions"`
}

// JStatusServs are the mentioned systemd services in 'saptune status'
type JStatusServs struct {
	SaptuneService JObj    `json:"saptune"`
	SapconfService JObj    `json:"sapconf"`
	TunedService   JObj    `json:"tuned"`
	TunedProfile   *string `json:"tuned profile,omitempty"`
}

// JSolListEntry is one line of 'saptune solution list'
type JSolListEntry struct {
	SolName     string   `json:"Solution name"`
	NotesList   []string `json:"Note list"`
	SolEnabled  bool     `json:"Solution enabled"`
	SolOverride bool     `json:"Solution override exists"`
	CustomSol   bool     `json:"custom Solution"`
	DepSol      bool     `json:"Solution deprecated"`
}

// JSolList is the whole 'saptune solution list'
type JSolList struct {
	SolsList []JSolListEntry `json:"available Solutions"`
	Msg      string          `json:"remember message"`
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
	if GetFlagVal("format") != "json" {
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
	if GetFlagVal("format") != "json" {
		return err
	}
	// reset stdout to original setting
	os.Stdout = stdOutOrg
	jentry.CmdRet = exit
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
	if GetFlagVal("format") != "json" || racIsSupported(rac) {
		return
	}
	jentry.CmdResult = notSupportedResult{Implemented: false}
	ErrorExit("", 1)
}

// Jcollect collects the result data
func Jcollect(data interface{}) {
	rac := realmAndCmd()
	if GetFlagVal("format") != "json" || !racIsSupported(rac) {
		return
	}
	switch res := data.(type) {
	case string:
		if rac == "version" {
			jentry.CmdResult = versResults{ConfVers: res}
		}
	case []string:
		if len(res) == 1 && res[0] == "" {
			// replace empty string by empty slice
			res = make([]string, 0)
		}
		if rac == "note enabled" {
			jentry.CmdResult = notesOrder{NotesOrder: res}
		}
		if rac == "note applied" {
			jentry.CmdResult = appliedNotes{AppliedNotes: res}
		}
		if rac == "solution enabled" {
			jentry.CmdResult = configuredSol{ConfiguredSol: res}
		}
		if rac == "solution applied" {
			jentry.CmdResult = appliedSol{AppliedSol: res}
		}
	case JSolList, JNoteList, JStatus:
		switch rac {
		case "solution list", "note list", "status", "daemon status", "service status":
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
