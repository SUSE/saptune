package system

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
)

var tunedAdmCmd = "/usr/sbin/tuned-adm"
var actTunedProfile = "/etc/tuned/active_profile"

// WriteTunedAdmProfile write new profile to tuned, used instead of sometimes
// unreliable 'tuned-adm' command
func WriteTunedAdmProfile(profileName string) error {
	err := ioutil.WriteFile(actTunedProfile, []byte(profileName), 0644)
	if err != nil {
		return ErrorLog("Failed to write tuned profile '%s' to '%s': %v", profileName, actTunedProfile, err)
	}
	return nil
}

// GetTunedProfile returns the currently active tuned profile by reading the
// file /etc/tuned/active_profile
// may be unreliable in newer tuned versions, so better use 'tuned-adm active'
// Return empty string if it cannot be determined.
func GetTunedProfile() string {
	content, err := ioutil.ReadFile(actTunedProfile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// TunedAdmOff calls tuned-adm to switch off the active profile.
func TunedAdmOff() error {
	active, err := SystemctlIsRunning("tuned.service")
	if err != nil {
		return err
	}
	if !active {
		// 'tuned-adm off' does not work without running tuned
		return nil
	}
	if out, err := exec.Command(tunedAdmCmd, "off").CombinedOutput(); err != nil {
		return ErrorLog("Failed to call tuned-adm to switch off the active profile - %v %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// TunedAdmProfile calls tuned-adm to switch to the specified profile.
// newer versions of tuned seems to be reliable with this command and they
// changed the behaviour/handling of the file /etc/tuned/active_profile
func TunedAdmProfile(profileName string) error {
	if out, err := exec.Command(tunedAdmCmd, "profile", profileName).CombinedOutput(); err != nil {
		return ErrorLog("Failed to call tuned-adm to active profile %s - %v %s", profileName, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// GetTunedAdmProfile return the currently active tuned profile.
// Return empty string if it cannot be determined.
func GetTunedAdmProfile() string {
	out, err := exec.Command(tunedAdmCmd, "active").CombinedOutput()
	if err != nil {
		InfoLog("Failed to call tuned-adm to get the active profile - %v %s", err, strings.TrimSpace(string(out)))
		return ""
	}
	re := regexp.MustCompile(`Current active profile: ([\w-]+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) == 0 {
		return ""
	}
	return matches[1]
}
