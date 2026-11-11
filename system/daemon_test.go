package system

import (
	"os"
	"os/exec"
	"path"
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
		t.Errorf("service 'rpcbind' not available on the system\n")
	}
	if !IsServiceAvailable(testService) {
		t.Errorf("service '%s' not available on the system\n", testService)
	}
	if err := SystemctlEnable(testService); err != nil {
		t.Error(err)
	}
	if err := SystemctlDisable(testService); err != nil {
		t.Error(err)
	}
	if err := SystemctlStart(testService); err != nil {
		t.Error(err)
	}
	if err := SystemctlStatus(testService); err != nil {
		t.Error(err)
	}
	active, _ := SystemctlIsRunning(testService)
	if !active {
		t.Errorf("service '%s' not running\n", testService)
	}
	if err := SystemctlRestart(testService); err != nil {
		t.Error(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if !active {
		t.Errorf("service '%s' not running\n", testService)
	}
	if err := SystemctlReloadTryRestart(testService); err != nil {
		t.Error(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if !active {
		t.Errorf("service '%s' not running\n", testService)
	}
	if err := SystemctlStop(testService); err != nil {
		t.Error(err)
	}
	if err := SystemctlStatus(testService); err == nil {
		t.Error(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if active {
		t.Errorf("service '%s' still running\n", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Error(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if !active {
		t.Errorf("service '%s' not running\n", testService)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Error(err)
	}
	active, _ = SystemctlIsRunning(testService)
	if active {
		t.Errorf("service '%s' still running\n", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Error(err)
	}
	isactive, err := SystemctlIsActive(testService)
	if isactive == "" {
		t.Errorf("problems getting state of service '%s' - '%+v'\n", testService, err)
	}
	if isactive != "active" {
		t.Errorf("service '%s' not running, state is '%s' - '%+v'\n", testService, isactive, err)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Error(err)
	}
	isactive, err = SystemctlIsActive(testService)
	if isactive == "" {
		t.Errorf("problems getting state of service '%s' - '%+v'\n", testService, err)
	}
	if isactive != "inactive" {
		t.Errorf("service '%s' still running, state is '%s' - '%+v'\n", testService, isactive, err)
	}

	if IsServiceAvailable("UnkownService") {
		t.Errorf("service '%s' should not, but is available on the system\n", testService)
	}
	if err := SystemctlEnable("UnkownService"); err == nil {
		t.Error(err)
	}
	if err := SystemctlDisable("UnkownService"); err == nil {
		t.Error(err)
	}
	if err := SystemctlEnableStart("UnkownService"); err == nil {
		t.Error(err)
	}
	if err := SystemctlDisableStop("UnkownService"); err == nil {
		t.Error(err)
	}
	if err := SystemctlStatus("UnkownService"); err == nil {
		t.Error(err)
	}

	if SysIsStarting() {
		t.Error("systemctl reports system is in state 'starting'")
	}
	sysState, err := GetSystemState()
	if sysState == "degraded" {
		err = SystemctlResetFailed()
	}
	if err != nil {
		t.Error("systemctl reset-failed did not help")
	}
	sysState, err = GetSystemState()
	if err != nil {
		t.Error(err, sysState)
	}
	if sysState != "running" {
		t.Errorf("'%s'\n", sysState)
	}

	err = SystemctlResetFailed()
	if err != nil {
		t.Error(err)
	}

	err = SystemctlDaemonReload(false)
	if err != nil {
		t.Error(err)
	}
}

func TestIsSapconfActive(t *testing.T) {
	sapconf := "sapconf.service"
	if IsSapconfActive(sapconf) {
		t.Errorf("sapconf service active")
	}
	_, _ = ReadConfigFile("/run/sapconf/active", true)
	if !IsSapconfActive(sapconf) {
		t.Errorf("sapconf service NOT active")
	}
	os.RemoveAll("/run/sapconf")
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
		t.Error("'dbus.service' not running")
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
	if _, err := SystemctlIsEnabled("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if _, err := SystemctlIsRunning("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if _, err := SystemctlIsActive("tstserv"); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	if err := SystemctlResetFailed(); err == nil {
		t.Error("should return an error and not 'nil'")
	}
	systemctlCmd = oldSystemctlCmd

	_ = SystemctlStart("tuned.service")
}

func TestSystemdDetectVirt(t *testing.T) {
	opt := ""
	oldSystemddvCmd := systemddvCmd
	// test: virtualization found
	systemddvCmd = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/systemdDVOK")
	exp := "kvm"
	virt, vtype, err := SystemdDetectVirt("-v")
	if !virt {
		t.Error("virtualization should be true, but is false")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}

	exp = "lxc"
	virt, vtype, err = SystemdDetectVirt("-c")
	if !virt {
		t.Error("virtualization should be true, but is false")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}

	exp = ""
	virt, vtype, err = SystemdDetectVirt("-r")
	if !virt {
		t.Error("virtualization should be true, but is false")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}

	exp = "bare-metal"
	virt, vtype, err = SystemdDetectVirt(opt)
	if !virt {
		t.Error("virtualization should be true, but is false")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}

	// test: virtualization NOT available
	systemddvCmd = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/systemdDVNOK")
	exp = "bare-metal"
	virt, vtype, err = SystemdDetectVirt("-v")
	if virt {
		t.Error("virtualization should be false, but is true")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}
	virt, vtype, err = SystemdDetectVirt("-c")
	if virt {
		t.Error("virtualization should be false, but is true")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}
	virt, vtype, err = SystemdDetectVirt(opt)
	if virt {
		t.Error("virtualization should be false, but is true")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err != nil {
		t.Errorf("Test failed, returned error should be 'nil', but got: '%v'", err)
	}
	exp = ""
	virt, vtype, err = SystemdDetectVirt("-r")
	if virt {
		t.Error("virtualization should be false, but is true")
	}
	if vtype != exp {
		t.Errorf("Test failed, expected: '%s', got: '%s'", exp, vtype)
	}
	if err == nil {
		t.Errorf("Test failed, returned error should be NOT 'nil', but got: '%v'", err)
	}

	systemddvCmd = oldSystemddvCmd
}
