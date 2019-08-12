package system

import (
	"os"
	"testing"
)

func TestGetCurrentLogins(t *testing.T) {
	val := ""
	for _, userID := range GetCurrentLogins() {
		val = userID
	}
	if val == "" {
		t.Logf("no users currently logged in")
	} else {
		t.Logf("at least user '%s' is logged in\n", val)
	}
}

func TestSetTasksMax(t *testing.T) {
	userID := "65534"
	val := "18446744073709"
	err := SetTasksMax(userID, val)
	if err != nil {
		t.Fatal(err)
	}
	value := GetTasksMax(userID)
	if value != val {
		t.Logf("expected '%s', actual '%s'\n", val, value)
	}
	val = "infinity"
	err = SetTasksMax(userID, val)
	if err != nil {
		t.Fatal(err)
	}
	value = GetTasksMax(userID)
	if value != val {
		t.Logf("expected '%s', actual '%s'\n", val, value)
	}
}

// test with missing loginctl command
func TestMissingLoginctlCmd(t *testing.T) {
	val := ""
	cmdName := "/usr/bin/loginctl"
	savName := "/usr/bin/loginctl_SAVE"
	if err := os.Rename(cmdName, savName); err != nil {
		t.Fatal(err)
	}
	for _, userID := range GetCurrentLogins() {
		val = userID
	}
	if val != "" {
		t.Fatalf("cmd '%s' not available, but user '%s' reported as logged in\n", cmdName, val)
	}
	if err := os.Rename(savName, cmdName); err != nil {
		t.Fatal(err)
	}
}

// test with missing systemctl command
func TestMissingSystemctlCmd(t *testing.T) {
	userID := "65534"
	val := "18446744073709"
	cmdName := "/usr/bin/systemctl"
	savName := "/usr/bin/systemctl_SAVE"
	if err := os.Rename(cmdName, savName); err != nil {
		t.Fatal(err)
	}
	value := GetTasksMax(userID)
	if value != "" {
		t.Fatalf("cmd '%s' not available, but TasksMax='%s' value reported for user '%s'\n", cmdName, value, userID)
	}

	err := SetTasksMax(userID, val)
	if err == nil {
		t.Fatalf("cmd '%s' not available, but error is '%+v'\n", cmdName, err)
	}
	if err := os.Rename(savName, cmdName); err != nil {
		t.Fatal(err)
	}
}
