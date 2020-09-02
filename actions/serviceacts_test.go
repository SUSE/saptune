package actions

import (
	"bytes"
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/system"
	"os"
	"testing"
)

var sApp *app.App
var saptuneVersion = "3"

var setupSaptuneService = func(t *testing.T) {
	t.Helper()
	_ = system.CopyFile(fmt.Sprintf("%s/etc/sysconfig/saptune", TstFilesInGOPATH), "/etc/sysconfig/saptune")
	sApp = app.InitialiseApp("", "", tuningOpts, AllTestSolutions)
	if err := system.CopyFile("/usr/bin/true", "/usr/sbin/saptune"); err != nil {
		t.Errorf("copy '/usr/bin/true' to '/usr/sbin/saptune' failed - '%v'", err)
	}
	if err := os.Chmod("/usr/sbin/saptune", 0755); err != nil {
		t.Errorf("chmod '/usr/sbin/saptune' failed - '%v'", err)
	}
	if err := system.CopyFile("/app/ospackage/svc/saptune.service", "/usr/lib/systemd/system/saptune.service"); err != nil {
		t.Errorf("copy '/app/ospackage/svc/saptune.service' to '/usr/lib/systemd/system/saptune.service' failed - '%v'", err)
	}
	if err := os.Symlink("/usr/sbin/service", "/usr/sbin/rcsaptune"); err != nil {
		t.Errorf("linking '/usr/sbin/service' to '/usr/sbin/rcsaptune' failed - '%v'", err)
	}
	if err := os.Mkdir("/var/log/saptune", 0755); err != nil {
		t.Errorf("mkdir for '/var/log/saptune' failed - '%v'", err)
	}

	sApp.TuneForSolutions = []string{"sol1"}
	sApp.TuneForNotes = []string{"2205917"}
	sApp.NoteApplyOrder = []string{"2205917"}
}

var teardownSaptuneService = func(t *testing.T) {
	t.Helper()
	os.Remove("/etc/sysconfig/saptune")
	os.Remove("/usr/sbin/saptune")
	os.Remove("/usr/lib/systemd/system/saptune.service")
	os.Remove("/usr/sbin/rcsaptune")
	os.RemoveAll("/var/log/saptune")
}

func TestDaemonActions(t *testing.T) {
	// test setup
	setupSaptuneService(t)
	testService := "saptune.service"

	// Test DaemonActionStart
	t.Run("DaemonActionStart", func(t *testing.T) {
		DaemonAction("start", saptuneVersion, sApp)
		if !system.SystemctlIsRunning(testService) {
			t.Errorf("'%s' not started", testService)
		}
	})
	// Test DaemonActionStatus
	t.Run("DaemonActionStatus", func(t *testing.T) {
		DaemonAction("status", saptuneVersion, sApp)
	})
	// Test DaemonActionStop
	t.Run("DaemonActionStop", func(t *testing.T) {
		DaemonAction("stop", saptuneVersion, sApp)
		if system.SystemctlIsRunning(testService) {
			t.Errorf("'%s' not stopped", testService)
		}
	})

	teardownSaptuneService(t)
}

func TestServiceActions(t *testing.T) {
	// test setup
	setupSaptuneService(t)
	testService := "saptune.service"

	// Test ServiceActionStart
	t.Run("ServiceActionStartandEnable", func(t *testing.T) {
		ServiceActionStart(true, sApp)
		if !system.SystemctlIsRunning(testService) {
			t.Errorf("'%s' not started", testService)
		}
		if !system.SystemctlIsEnabled(testService) {
			t.Errorf("'%s' not enabled", testService)
		}
	})
	// Test ServiceActionStop
	t.Run("ServiceActionStopandDisable", func(t *testing.T) {
		ServiceActionStop(true)
		if system.SystemctlIsEnabled(testService) {
			t.Errorf("'%s' not disabled", testService)
		}
		if system.SystemctlIsRunning(testService) {
			t.Errorf("'%s' not stopped", testService)
		}
	})

	// Test ServiceActionStart
	t.Run("ServiceActionStart", func(t *testing.T) {
		ServiceActionStart(false, sApp)
		if !system.SystemctlIsRunning(testService) {
			t.Errorf("'%s' not started", testService)
		}
	})
	// Test ServiceActionStop
	t.Run("ServiceActionStop", func(t *testing.T) {
		ServiceActionStop(false)
		if system.SystemctlIsRunning(testService) {
			t.Errorf("'%s' not stopped", testService)
		}
	})
	// Test ServiceActionEnable
	t.Run("ServiceActionEnable", func(t *testing.T) {
		ServiceActionEnable()
		if !system.SystemctlIsEnabled(testService) {
			t.Errorf("'%s' not enabled", testService)
		}
	})
	// Test ServiceActionDisable
	t.Run("ServiceActionDisable", func(t *testing.T) {
		ServiceActionDisable()
		if system.SystemctlIsEnabled(testService) {
			t.Errorf("'%s' not disabled", testService)
		}
	})

	// Test ServiceActionApply
	t.Run("ServiceActionApply", func(t *testing.T) {
		ServiceActionApply(sApp)
	})
	// Test ServiceActionRevert
	t.Run("ServiceActionRevert", func(t *testing.T) {
		ServiceActionRevert(sApp)
	})

	// Test ServiceActionStatus
	t.Run("ServiceActionStatus", func(t *testing.T) {
		var serviceStatusMatchText = `The system has been tuned for the following solutions and notes:	sol1	2205917
current order of enabled notes is: 2205917

`
		ServiceActionStart(false, sApp)
		buffer := bytes.Buffer{}
		ServiceActionStatus(&buffer, sApp, saptuneVersion)
		txt := buffer.String()
		checkOut(t, txt, serviceStatusMatchText)
		ServiceActionStop(false)
	})

	teardownSaptuneService(t)
}
