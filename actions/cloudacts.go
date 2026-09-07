package actions

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"io"
	"strings"
)

// CloudAction provides cloud actions like get instance type asm.
func CloudAction(writer io.Writer, actionName, actionType, typeValue string) {
	system.DebugLog("CloudAction - actionName is '%s', actionType is '%s', typeValue is '%s'", actionName, actionType, typeValue)
	// check command line syntax
	if system.CliArg(3) == "" {
		PrintHelpAndExit(writer, 1)
	}
	switch actionName {
	case "get":
		CloudActionGet(writer, actionType)
	case "set":
		if actionType == "cloud_detection_on_boot" {
			configureActionSetDetectionOnBoot(typeValue)
			_ = detectAction()
			system.ErrorExit("", 0)
		}
		CloudActionSet(writer, actionType, typeValue)
	case "clear":
		CloudActionClear(writer, actionType)
	case "detect":
		CloudActionDetect(writer, actionType)
	default:
		PrintHelpAndExit(writer, 1)
	}
}

// CloudActionGet prints the cloud information from
// /var/lib/saptune/config/cloud.json
func CloudActionGet(writer io.Writer, actionType string) {
	system.DebugLog("CloudActionGet - actionType is '%s'", actionType)
	csp := system.GetCSP()
	if csp == "" {
		system.NoticeLog("Attention - not running on cloud!")
	}
	cspInstance, err := system.GetCSPInstanceInfo(csp)
	if err != nil {
		system.ErrorLog("Failed to retrieve cloud information from file '%s': '%v'", system.CSPInstanceConfig, err)
		system.ErrorExit("", 128)
	}
	switch actionType {
	case "instance_type":
		fmt.Fprintf(writer, "Cloud Instance Type: %s\n", cspInstance.InstanceType)
		if cspInstance.ManualSet {
			fmt.Fprintf(writer, "was set manually.\n")
		} else {
			fmt.Fprintf(writer, "has been determined by detection.\n")
		}
	default:
		PrintHelpAndExit(writer, 1)
	}
	system.DebugLog("CloudActionGet - return")
}

// CloudActionSet sets/updates the cloud information in /var/lib/saptune/config/cloud.json
func CloudActionSet(writer io.Writer, actionType, value string) {
	system.DebugLog("CloudActionSet - actionType is '%s', value is '%s'", actionType, value)
	switch actionType {
	case "instance_type":
		fields := strings.Split(value, "%")
		if len(fields) <= 1 {
			system.ErrorExit("Wrong input value '%s' for cloud vendor and instance type. Please check.", value)
		}
		if fields[0] == "" || fields[1] == "" {
			system.ErrorExit("Wrong input value '%s' for cloud vendor and instance type. Please check.", value)
		}
		cspInstance := system.CSPInstance{
			Provider:     fields[0],
			InstanceType: fields[1],
			ManualSet:    true,
		}
		err := system.WriteCloudConfigFile(cspInstance)
		if err != nil {
			system.ErrorExit("Problems writing cloud instance information to file - '%v'", err)
		}
	default:
		PrintHelpAndExit(writer, 1)
	}
	system.DebugLog("CloudActionSet - return")
}

// CloudActionClear clears the cloud information in /var/lib/saptune/config/cloud.json
func CloudActionClear(writer io.Writer, actionType string) {
	system.DebugLog("CloudActionClear - actionType is '%s'", actionType)
	switch actionType {
	case "instance_type":
		system.InfoLog("Clear cloud information in file '%s'", system.CSPInstanceConfig)
		err := system.ResetCloudConfigFile()
		if err != nil {
			system.ErrorExit("Failed to retrieve cloud information from file '%s': '%v'", system.CSPInstanceConfig, err)
		}
	default:
		PrintHelpAndExit(writer, 1)
	}
	system.DebugLog("CloudActionClear - return")
}

// CloudActionDetect runs the cloud instance detection, stores the result in
// /var/lib/saptune/config/cloud.json and prints it to stdout
func CloudActionDetect(writer io.Writer, actionType string) {
	system.DebugLog("CloudActionDetect - actionType is '%s'", actionType)
	if actionType != "instance_type" {
		PrintHelpAndExit(writer, 1)
	}
	instanceType := detectAction()
	fmt.Fprintf(writer, "Cloud Instance Type: %s\n", instanceType)
	system.DebugLog("CloudActionDetect - return")
}

// detectAction runs the cloud instance detection and stores the result in
// /var/lib/saptune/config/cloud.json
func detectAction() string {
	cspInstance := system.CSPInstance{
		Provider:     system.GetCSP(),
		InstanceType: "",
		ManualSet:    false,
	}
	if cspInstance.Provider == "" {
		system.ErrorExit("Not on cloud, instance detection not possible")
	}
	cspInst, err := system.DetectAndStoreInstanceType(cspInstance)
	if cspInst.InstanceType == "" {
		system.ErrorExit("Cloud instance detection failed.", 128)
	}
	if err != nil {
		system.ErrorExit("Problems storing cloud instance type - '%v'", err, 128)
	}
	system.DebugLog("CloudActionDetect - return cspInst.InstanceType as '%s'")
	return cspInst.InstanceType
}
