package actions

import (
	"bytes"
	"github.com/SUSE/saptune/system"
	"testing"
)

func TestSolutionActions(t *testing.T) {
	// test setup
	setUp(t)

	// Test SolutionActionList
	t.Run("SolutionActionList", func(t *testing.T) {
		var listMatchText = `
All solutions (* denotes enabled solution, O denotes override file exists for solution, C denotes custom solutions, D denotes deprecated solutions):
	BWA                - 941735 2534844 SAP_BWA
	HANA               - 941735 1771258 1980196 1984787 2205917 2382421 2534844
	MAXDB              - 941735 1771258 1984787
	NETW               - 941735 1771258 1980196 1984787 2534844

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must enable and start saptune.service by running:
    saptune service enablestart
`

		buffer := bytes.Buffer{}
		SolutionActionList(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)
	})

	// Test SolutionActionListCustomOverride
	t.Run("SolutionActionListCustomOVerride", func(t *testing.T) {
		// prepare custom solution and override
		setUpSol(t)

		var listMatchText = `
All solutions (* denotes enabled solution, O denotes override file exists for solution, C denotes custom solutions, D denotes deprecated solutions):
	BWA                - 941735 2534844 SAP_BWA
 O	HANA               - HANA1 NEWNOTE HANA2
 D	MAXDB              - 941735 1771258 1984787
	NETW               - 941735 1771258 1980196 1984787 2534844
 C	NEWSOL1            - SOL1NOTE1 NEWSOL1NOTE SOL1NOTE2
 C	NEWSOL2            - SOL2NOTE1 NEWSOL2NOTE SOL2NOTE2

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must enable and start saptune.service by running:
    saptune service enablestart
`

		buffer := bytes.Buffer{}
		SolutionActionList(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)
		tearDownSol(t)
	})

	// Test SolutionActionSimulate
	t.Run("SolutionActionSimulate", func(t *testing.T) {
		var simulateMatchText = `If you run ` + "`saptune solution apply sol1`" + `, the following changes will be applied to your system:

   Parameter                    | Value set   | Value expected  | Override  | Comment
--------------------------------+-------------+-----------------+-----------+--------------
   net.ipv4.ip_local_port_range | 32768 60999 | 31768 61999     |           |   


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
`
		simBuf := bytes.Buffer{}
		sName := "sol1"
		SolutionActionSimulate(&simBuf, sName, tApp)
		txt := simBuf.String()
		checkOut(t, txt, simulateMatchText)
	})

	// Test SolutionActionApply
	// need to run before 'Test SolutionActionVerify'
	t.Run("SolutionActionApply", func(t *testing.T) {
		var applyMatchText = `All tuning options for the SAP solution have been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must enable and start saptune.service by running:
    saptune service enablestart
`
		buffer := bytes.Buffer{}
		sName := "sol1"
		SolutionActionApply(&buffer, sName, tApp)
		txt := buffer.String()
		checkOut(t, txt, applyMatchText)
		SolutionActionList(&buffer, tApp)
	})

	// Test SolutionActionVerify
	// need to run after 'Test SolutionActionApply'
	t.Run("SolutionActionVerify", func(t *testing.T) {
		var verifyMatchText = `
   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
--------------------+------------------------------+-------------+-----------+-------------+-----------
   simpleNote, 1    | net.ipv4.ip_local_port_range | 31768 61999 |           | 31768 61999 | yes


[31mAttention for SAP Note simpleNote:
Hints or values not yet handled by saptune. So please read carefully, check and set manually, if needed:
# Text to ignore for apply but to display.
# Everything the customer should know about this note, especially
# which parameters are NOT handled and the reason.
[0m
The system fully conforms to the tuning guidelines of the specified SAP solution.
`
		buffer := bytes.Buffer{}
		sName := "sol1"
		SolutionActionVerify(&buffer, sName, tApp)
		txt := buffer.String()
		checkOut(t, txt, verifyMatchText)
	})

	// Test SolutionActionEnabled
	t.Run("SolutionActionEnabled", func(t *testing.T) {
		enabledMatchText := "sol1"

		buffer := bytes.Buffer{}
		SolutionActionEnabled(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, enabledMatchText)
	})

	// Test SolutionActionRevert
	t.Run("SolutionActionRevert", func(t *testing.T) {
		var revertMatchText = `Parameters tuned by the notes referred by the SAP solution have been successfully reverted.
`
		buffer := bytes.Buffer{}
		sName := "sol1"
		SolutionActionRevert(&buffer, sName, tApp)
		txt := buffer.String()
		checkOut(t, txt, revertMatchText)
	})

	tearDown(t)
}

func TestSolutionActionsErrors(t *testing.T) {
	// the error texts returned by the commands and by ErrorExit
	// differs from the 'real' texts because of the test situation.
	// the exit in the ErrorExit function is not executed (as desinged for
	// testing)
	// test setup
	setUp(t)

	// Test SolutionActionApplySecondSol
	t.Run("SolutionActionApplySecondSol", func(t *testing.T) {
		var applyErrorText = `All tuning options for the SAP solution have been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must enable and start saptune.service by running:
    saptune service enablestart
`
		var testErrorText = `ERROR: There is already one solution applied. Applying another solution is NOT supported.
`
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer

		buffer := bytes.Buffer{}
		sName1 := "sol1"
		SolutionActionApply(&buffer, sName1, tApp)
		sol2buffer := bytes.Buffer{}
		sName2 := "sol2"
		SolutionActionApply(&sol2buffer, sName2, tApp)
		txt := sol2buffer.String()
		checkOut(t, txt, applyErrorText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		checkOut(t, errExOut, testErrorText)
		// cleanup, revert the second solution, so that only sol1 is
		// applied
		SolutionActionRevert(&sol2buffer, sName2, tApp)
	})

	// Test SolutionActionSimulateError
	t.Run("SolutionActionSimulateError", func(t *testing.T) {
		// test for PrintHelpAndExit
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		var errExitMatchText = `ERROR: Failed to test the current system against the specified note: solution name "" is not recognised by saptune.
Run "saptune solution list" for a complete list of supported solutions,
and then please double check your input and /etc/sysconfig/saptune
`
		var simErrorMatchText = `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | revertall | enabled | applied ]
  saptune note [ apply | simulate | verify | customise | create | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled ]
  saptune solution [ apply | simulate | verify | revert ] SolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff ]
   saptune staging [ analysis | diff | release ] [ NoteID | solutions | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Print current saptune version:
  saptune version
Print this message:
  saptune help
`

		simBuf := bytes.Buffer{}
		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer
		SolutionActionSimulate(&simBuf, "", tApp)
		txt := simBuf.String()
		checkOut(t, txt, simErrorMatchText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		checkOut(t, errExOut, errExitMatchText)
	})

	// Test SolutionActionApplyError
	t.Run("SolutionActionApplyError", func(t *testing.T) {
		// test for PrintHelpAndExit
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		var errExitMatchText = `ERROR: There is already one solution applied. Applying another solution is NOT supported.
ERROR: Failed to tune for solution : solution name "" is not recognised by saptune.
Run "saptune solution list" for a complete list of supported solutions,
and then please double check your input and /etc/sysconfig/saptune
`
		var applyErrorMatchText = `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | revertall | enabled | applied ]
  saptune note [ apply | simulate | verify | customise | create | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled ]
  saptune solution [ apply | simulate | verify | revert ] SolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff ]
   saptune staging [ analysis | diff | release ] [ NoteID | solutions | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Print current saptune version:
  saptune version
Print this message:
  saptune help
All tuning options for the SAP solution have been applied successfully.

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must enable and start saptune.service by running:
    saptune service enablestart
`

		buffer := bytes.Buffer{}
		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer
		SolutionActionApply(&buffer, "", tApp)
		txt := buffer.String()
		checkOut(t, txt, applyErrorMatchText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		checkOut(t, errExOut, errExitMatchText)
	})

	// Test SolutionActionRevertError
	t.Run("SolutionActionRevertError", func(t *testing.T) {
		// test for PrintHelpAndExit
		oldOSExit := system.OSExit
		defer func() { system.OSExit = oldOSExit }()
		system.OSExit = tstosExit
		oldErrorExitOut := system.ErrorExitOut
		defer func() { system.ErrorExitOut = oldErrorExitOut }()
		system.ErrorExitOut = tstErrorExitOut

		var errExitMatchText = `ERROR: Failed to revert tuning for solution : solution name "" is not recognised by saptune.
Run "saptune solution list" for a complete list of supported solutions,
and then please double check your input and /etc/sysconfig/saptune
`
		var revertErrorMatchText = `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune daemon [ start | status | stop ]  ATTENTION: deprecated
  saptune service [ start | status | stop | restart | takeover | enable | disable | enablestart | disablestop ]
Tune system according to SAP and SUSE notes:
  saptune note [ list | verify | revertall | enabled | applied ]
  saptune note [ apply | simulate | verify | customise | create | revert | show | delete ] NoteID
  saptune note rename NoteID newNoteID
Tune system for all notes applicable to your SAP solution:
  saptune solution [ list | verify | enabled ]
  saptune solution [ apply | simulate | verify | revert ] SolutionName
Staging control:
   saptune staging [ status | enable | disable | is-enabled | list | diff ]
   saptune staging [ analysis | diff | release ] [ NoteID | solutions | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune revert all
Remove the pending lock file from a former saptune call
  saptune lock remove
Print current saptune version:
  saptune version
Print this message:
  saptune help
Parameters tuned by the notes referred by the SAP solution have been successfully reverted.
`

		buffer := bytes.Buffer{}
		errExitbuffer := bytes.Buffer{}
		tstwriter = &errExitbuffer
		SolutionActionRevert(&buffer, "", tApp)
		txt := buffer.String()
		checkOut(t, txt, revertErrorMatchText)
		if tstRetErrorExit != 1 {
			t.Errorf("error exit should be '1' and NOT '%v'\n", tstRetErrorExit)
		}
		errExOut := errExitbuffer.String()
		checkOut(t, errExOut, errExitMatchText)
	})

	tearDown(t)
}
