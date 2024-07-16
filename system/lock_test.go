package system

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestLock(t *testing.T) {
	if saptuneIsLocked() {
		_, err := os.Stat(stLockFile)
		if os.IsNotExist(err) {
			t.Errorf("saptune lock does NOT exists, but is reported as existing\n")
		} else {
			t.Errorf("saptune lock exists, but shouldn't\n")
		}
	}

	os.Args = []string{"saptune", "note", "apply"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()

	SaptuneLock()
	if !saptuneIsLocked() {
		_, err := os.Stat(stLockFile)
		if os.IsNotExist(err) {
			t.Errorf("saptune should be locked, but isn't\n")
		} else {
			t.Errorf("saptune lock exists, but is reported as non-existing\n")
		}
	}
	if !isOwnLock() {
		pid := -1
		p, err := os.ReadFile(stLockFile)
		if err == nil {
			pid, _ = strconv.Atoi(string(p))
		}
		t.Errorf("wrong pid found in lock file: '%d' instead of '%d'\n", pid, os.Getpid())
	}
	ReleaseSaptuneLock()
	if saptuneIsLocked() {
		_, err := os.Stat(stLockFile)
		if os.IsNotExist(err) {
			t.Errorf("saptune lock does NOT exists, but is reported as existing\n")
		} else {
			t.Errorf("saptune lock exists, but shouldn't\n")
			os.Remove(stLockFile)
		}
	}

	os.Args = []string{"saptune", "note", "list"}
	// parse command line, to get the test parameters
	saptArgs, saptFlags = ParseCliArgs()
	SaptuneLock()
	_, err := os.Stat(stLockFile)
	if err == nil {
		t.Errorf("saptune lock exists, but shouldn't\n")
		os.Remove(stLockFile)
	}

	sl, _ := os.OpenFile(stLockFile, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
	fmt.Fprintf(sl, "")
	saptuneIsLocked()
	os.Remove(stLockFile)
	sl, _ = os.OpenFile(stLockFile, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
	fmt.Fprintf(sl, "%d", 4711)
	saptuneIsLocked()
	os.Remove(stLockFile)
	ReleaseSaptuneLock()
}
