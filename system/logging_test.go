package system

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	logFile := "/tmp/saptune_tst.log"
	debug := "1"
	verbose := "on"
	LogInit(logFile, debug, verbose)
	DebugLog("TestMessage%s_%s", "1", "Debug")
	if !CheckForPattern(logFile, "TestMessage1_Debug") {
		t.Fatal("Debug message found in log file")
	}
	InfoLog("TestMessage%s_%s", "2", "Info")
	if !CheckForPattern(logFile, "TestMessage2_Info") {
		t.Fatal("Info message not found in log file")
	}
	WarningLog("TestMessage%s_%s", "3", "Warning")
	if !CheckForPattern(logFile, "TestMessage3_Warning") {
		t.Fatal("Warning message not found in log file")
	}
	ErrorLog("TestMessage%s_%s", "4", "Error")
	if !CheckForPattern(logFile, "TestMessage4_Error") {
		t.Fatal("Error message not found in log file")
	}
	SwitchOffLogging()
	os.Remove(logFile)
}
