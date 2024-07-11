package system

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	logFile := "/tmp/saptune_tst.log"
	logSwitch := map[string]string{"verbose": "on", "debug": "on", "error": "on"}

	LogInit(logFile, logSwitch)
	DebugLog("TestMessage%s_%s", "1", "Debug")
	if !CheckForPattern(logFile, "TestMessage1_Debug") {
		t.Error("Debug message found in log file")
	}
	InfoLog("TestMessage%s_%s", "2", "Info")
	if !CheckForPattern(logFile, "TestMessage2_Info") {
		t.Error("Info message not found in log file")
	}
	WarningLog("TestMessage%s_%s", "3", "Warning")
	if !CheckForPattern(logFile, "TestMessage3_Warning") {
		t.Error("Warning message not found in log file")
	}
	ErrorLog("TestMessage%s_%s", "4", "Error")
	if !CheckForPattern(logFile, "TestMessage4_Error") {
		t.Error("Error message not found in log file")
	}
	NoticeLog("TestMessage%s_%s", "5", "Notice")
	if !CheckForPattern(logFile, "TestMessage5_Notice") {
		t.Error("Error message not found in log file")
	}
	ErrorLogNoStdErr("TestMessage%s_%s", "6", "Error")
	if !CheckForPattern(logFile, "TestMessage6_Error") {
		t.Error("Error message not found in log file")
	}
	SwitchOffLogging()
	os.Remove(logFile)
}
