package actions

import (
	"bytes"
	"testing"
)

func TestSolutionActions(t *testing.T) {
	// test setup

	// Test SolutionActionList
	t.Run("SolutionActionList", func(t *testing.T) {
		var listMatchText = `
All solutions (* denotes enabled solution, O denotes override file exists for solution, D denotes deprecated solutions):
	BWA                - 941735 2534844 SAP_BWA
	HANA               - 941735 1771258 1980196 1984787 2205917 2382421 2534844
	NETW               - 941735 1771258 1980196 1984787 2534844

Remember: if you wish to automatically activate the solution's tuning options after a reboot,you must enable and start saptune.service by running:
    saptune service enablestart
`

		buffer := bytes.Buffer{}
		SolutionActionList(&buffer, tApp)
		txt := buffer.String()
		checkOut(t, txt, listMatchText)
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
	})

	// Test SolutionActionVerify
	// need to run after 'Test SolutionActionApply'
	t.Run("SolutionActionVerify", func(t *testing.T) {
		var verifyMatchText = `   SAPNote, Version | Parameter                    | Expected    | Override  | Actual      | Compliant
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
