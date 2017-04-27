package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"strconv"
	"strings"
	"runtime"
	"path"
	"log"
)

const (
	INISectionSysctl = "sysctl"
	INISectionVM     = "vm"
	INISectionBlock  = "block"
	INISectionLimits = "limits"
	SYS_THP = "kernel/mm/transparent_hugepage/enabled"
)

// Calculate optimum parameter value given the current value, comparison operator, and expected value. Return optimised value.
func CalculateOptimumValue(operator txtparser.Operator, currentValue string, expectedValue string) (string, error) {
	if operator == txtparser.OperatorEqual {
		return expectedValue, nil
	}
	// Numeric comparisons
	var iCurrentValue int64
	iExpectedValue, err := strconv.ParseInt(expectedValue, 10, 64)
	if err != nil {
		return "", fmt.Errorf("Expected value \"%s\" should be but is not an integer", expectedValue)
	}
	if currentValue == "" {
		switch operator {
		case txtparser.OperatorLessThan:
			iCurrentValue = iExpectedValue - 1
		case txtparser.OperatorMoreThan:
			iCurrentValue = iExpectedValue + 1
		}
	} else {
		iCurrentValue, err = strconv.ParseInt(currentValue, 10, 64)
		if err != nil {
			return "", fmt.Errorf("Current value \"%s\" should be but is not an integer", currentValue)
		}
		switch operator {
		case txtparser.OperatorLessThan:
			if iCurrentValue >= iExpectedValue {
				iCurrentValue = iExpectedValue - 1
			}
		case txtparser.OperatorMoreThan:
			if iCurrentValue <= iExpectedValue {
				iCurrentValue = iExpectedValue + 1
			}
		}
	}
	return strconv.FormatInt(iCurrentValue, 10), nil
}

// section handling

// section [block]
// section [limits]
func GetLimitsVal(key string) string {
	// Find out current memlock limits
	LimitMemlock := 0
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return "0"
	}
	switch key {
	case "MEMLOCK_HARD":
		LimitMemlock, _ = secLimits.Get("sybase", "hard", "memlock")
	case "MEMLOCK_SOFT":
		LimitMemlock, _ = secLimits.Get("sybase", "soft", "memlock")

	}
	return strconv.Itoa(LimitMemlock)
}

func OptLimitsVal(act_value, cfg_value  string) string {
	LimitMemlock, _ := strconv.Atoi(act_value)
	if cfg_value == "0" {
		//RAM in KB - 10%
		memlock := system.GetMainMemSizeMB() * 1024 - (system.GetMainMemSizeMB() * 1024 * 10 / 100)
		LimitMemlock = param.MaxI(LimitMemlock, int(memlock))
	} else {
		LimLockCFG, _ := strconv.Atoi(cfg_value)
		LimitMemlock = param.MaxI(LimitMemlock, LimLockCFG)
	}
	return strconv.Itoa(LimitMemlock)

}

func SetLimitsVal(value string) error {
	LimitMemlock, _ := strconv.Atoi(value)
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return err
	}
	//ANGI TODO: set hard ans soft limit independent
	//ANGI TODO: user should be variable, not fix 'sybase'
	secLimits.Set("sybase", "hard", "memlock", LimitMemlock)
	secLimits.Set("sybase", "soft", "memlock", LimitMemlock)
	err = secLimits.Apply()
	return err
}

// section [vm]
// Manipulate /sys/kernel/mm switches.
func GetVmVal(parameter string) string {
	var val string
	switch parameter {
	case "INI_THP":
		val = system.GetSysChoice(SYS_THP)
	}
	return val
}

func OptVmVal(parameter, act_value, cfg_value  string) string {
	val := act_value
	switch parameter {
	case "INI_THP":
		sval := strings.ToLower(cfg_value)
		if sval != "yes" && sval != "no" {
			fmt.Println("wrong selection for INI_THP. Now set to 'yes' to disable transarent huge pages")
			sval = "yes"
		}
		if sval == "yes" && act_value != "never" {
			val = "never"
		}
		if sval == "no" && act_value == "never" {
			val = "always"
		}
	}
	return val
}


// Tuning options composed by a third party vendor.
type INISettings struct {
	ConfFilePath    string            // Full path to the 3rd party vendor's tuning configuration file
	ID              string            // ID portion of the tuning configuration
	DescriptiveName string            // Descriptive name portion of the tuning configuration
	SysctlParams    map[string]string // Sysctl parameter values from the computer system
}

func (vend INISettings) Name() string {
	return vend.DescriptiveName
}

func (vend INISettings) Initialise() (Note, error) {

	//actBlk := BlockDeviceQueue{}
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}

	// Read current parameter values
	vend.SysctlParams = make(map[string]string)
	for _, param := range ini.AllValues {
		switch param.Section {
		case INISectionSysctl:
			currValue := system.GetSysctlString(param.Key, "")
			if currValue == "" {
				return vend, fmt.Errorf("INISettings %s: cannot find parameter \"%s\" in system", vend.ID, param)
			}
			vend.SysctlParams[param.Key] = currValue
		case INISectionVM:
			vend.SysctlParams[param.Key] = GetVmVal(param.Key)
		case INISectionBlock:
			//vend.SysctlParams[param.Key] = GetBlockVal(param.Key)
		case INISectionLimits:
			vend.SysctlParams[param.Key] = GetLimitsVal(param.Key)
		default:
			// saptune does not yet understand settings outside of [sysctl] section
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	return vend, nil
}

func (vend INISettings) Optimise() (Note, error) {
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return vend, err
	}

	vend.SysctlParams = make(map[string]string)
	for _, param := range ini.AllValues {
		// Compare current values against INI's definition
		switch param.Section {
		case INISectionSysctl:
			optimisedValue, err := CalculateOptimumValue(param.Operator, vend.SysctlParams[param.Key], param.Value)
			if err != nil {
				return vend, err
			}
			vend.SysctlParams[param.Key] = optimisedValue
		case INISectionVM:
			vend.SysctlParams[param.Key] = OptVmVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionBlock:
			//vend.SysctlParams[param.Key] = OptBlkVal(param.Key, vend.SysctlParams[param.Key], param.Value)
		case INISectionLimits:
			vend.SysctlParams[param.Key] = OptLimitsVal(vend.SysctlParams[param.Key], param.Value)
		default:
			// saptune does not yet understand settings outside of [sysctl] section
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	return vend, nil
}

func (vend INISettings) Apply() error {
	// Parse the configuration file
	ini, err := txtparser.ParseINIFile(vend.ConfFilePath, false)
	if err != nil {
		return err
	}
	//for key, value := range vend.SysctlParams {
	for _, param := range ini.AllValues {
		switch param.Section {
		case INISectionSysctl:
			// Apply sysctl parameters
			system.SetSysctlString(param.Key, vend.SysctlParams[param.Key])
		case INISectionVM:
			if runtime.GOARCH == ARCH_X86 {
				system.SetSysString(system.SYS_THP, vend.SysctlParams[param.Key])
			}
		case INISectionBlock:
		case INISectionLimits:
			if runtime.GOARCH == ARCH_X86 {
				SetLimitsVal(vend.SysctlParams[param.Key])
			}
		default:
			// saptune does not yet understand settings outside of [sysctl] section
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	return nil
}


// as workaround till SetBlockVal and GetBlockVal is running
// as 'shadow' note "Block"
/*
1680803 - SYB: SAP Adaptive Server Enterprise - Best Practice for SAP Business Suite and SAP BW
Set BlockDeviceSchedulers
Set BlockDeviceNrRequests
*/
const SYBASE_SYSCONFIG = "/etc/saptune/extra/SAP_ASE-SAP_Adaptive_Server_Enterprise.conf"

type ASERecommendedOSSettings struct {
	BlockDeviceSchedulers     param.BlockDeviceSchedulers
	BlockDeviceNrRequests     param.BlockDeviceNrRequests
}

func (ase ASERecommendedOSSettings) Name() string {
	return "SAP Adaptive Server Enterprise"
}

func (ase ASERecommendedOSSettings) Initialise() (Note, error) {
	actASE := ase
	newBlkSchedulers, err := actASE.BlockDeviceSchedulers.Inspect()
	if err != nil {
		return nil, err
	}
	actASE.BlockDeviceSchedulers = newBlkSchedulers.(param.BlockDeviceSchedulers)

	newBlkReq, err := actASE.BlockDeviceNrRequests.Inspect()
	if err != nil {
		return nil, err
	}
	actASE.BlockDeviceNrRequests = newBlkReq.(param.BlockDeviceNrRequests)
	return actASE, nil
}
func (ase ASERecommendedOSSettings) Optimise() (Note, error) {
	newASE := ase
	config, err := txtparser.ParseSysconfigFile(SYBASE_SYSCONFIG, false)
	if err != nil {
		return nil, err
	}
	sval := config.GetString("SYBASE_IO_SCHEDULER", "")
	switch sval {
	case "noop", "cfg", "deadline":
		//nothing to do
	default:
		fmt.Printf("wrong selection for SYBASE_IO_SCHEDULER in %s. Now set to 'noop'\n", SYBASE_SYSCONFIG)
		sval = "noop"
	}

	for blk := range newASE.BlockDeviceSchedulers.SchedulerChoice {
		newASE.BlockDeviceSchedulers.SchedulerChoice[blk] = sval
	}

	ival := config.GetInt("SYBASE_NRREQ", 0)
	if ival == 0 {
		fmt.Printf("nr_request set to '%s'\n", ival)
		ival = 1024
	}

	for blk := range newASE.BlockDeviceNrRequests.NrRequests {
		file := path.Join("block", blk, "queue", "nr_requests")
		tst_err := system.TestSysString(file, strconv.Itoa(ival))
		if tst_err != nil {
			fmt.Printf("Write error on file '%s'.\n Can't set nr_request to '%d', seems to large for the device. Leaving untouched.\n", file, ival)
		} else {
			newASE.BlockDeviceNrRequests.NrRequests[blk] = ival
		}
	}
	return newASE, nil
}
func (ase ASERecommendedOSSettings) Apply() error {
	err := ase.BlockDeviceSchedulers.Apply()
	if err != nil {
		return err
	}
	err = ase.BlockDeviceNrRequests.Apply()
	return err
}
