package system

import (
	"os/exec"
	"testing"
)

func TestSystemctl(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}
	if !IsSystemRunning() {
		_, _ = exec.Command("/usr/bin/systemctl", "reset-failed").CombinedOutput()
		if !IsSystemRunning() {
			t.Skip("command '/usr/bin/systemctl is-system-running' reports errors. Skip daemon tests")
		}
	}

	testService := "rpcbind.service"
	if !IsServiceAvailable("rpcbind") {
		t.Fatalf("service 'rpcbind' not available on the system\n")
	}
	if !IsServiceAvailable(testService) {
		t.Fatalf("service '%s' not available on the system\n", testService)
	}
	if err := SystemctlEnable(testService); err != nil {
		t.Fatal(err)
	}
	if err := SystemctlDisable(testService); err != nil {
		t.Fatal(err)
	}
	if err := SystemctlStart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlRestart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlReloadTryRestart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlStop(testService); err != nil {
		t.Fatal(err)
	}
	if SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' still running\n", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Fatal(err)
	}
	if SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' still running\n", testService)
	}

	if IsServiceAvailable("UnkownService") {
		t.Fatalf("service '%s' should not, but is available on the system\n", testService)
	}
	if err := SystemctlEnable("UnkownService"); err == nil {
		t.Fatal(err)
	}
	if err := SystemctlDisable("UnkownService"); err == nil {
		t.Fatal(err)
	}
	if err := SystemctlEnableStart("UnkownService"); err == nil {
		t.Fatal(err)
	}
	if err := SystemctlDisableStop("UnkownService"); err == nil {
		t.Fatal(err)
	}
}

func TestSystemctlIsEnabled(t *testing.T) {
	testService := "rpcbind.service"
	if SystemctlIsEnabled(testService) {
		t.Errorf("service '%s' is detected as enabled, but should be disabled", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Errorf("Error enable and start '%s': '%v'\n", testService, err)
	}
	if !SystemctlIsEnabled(testService) {
		t.Errorf("service '%s' is detected as disabled, but should be enabled", testService)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Errorf("Error disable and stop '%s': '%v'\n", testService, err)
	}

	if SystemctlIsEnabled("UnkownService") {
		t.Errorf("service 'UnkownService' is detected as enabled, which is not possible")
	}
}

func TestSystemctlIsRunning(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}
	if !SystemctlIsRunning("dbus.service") {
		t.Fatal("'dbus.service' not running")
	}
	if !SystemctlIsRunning("tuned.service") {
		t.Log("'tuned.service' not running")
		t.Log("start 'tuned.service' for following tests")
		if err := SystemctlStart("tuned.service"); err != nil {
			t.Log(err)
		}
	}
}

func TestWriteTunedAdmProfile(t *testing.T) {
	profileName := "balanced"
	if err := WriteTunedAdmProfile(profileName); err != nil {
		t.Fatal(err)
	}
	if !CheckForPattern("/etc/tuned/active_profile", profileName) {
		t.Fatal("wrong profile in '/etc/tuned/active_profile'")
	}
	actProfile := GetTunedProfile()
	if actProfile != profileName {
		t.Fatalf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
	profileName = ""
	if err := WriteTunedAdmProfile(profileName); err != nil {
		t.Fatal(err)
	}
	actProfile = GetTunedProfile()
	if actProfile != "" {
		t.Fatalf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
}

func TestGetTunedProfile(t *testing.T) {
	if err := TunedAdmProfile("balanced"); err != nil {
		t.Fatalf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actVal := GetTunedProfile()
	if actVal == "" {
		t.Fatal("seams there is no tuned profile")
	}

	if err := TunedAdmOff(); err != nil {
		t.Fatalf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actVal = GetTunedProfile()
	if actVal != "" {
		t.Fatalf("seams 'tuned-adm off' does not work: profile is '%v'\n", actVal)
	}
}

func TestTunedAdmOff(t *testing.T) {
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmOff(); err != nil {
		t.Fatalf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actProfile := GetTunedProfile()
	if actProfile != "" {
		t.Fatalf("expected profile '%s', current profile '%s'\n", "", actProfile)
	}
	if err := SystemctlStop("tuned"); err != nil {
		t.Fatal(err)
	}
}

func TestTunedAdmProfile(t *testing.T) {
	profileName := "balanced"
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmProfile(profileName); err != nil {
		t.Fatalf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actProfile := GetTunedProfile()
	if actProfile != profileName {
		t.Fatalf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
	if err := TunedAdmOff(); err != nil {
		t.Fatalf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	if err := SystemctlStop("tuned"); err != nil {
		t.Fatal(err)
	}
}

func TestGetTunedAdmProfile(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmProfile("balanced"); err != nil {
		t.Fatalf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actVal := GetTunedAdmProfile()
	if actVal == "" {
		t.Fatal("seams there is no tuned profile")
	}
	if err := TunedAdmOff(); err != nil {
		t.Fatalf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actVal = GetTunedAdmProfile()
	if actVal != "" {
		t.Fatalf("seams 'tuned-adm off' does not work: profile is '%v'\n", actVal)
	}
}
