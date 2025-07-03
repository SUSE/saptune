package actions

func cmdLineSyntax() string {
	return `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune [--format FORMAT] [--force-color] [--fun] daemon ( start | stop | status [--non-compliance-check] ) ATTENTION: deprecated
  saptune [--format FORMAT] [--force-color] [--fun] service ( start | stop | restart | takeover | enable | disable | enablestart | disablestop | status [--non-compliance-check] )
Tune system according to SAP and SUSE notes:
  saptune [--format FORMAT] [--force-color] [--fun] note ( list | verify | revertall | enabled | applied )
  saptune [--format FORMAT] [--force-color] [--fun] note ( apply | simulate | customise | create | edit | revert | show | delete ) NOTEID
  saptune [--format FORMAT] [--force-color] [--fun] note refresh [NOTEID|applied] ATTENTION: experimental
  saptune [--format FORMAT] [--force-color] [--fun] note verify [--colorscheme SCHEME] [--show-non-compliant] [NOTEID|applied]
  saptune [--format FORMAT] [--force-color] [--fun] note rename NOTEID NEWNOTEID
Tune system for all notes applicable to your SAP solution:
  saptune [--format FORMAT] [--force-color] [--fun] solution ( list | verify | enabled | applied )
  saptune [--format FORMAT] [--force-color] [--fun] solution ( apply | simulate | customise | create | edit | revert | show | delete ) SOLUTIONNAME
  saptune [--format FORMAT] [--force-color] [--fun] solution change [--force] SOLUTIONNAME
  saptune [--format FORMAT] [--force-color] [--fun] solution verify [--colorscheme SCHEME] [--show-non-compliant] [SOLUTIONNAME]
  saptune [--format FORMAT] [--force-color] [--fun] solution rename SOLUTIONNAME NEWSOLUTIONNAME
Staging control:
   saptune [--format FORMAT] [--force-color] [--fun] staging ( status | enable | disable | is-enabled | list )
   saptune [--format FORMAT] [--force-color] [--fun] staging ( analysis | diff ) [ ( NOTEID | SOLUTIONNAME.sol )... | all ]
   saptune [--format FORMAT] [--force-color] [--fun] staging release [--force|--dry-run] [ ( NOTEID | SOLUTIONNAME.sol )... | all ]
Config (re-)settings:
  saptune [--format FORMAT] [--force-color] [--fun] configure ( COLOR_SCHEME | SKIP_SYSCTL_FILES | IGNORE_RELOAD | DEBUG | TrentoASDP ) Value
  saptune [--format FORMAT] [--force-color] [--fun] configure ( reset | show )
Verify all applied Notes:
  saptune [--format FORMAT] [--force-color] [--fun] verify applied
Refresh all applied Notes:
  saptune [--format FORMAT] [--force-color] [--fun] refresh applied ATTENTION: experimental
Revert all parameters tuned by the SAP notes or solutions:
  saptune [--format FORMAT] [--force-color] [--fun] revert all
Remove the pending lock file from a former saptune call
  saptune [--format FORMAT] [--force-color] [--fun] lock remove
Call external script '/usr/sbin/saptune_check'
  saptune [--format FORMAT] [--force-color] [--fun] check
Print current saptune status:
  saptune [--format FORMAT] [--force-color] [--fun] status [--non-compliance-check]
Print current saptune version:
  saptune [--format FORMAT] [--force-color] [--fun] version
Print this message:
  saptune [--format FORMAT] [--force-color] [--fun] help

Deprecation list:
  all 'saptune daemon' actions
  'saptune note simulate'
  'saptune solution simulate'
  'Solution SAP-ASE.sol and related Notes 1680803, 1805750'
  'Note 1771258 and PAM limits handling'
`
}

func cmdLineSyntax16() string {
	return `saptune: Comprehensive system optimisation management for SAP solutions.
Daemon control:
  saptune [--format FORMAT] [--force-color] [--fun] service ( start | stop | restart | takeover | enable | disable | enablestart | disablestop | status [--non-compliance-check] )
Tune system according to SAP and SUSE notes:
  saptune [--format FORMAT] [--force-color] [--fun] note ( list | verify | revertall | enabled | applied )
  saptune [--format FORMAT] [--force-color] [--fun] note ( apply | customise | create | edit | revert | show | delete ) NOTEID
  saptune [--format FORMAT] [--force-color] [--fun] note refresh [NOTEID|applied] ATTENTION: experimental
  saptune [--format FORMAT] [--force-color] [--fun] note verify [--colorscheme SCHEME] [--show-non-compliant] [NOTEID|applied]
  saptune [--format FORMAT] [--force-color] [--fun] note rename NOTEID NEWNOTEID
Tune system for all notes applicable to your SAP solution:
  saptune [--format FORMAT] [--force-color] [--fun] solution ( list | verify | enabled | applied )
  saptune [--format FORMAT] [--force-color] [--fun] solution ( apply | customise | create | edit | revert | show | delete ) SOLUTIONNAME
  saptune [--format FORMAT] [--force-color] [--fun] solution change [--force] SOLUTIONNAME
  saptune [--format FORMAT] [--force-color] [--fun] solution verify [--colorscheme SCHEME] [--show-non-compliant] [SOLUTIONNAME]
  saptune [--format FORMAT] [--force-color] [--fun] solution rename SOLUTIONNAME NEWSOLUTIONNAME
Staging control:
   saptune [--format FORMAT] [--force-color] [--fun] staging ( status | enable | disable | is-enabled | list )
   saptune [--format FORMAT] [--force-color] [--fun] staging ( analysis | diff ) [ ( NOTEID | SOLUTIONNAME.sol )... | all ]
   saptune [--format FORMAT] [--force-color] [--fun] staging release [--force|--dry-run] [ ( NOTEID | SOLUTIONNAME.sol )... | all ]
Config (re-)settings:
  saptune [--format FORMAT] [--force-color] [--fun] configure ( COLOR_SCHEME | SKIP_SYSCTL_FILES | IGNORE_RELOAD | DEBUG | TrentoASDP ) Value
  saptune [--format FORMAT] [--force-color] [--fun] configure ( reset | show )
Verify all applied Notes:
  saptune [--format FORMAT] [--force-color] [--fun] verify applied
Refresh all applied Notes:
  saptune [--format FORMAT] [--force-color] [--fun] refresh applied ATTENTION: experimental
Revert all parameters tuned by the SAP notes or solutions:
  saptune [--format FORMAT] [--force-color] [--fun] revert all
Remove the pending lock file from a former saptune call
  saptune [--format FORMAT] [--force-color] [--fun] lock remove
Call external script '/usr/sbin/saptune_check'
  saptune [--format FORMAT] [--force-color] [--fun] check
Print current saptune status:
  saptune [--format FORMAT] [--force-color] [--fun] status [--non-compliance-check]
Print current saptune version:
  saptune [--format FORMAT] [--force-color] [--fun] version
Print this message:
  saptune [--format FORMAT] [--force-color] [--fun] help

Deprecation list:
`
}
