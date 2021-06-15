package system

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var infoLogger *log.Logger    // Info logger
var debugLogger *log.Logger   // Debug logger
var errorLogger *log.Logger   // Error logger
var warningLogger *log.Logger // Warning logger
var logOnlyLogger *log.Logger // Log only logger
var debugSwitch string        // Switch Debug on or off
var verboseSwitch string      // Switch verbose mode on or off
var errorSwitch = ""          // Switch error mode on or off
var severInfoFormat = "INFO     "
var severWarnFormat = "WARNING  "
var severErrorFormat = "ERROR    "
var logpidFormat = fmt.Sprintf("saptune[%v] ", os.Getpid()) // format to add pid of current saptune process to the log message

// DebugLog sents text to the debugLogger and stderr
func DebugLog(txt string, stuff ...interface{}) {
	if debugLogger != nil && debugSwitch == "1" {
		debugLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		fmt.Fprintf(os.Stderr, "DEBUG: "+txt+"\n", stuff...)
	}
}

// InfoLog sents text to the infoLogger and stdout
func InfoLog(txt string, stuff ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		if verboseSwitch == "on" {
			fmt.Fprintf(os.Stdout, "    INFO: "+txt+"\n", stuff...)
		}
	}
}

// WarningLog sents text to the warningLogger and stderr
func WarningLog(txt string, stuff ...interface{}) {
	if warningLogger != nil {
		warningLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		if verboseSwitch == "on" {
			fmt.Fprintf(os.Stderr, "    WARNING: "+txt+"\n", stuff...)
		}
	}
}

// ErrorLog sents text to the errorLogger and stderr
func ErrorLog(txt string, stuff ...interface{}) error {
	if errorLogger != nil {
		errorLogger.Printf(CalledFrom()+txt+"\n", stuff...)
		if errorSwitch == "on" {
			fmt.Fprintf(os.Stderr, "ERROR: "+txt+"\n", stuff...)
		}
	}
	return fmt.Errorf(txt+"\n", stuff...)
}

// LogOnlyLog sents text only to the logOnlyLogger
func LogOnlyLog(info, txt string, stuff ...interface{}) (err error) {
	err = nil
	severFormat := "SAPTUNE  "
	if logOnlyLogger != nil {
		switch info {
		case "INFO":
			severFormat = severInfoFormat
		case "WARNING":
			severFormat = severWarnFormat
		case "ERROR":
			severFormat = severErrorFormat
			err = fmt.Errorf(txt+"\n", stuff...)
		}
		logOnlyLogger.Printf(severFormat+logpidFormat+CalledFrom()+txt+"\n", stuff...)
	}
	return
}

// LogInit initialise the different log writer saptune will use
func LogInit(logFile string, logSwitch map[string]string) {
	var saptuneLog io.Writer

	//define log format
	logTimeFormat := time.Now().Format("2006-01-02 15:04:05.000 ")

	//create log file with desired read/write permissions
	saptuneLog, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err.Error())
	}

	//saptuneWriter := io.MultiWriter(os.Stderr, saptuneLog)
	//log.SetOutput(saptuneWriter)
	//log.SetFlags(0)

	debugLogger = log.New(saptuneLog, logTimeFormat+"DEBUG    "+logpidFormat, 0)
	infoLogger = log.New(saptuneLog, logTimeFormat+severInfoFormat+logpidFormat, 0)
	warningLogger = log.New(saptuneLog, logTimeFormat+severWarnFormat+logpidFormat, 0)
	errorLogger = log.New(saptuneLog, logTimeFormat+severErrorFormat+logpidFormat, 0)
	logOnlyLogger = log.New(saptuneLog, logTimeFormat, 0)

	debugSwitch = logSwitch["debug"]
	verboseSwitch = logSwitch["verbose"]
	errorSwitch = "on"
}

// SwitchOffLogging disables logging
func SwitchOffLogging() {
	debugSwitch = "0"
	verboseSwitch = "off"
	errorSwitch = "off"
	log.SetOutput(ioutil.Discard)
}
