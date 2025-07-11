package system

import (
	"fmt"
	"io"
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

type loggerInfo struct {
	logger  *log.Logger
	where   *os.File
	typeLog string
	format  string
	called  string
}

// Standard log handler
func messageLogger(infoLogger loggerInfo, txt string, logSwitch string, stuff []interface{}) {
	logger := infoLogger.logger
	where := infoLogger.where
	typeLog := infoLogger.typeLog
	logFormat := infoLogger.format
	calledFrom := infoLogger.called
	if logger != nil {
		logger.SetPrefix(logTimeFormat() + logFormat + logpidFormat)
		logger.Printf(calledFrom+txt+"\n", stuff...)
	}
	if where != nil {
		jWriteMsg(typeLog, fmt.Sprintf(calledFrom+txt+"\n", stuff...))
	}
	if logSwitch == "on" && where != nil {
		fmt.Fprintf(where, typeLog+": "+txt+"\n", stuff...)
	}
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
	messageLogger(loggerInfo{noticeLogger, os.Stdout, "NOTICE", severNoticeFormat, CalledFrom()}, txt, verboseSwitch, stuff)
}

// InfoLog sends text only to the infoLogger
func InfoLog(txt string, stuff ...interface{}) {
	messageLogger(loggerInfo{infoLogger, nil, "INFO", severInfoFormat, CalledFrom()}, txt, "off", stuff)
}

// WarningLog sends text to the warningLogger and stderr
func WarningLog(txt string, stuff ...interface{}) {
	messageLogger(loggerInfo{warningLogger, os.Stderr, "WARNING", severWarnFormat, CalledFrom()}, txt, verboseSwitch, stuff)
}

// ErrorLogNoStdErr sends text only to the errorLogger
func ErrorLogNoStdErr(txt string, stuff ...interface{}) {
	messageLogger(loggerInfo{errorLogger, nil, "ERROR", severErrorFormat, CalledFrom()}, txt, "off", stuff)
}

// ErrorLog sends text to the errorLogger and stderr
func ErrorLog(txt string, stuff ...interface{}) error {
	messageLogger(loggerInfo{errorLogger, os.Stderr, "ERROR", severErrorFormat, CalledFrom()}, txt, errorSwitch, stuff)
	return fmt.Errorf(txt+"\n", stuff...)
}

// LogInit initialise the different log writer saptune will use
func LogInit(logFile string, logSwitch map[string]string) {
	if os.Geteuid() == 0 {
		var saptuneLog io.Writer

		// setup logger and the log destination if called as root
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
	}

	debugSwitch = logSwitch["debug"]
	verboseSwitch = logSwitch["verbose"]
	errorSwitch = logSwitch["error"]
}

// SwitchOffLogging disables logging
func SwitchOffLogging() {
	debugSwitch = "off"
	verboseSwitch = "off"
	errorSwitch = "off"
	log.SetOutput(io.Discard)
}

// PrintLog is a wrapper for the logger to switch on/off writing the
// warning, error and info messages
// not usable, if we need the error return from ErrorLog function
func PrintLog(cnt int, level, msg string, stuff ...interface{}) {
	if cnt == 0 {
		switch level {
		case "warn":
			messageLogger(loggerInfo{warningLogger, os.Stderr, "WARNING", severWarnFormat, CalledFrom()}, msg, verboseSwitch, stuff)
		case "err":
			messageLogger(loggerInfo{errorLogger, os.Stderr, "ERROR", severErrorFormat, CalledFrom()}, msg, errorSwitch, stuff)
		case "info":
			messageLogger(loggerInfo{infoLogger, nil, "INFO", severInfoFormat, CalledFrom()}, msg, "off", stuff)
		case "notice":
			messageLogger(loggerInfo{noticeLogger, os.Stdout, "NOTICE", severNoticeFormat, CalledFrom()}, msg, verboseSwitch, stuff)
		default:
			messageLogger(loggerInfo{infoLogger, nil, "INFO", severInfoFormat, CalledFrom()}, msg, "off", stuff)
		}
	}
}
