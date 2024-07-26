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
	//sApp.TuneForNotes = []string{"2205917"}
	//sApp.NoteApplyOrder = []string{"2205917"}
	sApp.TuneForNotes = []string{"900929"}
	sApp.NoteApplyOrder = []string{"900929"}
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

	// ANGI TODO - need to clarify the problems with tuned.service
	// and 'Job for tuned.service canceled.'
	// Test DaemonActionStart
	//t.Run("DaemonActionStart", func(t *testing.T) {
	//	DaemonAction(os.Stdout, "start", saptuneVersion, sApp)
	//	if !system.SystemctlIsRunning(testService) {
	//		t.Errorf("'%s' not started", testService)
	//	}
	//})

	// Test DaemonActionStatus
	t.Run("DaemonActionStatus", func(t *testing.T) {
		ServiceActionStart(false, sApp)

		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut
		buffer := bytes.Buffer{}
		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer
		DaemonAction(&buffer, "status", saptuneVersion, sApp)
		txt := buffer.String()
		checkOut(t, txt, saptuneStatusMatchText)
		errExOut := errExitbuffer.String()
		if errExOut != "" {
			t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
		}
	})
	// Test DaemonActionStop
	t.Run("DaemonActionStop", func(t *testing.T) {
		DaemonAction(os.Stdout, "stop", saptuneVersion, sApp)
		enabled, _ := system.SystemctlIsEnabled(testService)
		if enabled {
			t.Errorf("'%s' not disabled", testService)
		}
		active, _ := system.SystemctlIsRunning(testService)
		if active {
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
		active, _ := system.SystemctlIsRunning(testService)
		if !active {
			t.Errorf("'%s' not started", testService)
		}
		enabled, _ := system.SystemctlIsEnabled(testService)
		if !enabled {
			t.Errorf("'%s' not enabled", testService)
		}
	})
	// Test ServiceActionStop
	t.Run("ServiceActionStopandDisable", func(t *testing.T) {
		ServiceActionStop(true)
		enabled, _ := system.SystemctlIsEnabled(testService)
		if enabled {
			t.Errorf("'%s' not disabled", testService)
		}
		active, _ := system.SystemctlIsRunning(testService)
		if active {
			t.Errorf("'%s' not stopped", testService)
		}
	})

	// Test ServiceActionStart
	t.Run("ServiceActionStart", func(t *testing.T) {
		ServiceActionStart(false, sApp)
		active, _ := system.SystemctlIsRunning(testService)
		if !active {
			t.Errorf("'%s' not started", testService)
		}
	})
	// Test ServiceActionStop
	t.Run("ServiceActionStop", func(t *testing.T) {
		ServiceActionStop(false)
		active, _ := system.SystemctlIsRunning(testService)
		if active {
			t.Errorf("'%s' not stopped", testService)
		}
	})
	// Test ServiceActionEnable
	t.Run("ServiceActionEnable", func(t *testing.T) {
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer

		ServiceActionEnable()
		if tstRetErrorExit != 0 {
			t.Logf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		if errExOut != "" {
			t.Logf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
		}

		enabled, _ := system.SystemctlIsEnabled(testService)
		if !enabled {
			t.Errorf("'%s' not enabled", testService)
		}
	})
	// Test ServiceActionDisable
	t.Run("ServiceActionDisable", func(t *testing.T) {
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer

		ServiceActionDisable()
		if tstRetErrorExit != 0 {
			t.Logf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		if errExOut != "" {
			t.Logf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
		}

		enabled, _ := system.SystemctlIsEnabled(testService)
		if enabled {
			t.Errorf("'%s' not disabled", testService)
		}
	})
	// Test ServiceActionApply
	t.Run("ServiceActionApply", func(t *testing.T) {
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer

		ServiceActionApply(sApp)
		if tstRetErrorExit != 0 {
			t.Logf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		if errExOut != "" {
			t.Logf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
		}
	})
	// Test ServiceActionRevert
	t.Run("ServiceActionRevert", func(t *testing.T) {
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer

		ServiceActionRevert(sApp)
		if tstRetErrorExit != 0 {
			t.Logf("error exit should be '0' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		if errExOut != "" {
			t.Logf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
		}
	})

	// Test ServiceActionStatus
	t.Run("ServiceActionStatus", func(t *testing.T) {
		ServiceActionStart(false, sApp)

		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut
		buffer := bytes.Buffer{}
		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer
		ServiceActionStatus(&buffer, sApp, saptuneVersion)
		txt := buffer.String()
		checkOut(t, txt, saptuneStatusMatchText)
		errExOut := errExitbuffer.String()
		if errExOut != "" {
			t.Errorf("wrong text returned by ErrorExit: '%v' instead of ''\n", errExOut)
		}

		ServiceActionStop(false)
	})

	// ANGI TODO - need to clarify the problems with tuned.service
	// and 'Job for tuned.service canceled.'
	// Test ServiceActionTakeover
	//t.Run("ServiceActionTakeover", func(t *testing.T) {
	//	ServiceActionTakeover(sApp)
	//})

	teardownSaptuneService(t)
}
