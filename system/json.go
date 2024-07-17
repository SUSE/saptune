package system

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var schemaDir = "file:///usr/share/saptune/schemas/1.0/"
var supportedRAC map[string]bool = SupportedRACMap()

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
	ConfiguredSol []string `json:"Solution enabled"`
}

// JAppliedSol is for 'Solution applied' in
// 'saptune status' and 'saptune solution applied'
type JAppliedSol struct {
	SolName string `json:"Solution ID,omitempty"`
	Partial *bool  `json:"applied partially,omitempty"`
}

// appliedSol is for 'saptune solution applied'
type appliedSol struct {
	AppliedSol []JAppliedSol `json:"Solution applied"`
}

// appliedNotes is for 'saptune note applied'
type appliedNotes struct {
	AppliedNotes []string `json:"Notes applied"`
}

// notesOrder is for 'saptune note enabled'
type notesOrder struct {
	NotesOrder []string `json:"Notes enabled"`
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
	NotesList  []JNoteListEntry `json:"Notes available"`
	NotesOrder []string         `json:"Notes enabled"`
	Msg        string           `json:"remember message"`
}

// JPNotesLine one row of 'saptune note verify|simulate'
// from PrintNoteFields
type JPNotesLine struct {
	NoteID    string       `json:"Note ID,omitempty"`
	NoteVers  string       `json:"Note version,omitempty"`
	Parameter string       `json:"parameter"`
	Compliant *bool        `json:"compliant,omitempty"`
	ExpValue  string       `json:"expected value,omitempty"`
	OverValue string       `json:"override value,omitempty"`
	ActValue  *string      `json:"actual value,omitempty"`
	Comment   string       `json:"comment,omitempty"`
	Footnotes []JFootNotes `json:"amendments,omitempty"`
}

// JFootNotes collects the footnotes per parameter
type JFootNotes struct {
	FNoteNumber int    `json:"index,omitempty"`
	FNoteTxt    string `json:"amendment,omitempty"`
}

// JPNotesRemind is the reminder section
type JPNotesRemind struct {
	NoteID       string `json:"Note ID,omitempty"`
	NoteReminder string `json:"attention,omitempty"`
}

// JPNotes is the whole 'PrintNoteFields' function
// if we need to differ between 'verify' and 'simulate' this
// can be done in PrintNoteFields' or in jcollect.
type JPNotes struct {
	Verifications []JPNotesLine   `json:"verifications"`
	Simulations   []JPNotesLine   `json:"simulations,omitempty"`
	Attentions    []JPNotesRemind `json:"attentions"`
	NotesOrder    []string        `json:"Notes enabled"`
	SysCompliance *bool           `json:"system compliance"`
}

// JSol - Solution name and related Note list
type JSol struct {
	SolName   string   `json:"Solution ID"`
	NotesList []string `json:"Note list"`
}

// JStatus is the whole 'saptune status'
type JStatus struct {
	Services        JStatusServs   `json:"services"`
	SystemdSysState string         `json:"systemd system state"`
	TuningState     string         `json:"tuning state"`
	VirtEnv         string         `json:"virtualization"`
	SaptuneVersion  string         `json:"configured version"`
	RPMVersion      string         `json:"package version"`
	ConfiguredSol   []string       `json:"Solution enabled"`
	ConfSolNotes    []JSol         `json:"Notes enabled by Solution"`
	AppliedSol      []JAppliedSol  `json:"Solution applied"`
	AppliedSolNotes []JSol         `json:"Notes applied by Solution"`
	ConfiguredNotes []string       `json:"Notes enabled additionally"`
	EnabledNotes    []string       `json:"Notes enabled"`
	AppliedNotes    []string       `json:"Notes applied"`
	Staging         JStatusStaging `json:"staging"`
	Msg             string         `json:"remember message"`
}

// JStatusStaging contains the staging infos for 'saptune status'
type JStatusStaging struct {
	StagingEnabled bool     `json:"staging enabled"`
	StagedNotes    []string `json:"Notes staged"`
	StagedSols     []string `json:"Solutions staged"`
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
	SolName     string   `json:"Solution ID"`
	NotesList   []string `json:"Note list"`
	SolEnabled  bool     `json:"Solution enabled"`
	SolOverride bool     `json:"Solution override exists"`
	CustomSol   bool     `json:"custom Solution"`
	DepSol      bool     `json:"Solution deprecated"`
}

// JSolList is the whole 'saptune solution list'
type JSolList struct {
	SolsList []JSolListEntry `json:"Solutions available"`
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
		//if rac == "solution applied" {
		//	jentry.CmdResult = appliedSol{AppliedSol: res}
		//}
	case JAppliedSol:
		// "saptune solution applied"
		var appSol appliedSol
		if res.SolName != "" {
			appSol.AppliedSol = append(appSol.AppliedSol, res)
		} else {
			appSol.AppliedSol = make([]JAppliedSol, 0)
		}
		jentry.CmdResult = appSol
	case JSolList, JNoteList, JStatus, JPNotes:
		//"solution list", "note list", "status", "daemon status", "service status", "note verify", "solution verify", "note simulate", "solution simulate":
		jentry.CmdResult = res
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
	if rac == "" {
		// check for alias
		if IsFlagSet("version") {
			rac = "version"
		}
		if IsFlagSet("help") {
			rac = "help"
		}
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

// JNoteListEntryInit initialises a JNoteListEntry variable
// used in NoteActionList
func JNoteListEntryInit() JNoteListEntry {
	newListEntry := JNoteListEntry{
		NoteID:       "",
		NoteDesc:     "",
		NoteRef:      make([]string, 0),
		NoteVers:     "",
		NoteRdate:    "",
		ManEnabled:   false,
		SolEnabled:   false,
		ManReverted:  false,
		NoteOverride: false,
		CustomNote:   false,
	}
	return newListEntry
}
