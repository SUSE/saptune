package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetCurrentLogins returns the user ids of the currently logged in users
func GetCurrentLogins() []string {
	uID := []string{}
	cmdName := "/usr/bin/loginctl"
	cmdArgs := []string{"--no-pager", "--no-legend", "--no-ask-password", "list-users"}
	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return uID
	}
	running, err := IsSystemRunning()
	if err != nil {
		ErrorLog("%v - Failed to call command systemctl", err)
		return uID
	}
	if running {
		cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
		if err != nil {
			WarningLog("failed to invoke external command '%s %v': %v, output: %s", cmdName, cmdArgs, err, string(cmdOut))
			return uID
		}
		for _, logins := range strings.Split(string(cmdOut), "\n") {
			if logins == "" {
				continue
			}
			user := strings.Split(strings.TrimSpace(logins), " ")
			uID = append(uID, user[0])
		}
	}
	return uID
}

// GetTasksMax returns the current limit of TasksMax for a given user id
// which is the value for UserTasksMax
func GetTasksMax(userID string) string {
	//systemctl show -p TasksMax user-<uid>.slice
	uSlice := "user-" + userID + ".slice"
	cmdName := "/usr/bin/systemctl"
	cmdArgs := []string{"show", "-p", "TasksMax", uSlice}

	if !CmdIsAvailable(cmdName) {
		WarningLog("command '%s' not found", cmdName)
		return ""
	}
	cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		WarningLog("failed to invoke external command '%s %v': %v, output: %s", cmdName, cmdArgs, err, string(cmdOut))
		return ""
	}
	tasksMax := strings.Split(strings.TrimSpace(string(cmdOut)), "=")
	// The result of strings.Split of an 'empty' string is a slice with
	// one element - the empty string.
	if len(tasksMax) == 1 && tasksMax[0] == "" {
		return tasksMax[0]
	}
	return tasksMax[1]
}

// SetTasksMax sets the limit of TasksMax for a given user id to 'limit'
func SetTasksMax(userID, limit string) error {
	//systemctl  --runtime set-property user-<uid>.slice TasksMax=infinity
	uSlice := "user-" + userID + ".slice"
	tmLimit := "TasksMax=" + limit
	cmdName := "/usr/bin/systemctl"
	cmdArgs := []string{"--runtime", "set-property", uSlice, tmLimit}

	if !CmdIsAvailable(cmdName) {
		return fmt.Errorf("command '%s' not found", cmdName)
	}
	_, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
	return err
}
