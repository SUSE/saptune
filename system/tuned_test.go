package system

import (
	"testing"
)

func TestWriteTunedAdmProfile(t *testing.T) {
	profileName := "balanced"
	if err := WriteTunedAdmProfile(profileName); err != nil {
		t.Log(err)
	}
	if !CheckForPattern("/etc/tuned/active_profile", profileName) {
		t.Log("wrong profile in '/etc/tuned/active_profile'")
	}
	actProfile := GetTunedProfile()
	if actProfile != profileName {
		t.Logf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
	profileName = ""
	if err := WriteTunedAdmProfile(profileName); err != nil {
		t.Log(err)
	}
	actProfile = GetTunedProfile()
	if actProfile != "" {
		t.Logf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
}

func TestGetTunedProfile(t *testing.T) {
	if err := TunedAdmProfile("balanced"); err != nil {
		t.Logf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actVal := GetTunedProfile()
	if actVal == "" {
		t.Log("seams there is no tuned profile")
	}

	if err := TunedAdmOff(); err != nil {
		t.Logf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actVal = GetTunedProfile()
	if actVal != "" {
		t.Logf("seams 'tuned-adm off' does not work: profile is '%v'\n", actVal)
	}
}

func TestTunedAdmOff(t *testing.T) {
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmOff(); err != nil {
		t.Logf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actProfile := GetTunedProfile()
	if actProfile != "" {
		t.Logf("expected profile '%s', current profile '%s'\n", "", actProfile)
	}
	if err := SystemctlStop("tuned"); err != nil {
		t.Log(err)
	}
}

func TestTunedAdmProfile(t *testing.T) {
	profileName := "balanced"
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmProfile(profileName); err != nil {
		t.Logf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actProfile := GetTunedProfile()
	if actProfile != profileName {
		t.Logf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
	if err := TunedAdmOff(); err != nil {
		t.Logf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	if err := SystemctlStop("tuned"); err != nil {
		t.Log(err)
	}
}

func TestGetTunedAdmProfile(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmProfile("balanced"); err != nil {
		t.Logf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actVal := GetTunedAdmProfile()
	if actVal == "" {
		t.Log("seams there is no tuned profile")
	}
	if err := TunedAdmOff(); err != nil {
		t.Logf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actVal = GetTunedAdmProfile()
	if actVal != "" {
		t.Logf("seams 'tuned-adm off' does not work: profile is '%v'\n", actVal)
	}
}

func TestTunedErrorCases(t *testing.T) {
	oldSystemctlCmd := systemctlCmd
	systemctlCmd = "/usr/bin/false"
	if err := TunedAdmOff(); err == nil {
		t.Log("should return an error and not 'nil'")
	}
	systemctlCmd = oldSystemctlCmd

	oldActTunedProfile := actTunedProfile
	actTunedProfile = "/etc/tst/tst/tstProfile"
	actProfile := GetTunedProfile()
	if actProfile != "" {
		t.Log(actProfile)
	}
	profileName := "balanced"
	if err := WriteTunedAdmProfile(profileName); err == nil {
		t.Log("should return an error and not 'nil'")
	}
	actTunedProfile = oldActTunedProfile

	oldTunedAdmCmd := tunedAdmCmd
	tunedAdmCmd = "/usr/bin/false"
	if err := TunedAdmOff(); err == nil {
		t.Log("should return an error and not 'nil'")
	}
	if err := TunedAdmProfile("balanced"); err == nil {
		t.Log("should return an error and not 'nil'")
	}
	tunedAdmCmd = "/usr/bin/true"
	actVal := GetTunedAdmProfile()
	if actVal != "" {
		t.Log(actVal)
	}

	tunedAdmCmd = oldTunedAdmCmd
	_ = SystemctlStop("tuned.service")
	if err := TunedAdmOff(); err != nil {
		t.Log(err)
	}
	_ = SystemctlStart("tuned.service")
}
