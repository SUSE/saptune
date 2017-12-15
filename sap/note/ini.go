package note

import (
	"fmt"
	"github.com/HouzuoGuo/saptune/sap"
	"github.com/HouzuoGuo/saptune/sap/param"
	"github.com/HouzuoGuo/saptune/system"
	"github.com/HouzuoGuo/saptune/txtparser"
	"log"
	"strconv"
	"strings"
)

const (
	INISectionSysctl    = "sysctl"
	INISectionVM        = "vm"
	INISectionBlock     = "block"
	INISectionLimits    = "limits"
	SysKernelTHPEnabled = "kernel/mm/transparent_hugepage/enabled"
	SysKSMRun           = "kernel/mm/ksm/run"
)

// Tuning options composed by a third party vendor.

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
type BlockDeviceQueue struct {
	BlockDeviceSchedulers param.BlockDeviceSchedulers
	BlockDeviceNrRequests param.BlockDeviceNrRequests
}

func GetBlockVal(key string) (string, error) {
	newQueue := make(map[string]string)
	newReq := make(map[string]int)
	ret_val := ""
	switch key {
	case "IO_SCHEDULER":
		newIOQ, err := BlockDeviceQueue{}.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return "", err
		}
		newQueue = newIOQ.(param.BlockDeviceSchedulers).SchedulerChoice
		for k, v := range newQueue {
			ret_val = ret_val + fmt.Sprintf("%s@%s ", k, v)
		}
	case "NRREQ":
		newNrR, err := BlockDeviceQueue{}.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return "", err
		}
		newReq = newNrR.(param.BlockDeviceNrRequests).NrRequests
		for k, v := range newReq {
			ret_val = ret_val + fmt.Sprintf("%s@%s ", k, strconv.Itoa(v))
		}
	}
	return ret_val, nil
}

func OptBlkVal(parameter, act_value, cfg_value string) string {
	sval := cfg_value
	val := act_value
	ret_val := ""
	switch parameter {
	case "IO_SCHEDULER":
		sval = strings.ToLower(cfg_value)
	case "NRREQ":
		if sval == "0" {
			sval = "1024"
		}
	}
	for _, entry := range strings.Fields(val) {
		fields := strings.Split(entry, "@")
		ret_val = ret_val + fmt.Sprintf("%s@%s ", fields[0], sval)
	}

	return ret_val
}

func SetBlkVal(key, value string) error {
	var err error

	switch key {
	case "IO_SCHEDULER":
		setIOQ, err := BlockDeviceQueue{}.BlockDeviceSchedulers.Inspect()
		if err != nil {
			return err
		}

		for _, entry := range strings.Fields(value) {
			fields := strings.Split(entry, "@")
			setIOQ.(param.BlockDeviceSchedulers).SchedulerChoice[fields[0]] = fields[1]
		}
		err = setIOQ.(param.BlockDeviceSchedulers).Apply()
		if err != nil {
			return err
		}
	case "NRREQ":
		setNrR, err := BlockDeviceQueue{}.BlockDeviceNrRequests.Inspect()
		if err != nil {
			return err
		}

		for _, entry := range strings.Fields(value) {
			fields := strings.Split(entry, "@")
			NrR, _ := strconv.Atoi(fields[1])
			setNrR.(param.BlockDeviceNrRequests).NrRequests[fields[0]] = NrR
		}
		err = setNrR.(param.BlockDeviceNrRequests).Apply()
		if err != nil {
			return err
		}
	}
	return err
}

// section [limits]
func GetLimitsVal(key string) (string, error) {
	// Find out current memlock limits
	LimitMemlock := "0"
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return "", err
	}
	switch key {
	case "MEMLOCK_HARD":
		LimitMemlock, _ = secLimits.Get("sybase", "hard", "memlock")
	case "MEMLOCK_SOFT":
		LimitMemlock, _ = secLimits.Get("sybase", "soft", "memlock")

	}
	return LimitMemlock, nil
}

func OptLimitsVal(act_value, cfg_value string) string {
	LimitMemlock := ""
	switch act_value {
	case "unlimited", "infinity", "-1":
		LimitMemlock = act_value
	default:
		LimMemlock, _ := strconv.Atoi(act_value)
		switch cfg_value {
		case "unlimited", "infinity", "-1":
			LimitMemlock = cfg_value
		case "0":
			//RAM in KB - 10%
			memlock := system.GetMainMemSizeMB()*1024 - (system.GetMainMemSizeMB() * 1024 * 10 / 100)
			LimitMemlock = strconv.Itoa(param.MaxI(LimMemlock, int(memlock)))
		default:
			LimLockCFG, _ := strconv.Atoi(cfg_value)
			LimitMemlock = strconv.Itoa(param.MaxI(LimMemlock, LimLockCFG))
		}
	}
	return LimitMemlock
}

func SetLimitsVal(key, value string) error {
	secLimits, err := system.ParseSecLimitsFile()
	if err != nil {
		return err
	}
	switch key {
	case "MEMLOCK_HARD":
		secLimits.Set("sybase", "hard", "memlock", value)
	case "MEMLOCK_SOFT":
		secLimits.Set("sybase", "soft", "memlock", value)
	}
	err = secLimits.Apply()
	return err
}

// section [vm]
// Manipulate /sys/kernel/mm switches.
func GetVmVal(parameter string) string {
	var val string
	switch parameter {
	case "INI_THP":
		val, _ = system.GetSysChoice(SysKernelTHPEnabled)
	}
	return val
}

func OptVmVal(parameter, act_value, cfg_value string) string {
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
			vend.SysctlParams[param.Key], _ = system.GetSysctlString(param.Key)
		case INISectionVM:
			vend.SysctlParams[param.Key] = GetVmVal(param.Key)
		case INISectionBlock:
			vend.SysctlParams[param.Key], _ = GetBlockVal(param.Key)
		case INISectionLimits:
			vend.SysctlParams[param.Key], _ = GetLimitsVal(param.Key)
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
			vend.SysctlParams[param.Key] = OptBlkVal(param.Key, vend.SysctlParams[param.Key], param.Value)
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
	errs := make([]error, 0, 0)
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
			errs = append(errs, system.SetSysctlString(param.Key, vend.SysctlParams[param.Key]))
		case INISectionVM:
			errs = append(errs, system.SetSysString(SysKernelTHPEnabled, vend.SysctlParams[param.Key]))
		case INISectionBlock:
			errs = append(errs, SetBlkVal(param.Key, vend.SysctlParams[param.Key]))
		case INISectionLimits:
			errs = append(errs, SetLimitsVal(param.Key, vend.SysctlParams[param.Key]))
		default:
			// saptune does not yet understand settings outside of [sysctl] section
			log.Printf("3rdPartyTuningOption %s: skip unknown section %s", vend.ConfFilePath, param.Section)
			continue
		}
	}
	err = sap.PrintErrors(errs)
	return err
}
