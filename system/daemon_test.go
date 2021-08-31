package system

import (
	"os/exec"
	"testing"
)

func TestSystemctl(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}
	running, _ := IsSystemRunning()
	if !running {
		_, _ = exec.Command("/usr/bin/systemctl", "reset-failed").CombinedOutput()
		running, _ = IsSystemRunning()
		if !running {
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
	if err := SystemctlStatus(testService); err != nil {
		t.Fatal(err)
	}
	active, _ := SystemctlIsRunning(testService)
	if !active {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlRestart(testService); err != nil {
		t.Fatal(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if !active {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlReloadTryRestart(testService); err != nil {
		t.Fatal(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if !active {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlStop(testService); err != nil {
		t.Fatal(err)
	}
	if err := SystemctlStatus(testService); err == nil {
		t.Fatal(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if active {
		t.Fatalf("service '%s' still running\n", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Fatal(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if !active {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Fatal(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if active {
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
	if err := SystemctlStatus("UnkownService"); err == nil {
		t.Fatal(err)
	}

	if SystemctlIsStarting() {
		t.Fatal("systemctl reports system is in state 'starting'")
	}
	sysState, err := GetSystemState()
	if err != nil {
		t.Fatal(err, sysState)
	}
	if sysState != "running" {
		t.Fatalf("'%s'\n", sysState)
	}
}

func TestIsSapconfActive(t *testing.T) {
	sapconf := "sapconf.service"
	if IsSapconfActive(sapconf) {
		t.Fatalf("sapconf service active")
	}
}

func TestSystemctlIsEnabled(t *testing.T) {
	testService := "rpcbind.service"
	enabled, _ := SystemctlIsEnabled(testService)
	if enabled {
		t.Errorf("service '%s' is detected as enabled, but should be disabled", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Errorf("Error enable and start '%s': '%v'\n", testService, err)
	}
	enabled, _ = SystemctlIsEnabled(testService)
	if !enabled {
		t.Errorf("service '%s' is detected as disabled, but should be enabled", testService)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Errorf("Error disable and stop '%s': '%v'\n", testService, err)
	}

	enabled, _ = SystemctlIsEnabled("UnkownService")
	if enabled {
		t.Errorf("service 'UnkownService' is detected as enabled, which is not possible")
	}
}

func TestSystemctlIsRunning(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}
	active, _ := SystemctlIsRunning("dbus.service")
	if !active {
		t.Fatal("'dbus.service' not running")
	}
	active, _ = SystemctlIsRunning("tuned.service")
	if !active {
		t.Log("'tuned.service' not running")
		t.Log("start 'tuned.service' for following tests")
		if err := SystemctlStart("tuned.service"); err != nil {
			t.Log(err)
		}
	}
}

func TestCmpServiceStates(t *testing.T) {
	match := false
	current := "stop, disable"
	expected := "stop"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = ""
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "enable, stop"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "start, disable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "disable, stop"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "stop, start, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, stop, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, hugo, stop, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "start, hugo, stop, enable, disable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "sToP, hugo, start, disable, enable"
	match = CmpServiceStates(current, expected)
	if !match {
		t.Errorf("'%s' should match '%s'\n", expected, current)
	}
	expected = "stop, hugo, start, enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "start, stop, enable"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
	expected = "hugo"
	match = CmpServiceStates(current, expected)
	if match {
		t.Errorf("'%s' should NOT match '%s'\n", expected, current)
	}
}

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

func TestDaemonErrorCases(t *testing.T) {
	oldSystemctlCmd := systemctlCmd
	systemctlCmd = "/usr/bin/false"
	if err := SystemctlRestart("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if err := SystemctlReloadTryRestart("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if err := SystemctlStart("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if err := SystemctlStop("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if IsServiceAvailable("tstserv") {
		t.Error("service 'tstserv' should not, but is available on the system")
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
