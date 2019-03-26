package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// SystemctlEnable call systemctl enable on thing.
func SystemctlEnable(thing string) error {
	if out, err := exec.Command("systemctl", "enable", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl enable on %s - %v %s", thing, err, string(out))
	}
	return nil
}

// SystemctlDisable call systemctl disable on thing.
func SystemctlDisable(thing string) error {
	if out, err := exec.Command("systemctl", "disable", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl disable on %s - %v %s", thing, err, string(out))
	}
	return nil
}

// SystemctlStart call systemctl start on thing.
func SystemctlStart(thing string) error {
	if out, err := exec.Command("systemctl", "start", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl start on %s - %v %s", thing, err, string(out))
	}
	return nil
}

// SystemctlStop call systemctl stop on thing.
func SystemctlStop(thing string) error {
	if out, err := exec.Command("systemctl", "stop", thing).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call systemctl stop on %s - %v %s", thing, err, string(out))
	}
	return nil
}

// SystemctlEnableStart call systemctl enable and then systemctl start on thing.
func SystemctlEnableStart(thing string) error {
	if err := SystemctlEnable(thing); err != nil {
		return err
	}
	if err := SystemctlStart(thing); err != nil {
		return err
	}
	return nil
}

// SystemctlDisableStop call systemctl disable and then systemctl stop on thing.
// Panic on error.
func SystemctlDisableStop(thing string) error {
	if err := SystemctlDisable(thing); err != nil {
		return err
	}
	if err := SystemctlStop(thing); err != nil {
		return err
	}
	return nil
}

// SystemctlIsRunning return true only if systemctl suggests that the thing is
// running.
func SystemctlIsRunning(thing string) bool {
	if _, err := exec.Command("systemctl", "is-active", thing).CombinedOutput(); err == nil {
		return true
	}
	return false
}

// WriteTunedAdmProfile write new profile to tuned, used instead of sometimes
// unreliable 'tuned-adm' command
func WriteTunedAdmProfile(profileName string) error {
	err := ioutil.WriteFile("/etc/tuned/active_profile", []byte(profileName), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write tuned profile '%s' to '%s': %v", profileName, "/etc/tuned/active_profile", err)
	}
	return nil
}

// GetTunedProfile returns the currently active tuned profile by reading the
// file /etc/tuned/active_profile
// may be unreliable in newer tuned versions, so better use 'tuned-adm active'
// Return empty string if it cannot be determined.
func GetTunedProfile() string {
	content, err := ioutil.ReadFile("/etc/tuned/active_profile")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// TunedAdmOff calls tuned-adm to switch off the active profile.
func TunedAdmOff() error {
	if out, err := exec.Command("tuned-adm", "off").CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call tuned-adm to switch off the active profile - %v %s", err, string(out))
	}
	return nil
}

// TunedAdmProfile calls tuned-adm to switch to the specified profile.
// newer versions of tuned seems to be reliable with this command and they
// changed the behaviour/handling of the file /etc/tuned/active_profile
func TunedAdmProfile(profileName string) error {
	if out, err := exec.Command("tuned-adm", "profile", profileName).CombinedOutput(); err != nil {
		return fmt.Errorf("Failed to call tuned-adm to active profile %s - %v %s", profileName, err, string(out))
	}
	return nil
}

// GetTunedAdmProfile return the currently active tuned profile.
// Return empty string if it cannot be determined.
func GetTunedAdmProfile() string {
	out, err := exec.Command("tuned-adm", "active").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to call tuned-adm to get the active profile - %v %s", err, string(out))
		return ""
	}
	re := regexp.MustCompile(`Current active profile: ([\w-]+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) == 0 {
		return ""
	}
	return matches[1]
}
