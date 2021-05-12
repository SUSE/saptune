package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

// saptune lock file
var stLockFile = "/run/.saptune.lock"

// isOwnLock return true, if lock file is from the current running process
// pid inside the lock file is the pid of current running saptune instance
func isOwnLock() bool {
	if !saptuneIsLocked() {
		// no lock file found, return false
		return false
	}
	p, err := ioutil.ReadFile(stLockFile)
	if err != nil {
		ErrorLog("problems during reading the lock file - '%v'", err)
		ReleaseSaptuneLock()
		OSExit(99)
	}
	// file exists, check if empty or if pid inside is from a dead process
	// if yes, remove file and return false
	pid, _ := strconv.Atoi(string(p))
	if pid == os.Getpid() {
		return true
	}
	return false
}

// SaptuneLock creates the saptune lock file
func SaptuneLock() {
	// check for saptune lock file
	if saptuneIsLocked() {
		ErrorExit("saptune currently in use, try later ...", 11)
	}
	stLock, err := os.OpenFile(stLockFile, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
	if err != nil {
		ErrorExit("problems setting lock", 12)
	} else {
		fmt.Fprintf(stLock, "%d", os.Getpid())
	}
	stLock.Close()
}

// saptuneIsLocked checks, if the lock file for saptune exists
func saptuneIsLocked() bool {
	f, err := os.Stat(stLockFile)
	if os.IsNotExist(err) {
		return false
	}
	// file is empty, remove file and return false
	if f.Size() == 0 {
		ReleaseSaptuneLock()
		return false
	}
	// file exists, read content
	p, err := ioutil.ReadFile(stLockFile)
	if err != nil {
		ErrorLog("problems during reading the lock file - '%v'", err)
		ReleaseSaptuneLock()
		OSExit(99)
	}
	// file contains a pid. Check, if process is still alive
	// if not (dead process) remove file and return false
	// TODO - check, if p is really a pid
	pid, _ := strconv.Atoi(string(p))
	if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
		// process exists, must not be the same process, which
		// created the lock file. Will be checked in ErrorExit
		return true
	}
	// process does not exists
	ReleaseSaptuneLock()
	return false
}

// ReleaseSaptuneLock removes the saptune lock file
func ReleaseSaptuneLock() {
	if err := os.Remove(stLockFile); os.IsNotExist(err) {
		// no lock file available, nothing to do
	} else if err != nil {
		ErrorLog("problems removing lock. Please remove lock file '%s' manually before the next start of saptune.\n", stLockFile)
	}
}
