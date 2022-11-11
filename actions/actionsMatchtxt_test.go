package actions

import (
	"fmt"
	"github.com/SUSE/saptune/system"
)

var noteListMatchText = `
All notes (+ denotes manually enabled notes, * denotes notes enabled by solutions, - denotes notes enabled by solutions but reverted manually later, O denotes override file exists for note, C denotes custom note):
	900929		Linux: STORAGE_PARAMETERS_WRONG_SET and 'mmap() failed'
			Version 7 from 31.07.2017
			https://launchpad.support.sap.com/#/notes/900929
	NEWSOL2NOTE	
	extraNote	Configuration drop in for extra tests
			Version 0 from 04.06.2019
	oldFile		Name_syntax
	simpleNote	Configuration drop in for simple tests
			Version 1 from 09.07.2019
	wrongFileNamesyntax	

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`

var solutionListMatchText = `
All solutions (* denotes enabled solution, O denotes override file exists for solution, C denotes custom solutions, D denotes deprecated solutions):
	BWA                - SAP_BWA
	HANA               - 941735 1771258 1980196 1984787 2205917 2382421 2534844
	MAXDB              - 941735 1771258 1984787
	NETW               - 941735 1771258 1980196 1984787 2534844

Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable and start saptune.service by running:
    saptune service enablestart
`

var saptuneStatusMatchText = fmt.Sprintf(`
saptune.service:          disabled/active
saptune package:          'undef'
configured version:       '3'
enabled Solution:         sol1 (simpleNote)
applied Solution:         
additional enabled Notes: 900929 
enabled Notes:            900929
applied Notes:            
staging:                  disabled
staged Notes:             
staged Solutions:         

sapconf.service:          not available
tuned.service:            disabled/active (profile: '%s')
systemd system state:     running
virtualization:           %s
tuning:                   not tuned

Remember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable saptune.service by running:
 'saptune service enable'.

`, system.GetTunedAdmProfile(), system.GetVirtStatus())

var saptuneStatMatchText = fmt.Sprintf(`
saptune.service:          disabled/inactive
saptune package:          'undef'
configured version:       '3'
enabled Solution:         
applied Solution:         
additional enabled Notes: 
enabled Notes:            
applied Notes:            
staging:                  disabled
staged Notes:             
staged Solutions:         

sapconf.service:          not available
tuned.service:            disabled/active (profile: '%s')
systemd system state:     running
virtualization:           %s
tuning:                   not tuned

Remember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable saptune.service by running:
 'saptune service enablestart'.
Your system has not yet been tuned. Please visit `+"`"+`saptune note`+"`"+` and `+"`"+`saptune solution`+"`"+` to start tuning.

`, system.GetTunedAdmProfile(), system.GetVirtStatus())

var PrintHelpAndExitMatchText = `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune [--output=FORMAT] daemon ( start | stop | status [--non-compliance-check] ) ATTENTION: deprecated
  saptune [--output=FORMAT] service ( start | stop | restart | takeover | enable | disable | enablestart | disablestop | status [--non-compliance-check] )
Tune system according to SAP and SUSE notes:
  saptune [--output=FORMAT] note ( list | revertall | enabled | applied )
  saptune [--output=FORMAT] note ( apply | simulate | customise | create | edit | revert | show | delete ) NOTEID
  saptune [--output=FORMAT] note verify [--colorscheme=SCHEME] [--show-non-compliant] [NOTEID]
  saptune [--output=FORMAT] note rename NOTEID NEWNOTEID
Tune system for all notes applicable to your SAP solution:
  saptune [--output=FORMAT] solution ( list | verify | enabled | applied )
  saptune [--output=FORMAT] solution ( apply | simulate | customise | create | edit | revert | show | delete ) SOLUTIONNAME
  saptune [--output=FORMAT] solution change [--force] SOLUTIONNAME
  saptune [--output=FORMAT] solution verify [--colorscheme=SCHEME] [--show-non-compliant] [SOLUTIONNAME]
  saptune [--output=FORMAT] solution rename SOLUTIONNAME NEWSOLUTIONNAME
Staging control:
   saptune [--output=FORMAT] staging ( status | enable | disable | is-enabled | list )
   saptune [--output=FORMAT] staging ( analysis | diff ) [ ( NOTEID | SOLUTIONNAME )... | all ]
   saptune [--output=FORMAT] staging release [--force|--dry-run] [ ( NOTEID | SOLUTIONNAME )... | all ]
Revert all parameters tuned by the SAP notes or solutions:
  saptune [--output=FORMAT] revert all
Remove the pending lock file from a former saptune call
  saptune [--output=FORMAT] lock remove
Call external script '/usr/sbin/saptune_check'
  saptune [--output=FORMAT] check
Print current saptune status:
  saptune [--output=FORMAT] status [--non-compliance-check]
Print current saptune version:
  saptune [--output=FORMAT] version
Print this message:
  saptune [--output=FORMAT] help
`
