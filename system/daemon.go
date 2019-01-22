package system

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

// Cal systemctl enable and then systemctl start on thing. Panic on error.
func SystemctlEnableStart(thing string) error {
	if out, err := exec.Command("systemctl", "enable", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl enable on %s - %v %s", thing, err, string(out))
	}
	if out, err := exec.Command("systemctl", "start", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl start on %s - %v %s", thing, err, string(out))
	}
	return nil
}

// Cal systemctl disable and then systemctl stop on thing. Panic on error.
func SystemctlDisableStop(thing string) error {
	if out, err := exec.Command("systemctl", "disable", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl disable on %s - %v %s", thing, err, string(out))
	}
	if out, err := exec.Command("systemctl", "stop", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl stop on %s - %v %s", thing, err, string(out))
	}
	return nil
}

// Return true only if systemctl suggests that the thing is running.
func SystemctlIsRunning(thing string) bool {
	if _, err := exec.Command("systemctl", "is-active", thing).CombinedOutput(); err == nil {
		return true
	}
	return false
}

// Write new profile to tuned, used instead of sometimes unreliable 'tuned-adm' command
func WriteTunedAdmProfile(profileName string) error {
	err := ioutil.WriteFile("/etc/tuned/active_profile", []byte(profileName), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write tuned profile '%s' to '%s': %v", profileName, "/etc/tuned/active_profile", err)
        }
        return nil
}

// Return the currently active tuned profile. Return empty string if it cannot be determined.
func GetTunedProfile() string {
	content, err := ioutil.ReadFile("/etc/tuned/active_profile")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}
