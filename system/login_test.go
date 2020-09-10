package system

import (
	"os"
	"testing"
)

func TestGetCurrentLogins(t *testing.T) {
	cmd := "/usr/bin/loginctl"
	ocmd := "/usr/bin/loginctl_OrG"
	val := ""
	for _, userID := range GetCurrentLogins() {
		val = userID
	}
	if val == "" {
		t.Error("no users currently logged in")
	} else if val != "65534" {
		t.Errorf("wrong user listed as logged in - '%s'\n", val)
	}
	err := CopyFile(cmd, ocmd)
	if err == nil {
		_ = os.Chmod(ocmd, 0755)
		_ = CopyFile("/usr/bin/false", cmd)
		curLogins := GetCurrentLogins()
		_ = CopyFile(ocmd, cmd)
		os.Remove(ocmd)
		if len(curLogins) != 0 {
			t.Errorf("found currently logged in users - '%+v'\n", curLogins)
		}
	}
}

func TestGetTasksMax(t *testing.T) {
	cmd := "/usr/bin/systemctl"
	ocmd := "/usr/bin/systemctl_OrG"
	userID := "65534"

	err := CopyFile(cmd, ocmd)
	if err == nil {
		_ = os.Chmod(ocmd, 0755)
		_ = CopyFile("/usr/bin/false", cmd)
		taskMax1 := GetTasksMax(userID)
		_ = CopyFile("/usr/bin/true", cmd)
		taskMax2 := GetTasksMax(userID)
		_ = CopyFile(ocmd, cmd)
		os.Remove(ocmd)
		if taskMax1 != "" {
			t.Errorf("value of UserTasksMax should be empty, but is '%s'\n", taskMax1)
		}
		if taskMax2 != "" {
			t.Errorf("value of UserTasksMax should be empty, but is '%s'\n", taskMax2)
		}
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
