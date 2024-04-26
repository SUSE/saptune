package system

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var saptuneLogDir = "/var/log/saptune"
var infoLogger *log.Logger    // Info logger
var noticeLogger *log.Logger  // Notice logger
var debugLogger *log.Logger   // Debug logger
var errorLogger *log.Logger   // Error logger
var warningLogger *log.Logger // Warning logger
var verboseSwitch string      // Switch verbose mode on or off
var errorSwitch string        // Switch error mode on or off
var severDebugFormat = "DEBUG    "
var severNoticeFormat = "NOTICE   "
var severInfoFormat = "INFO     "
var severWarnFormat = "WARNING  "
var severErrorFormat = "ERROR    "
var logpidFormat = fmt.Sprintf("saptune[%v] ", os.Getpid()) // format to add pid of current saptune process to the log message
var debugSwitch = os.Getenv("SAPTUNE_DEBUG")                // Switch Debug on or off

// define log format
func logTimeFormat() string {
	return time.Now().Format("2006-01-02 15:04:05.000 ")
}

// DebugLog sends text to the debugLogger and stderr
func DebugLog(txt string, stuff ...interface{}) {
	if debugSwitch == "on" {
		if debugLogger != nil {
			debugLogger.SetPrefix(logTimeFormat() + severDebugFormat + logpidFormat)
			debugLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		}
		fmt.Fprintf(os.Stderr, "DEBUG: "+txt+"\n", stuff...)
	}
}

// NoticeLog sends text to the noticeLogger and stdout
func NoticeLog(txt string, stuff ...interface{}) {
	if noticeLogger != nil {
		noticeLogger.SetPrefix(logTimeFormat() + severNoticeFormat + logpidFormat)
		noticeLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		jWriteMsg("NOTICE", fmt.Sprintf(CalledFrom()+txt+"\n", stuff...))
		if verboseSwitch == "on" {
			fmt.Fprintf(os.Stdout, "NOTICE: "+txt+"\n", stuff...)
		}
	}
}

// InfoLog sends text only to the infoLogger
func InfoLog(txt string, stuff ...interface{}) {
	if infoLogger != nil {
		infoLogger.SetPrefix(logTimeFormat() + severInfoFormat + logpidFormat)
		infoLogger.Printf(CalledFrom()+txt+"\n", stuff...)
	}
}

// WarningLog sends text to the warningLogger and stderr
func WarningLog(txt string, stuff ...interface{}) {
	if warningLogger != nil {
		warningLogger.SetPrefix(logTimeFormat() + severWarnFormat + logpidFormat)
		warningLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		jWriteMsg("WARNING", fmt.Sprintf(CalledFrom()+txt+"\n", stuff...))
		if verboseSwitch == "on" {
			fmt.Fprintf(os.Stderr, "WARNING: "+txt+"\n", stuff...)
		}
	}
}

// ErrLog sends text only to the errorLogger
func ErrLog(txt string, stuff ...interface{}) {
	if errorLogger != nil {
		errorLogger.SetPrefix(logTimeFormat() + severErrorFormat + logpidFormat)
		errorLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		jWriteMsg("ERROR", fmt.Sprintf(CalledFrom()+txt+"\n", stuff...))
	}
}

// ErrorLog sends text to the errorLogger and stderr
func ErrorLog(txt string, stuff ...interface{}) error {
	if errorLogger != nil {
		errorLogger.SetPrefix(logTimeFormat() + severErrorFormat + logpidFormat)
		errorLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		jWriteMsg("ERROR", fmt.Sprintf(CalledFrom()+txt+"\n", stuff...))
		if errorSwitch == "on" {
			fmt.Fprintf(os.Stderr, "ERROR: "+txt+"\n", stuff...)
		}
	}
	return fmt.Errorf(txt+"\n", stuff...)
}

// LogInit initialise the different log writer saptune will use
func LogInit(logFile string, logSwitch map[string]string) {
	var saptuneLog io.Writer

	if _, err := os.Stat(saptuneLogDir); err != nil {
		if err = os.MkdirAll(saptuneLogDir, 0755); err != nil {
			ErrorExit("", err)
		}
	}
	//create log file with desired read/write permissions
	saptuneLog, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		ErrorExit("", err)
	}

	debugLogger = log.New(saptuneLog, "", 0)
	noticeLogger = log.New(saptuneLog, "", 0)
	infoLogger = log.New(saptuneLog, "", 0)
	warningLogger = log.New(saptuneLog, "", 0)
	errorLogger = log.New(saptuneLog, "", 0)

	debugSwitch = logSwitch["debug"]
	verboseSwitch = logSwitch["verbose"]
	errorSwitch = logSwitch["error"]
}

// SwitchOffLogging disables logging
func SwitchOffLogging() {
	debugSwitch = "off"
	verboseSwitch = "off"
	errorSwitch = "off"
	log.SetOutput(ioutil.Discard)
}
