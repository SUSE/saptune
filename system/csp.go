package system

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// constant definitions
const (
	// Cloud Service Provider short and long names
	// microsoft-azure
	CSPAzure     = "azure"
	CSPAzureLong = "Microsoft Azure"
	// amazon-web-services
	CSPAWS     = "aws"
	CSPAWSLong = "Amazon Web Services"
	// GoogleCloud
	CSPGoogle     = "google"
	CSPGoogleLong = "Google Cloud Platform"
	// OracleCloud
	CSPOVM     = "ovm"
	CSPOVMLong = "Oracle Cloud"
	// Alibaba Cloud
	CSPAlibaba     = "alibaba"
	CSPAlibabaLong = "Alibaba Cloud"
	// IBM Cloud VPC (Not IBM Cloud Classic)
	CSPIBMVPC     = "ibmVPC"
	CSPIBMVPCLong = "IBM Cloud Virtual Server for VPC"
)

// Configuration file for cloud instance type definition
var CSPInstanceConfig = "/var/lib/saptune/config/cloud.json"

// CSPInstance defines the structure of the cloud instance configuration
type CSPInstance struct {
	Provider     string `json:"csp"`
	InstanceType string `json:"instance_type"`
	ManualSet    bool   `json:"manual_set"`
}

// metadataServerIP is the IPv4 address for the metadata server on Azure, AWS
// and GCP
var metadataServerIP = "169.254.169.254"

// CSPTimeout is the timeout for the Metadata Server query
// Set in the saptune main configuration file and changeable by customer
var CSPTimeout = 2

// CSPRetries is the number of retries for the Metadata Server query
// Set in the saptune main configuration file and changeable by customer
var CSPRetries = 1

// CSPDetectOnBoot defines when the cloud instance detection should run
// Set in the saptune main configuration file and changeable by customer
var CSPDetectOnBoot = "first"

// cloudDetectMarker is a marker file that the cloud detection had already run
var cloudDetectMarker = "/run/.saptune.cloud_detected"

// TCSP used to skip CSP tests
var TCSP = "skip"
var dmiDir = "/sys/class/dmi"

// CSP identifier
var isAzureCat = regexp.MustCompile(`.*(7783-7084-3265-9085-8269-3286-77|MSFT AZURE VM).*`)
var isAzure = regexp.MustCompile(`.*[mM]icrosoft [cC]orporation.*`)
var isAWS = regexp.MustCompile(`.*[aA]mazon.*`)
var isGoogle = regexp.MustCompile(`.*[gG]oogle.*`)
var isOVM = regexp.MustCompile(`.*OVM.*`)
var isAlibaba = regexp.MustCompile(`.*[aA]libaba.*`)
var isIBMVPCCat = regexp.MustCompile(`.*ibmcloud.*`)
var isIBMVPC = regexp.MustCompile(`.*IBM:Cloud Compute Server 1.0:.*`)

type manufacturerProviders struct {
	Manufacturer string
	Providers    map[*regexp.Regexp]string
}

var allManufacturerProviders = [...]manufacturerProviders{
	// dmidecode key files
	// /usr/sbin/dmidecode -s chassis-asset-tag
	{"/sys/class/dmi/id/chassis_asset_tag", map[*regexp.Regexp]string{isAzureCat: CSPAzure, isIBMVPCCat: CSPIBMVPC}},
	// /usr/sbin/dmidecode -s system-manufacturer
	{"/sys/class/dmi/id/sys_vendor", map[*regexp.Regexp]string{isAzure: CSPAzure, isAWS: CSPAWS, isGoogle: CSPGoogle, isAlibaba: CSPAlibaba}},
	// /usr/sbin/dmidecode -s board-vendor
	{"/sys/class/dmi/id/board_vendor", map[*regexp.Regexp]string{isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s bios-version
	{"/sys/class/dmi/id/bios_version", map[*regexp.Regexp]string{isAWS: CSPAWS, isGoogle: CSPGoogle, isOVM: CSPOVM}},
	// /usr/sbin/dmidecode -s bios-vendor
	{"/sys/class/dmi/id/bios_vendor", map[*regexp.Regexp]string{isGoogle: CSPGoogle, isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s system-version
	{"/sys/class/dmi/id/product_version", map[*regexp.Regexp]string{isAWS: CSPAWS}},
	// /usr/sbin/dmidecode -s chassis-manufacturer
	{"/sys/class/dmi/id/chassis_vendor", map[*regexp.Regexp]string{isIBMVPC: CSPIBMVPC}},
}

// GetDMIDecode
//func GetDMIDecode(key string) string {
// system_version = "dmidecode -s system-version"
//}

// GetCSP returns the identifier of the cloud service provider of the current
// running system or an empty string, if the system does not belong to a CSP
// use files in /sys/class/dmi/id/ instead of dmidecode command
func GetCSP() string {
	DebugLog("GetCSP")
	if TCSP != "skip" {
		return TCSP
	}
	cloudServiceProvider := ""
	getCloudServiceProvider := func(manufacturer string, providers map[*regexp.Regexp]string) string {
		if content, err := os.ReadFile(manufacturer); err == nil {
			for providerRegex, provider := range providers {
				matches := providerRegex.FindStringSubmatch(string(content))
				if len(matches) != 0 {
					return provider
				}
			}
		}
		return ""
	}

	if _, err := os.Stat(dmiDir); os.IsNotExist(err) {
		InfoLog("directory '%s' does not exist", dmiDir)
		return cloudServiceProvider
	}

	for _, mp := range allManufacturerProviders {
		if cloudServiceProvider == "" {
			cloudServiceProvider = getCloudServiceProvider(mp.Manufacturer, mp.Providers)
		} else {
			break
		}
	}
	return cloudServiceProvider
}

// setupCSPMeta returns the connection settings needed to query the CSP
// metadata server
func setupCSPMeta(vendor, request string, commParams CommParams) CommParams {
	DebugLog("setupCSPMeta - vendor is '%s', request is '%s', commParams is '%+v'", vendor, request, commParams)
	commParams.Server = metadataServerIP
	commParams.Timeout = CSPTimeout
	commParams.Retries = CSPRetries
	commParams.Canonical = true
	switch vendor {
	case "azure":
		commParams.IgnoreProxy = "yes"
		commParams.URL = "metadata/instance/compute/vmSize?api-version=" + request + "&format=text"
		commParams.Header["Metadata"] = "true"
		if request == "getAPIVersion" {
			commParams.URL = "metadata/versions"
		}
	case "aws":
		tlf := CSPTimeout*CSPRetries + 10
		commParams.IgnoreProxy = "no"
		//commParams.Canonical = false
		if request == "getToken" {
			commParams.URL = "latest/api/token"
			commParams.Header["X-aws-ec2-metadata-token-ttl-seconds"] = strconv.Itoa(tlf)
			commParams.Method = "PUT"
		} else {
			commParams.URL = "latest/meta-data/instance-type"
			commParams.Header["X-aws-ec2-metadata-token"] = request
			commParams.Method = "GET"
		}
	case "google":
		commParams.IgnoreProxy = "yes"
		commParams.URL = "/computeMetadata/v1/instance/machine-type"
		commParams.Header["Metadata-Flavor"] = "Google"
	}
	DebugLog("setupCSPMeta - return commParams as '%+v'", commParams)
	return commParams
}

// retrieveCloudMetadata retrieves the informations form the cloud metadata
// server
func retrieveCloudMetadata(csp string) string {
	DebugLog("retrieveCloudMetadata - cps is '%s'", csp)
	response := ""
	commParams := CommParamInit()
	switch csp {
	case "azure":
		// check API version
		commParams = setupCSPMeta(csp, "getAPIVersion", commParams)
		apiVersion := RetrieveAPIVersion(commParams)
		if apiVersion == "" {
			ErrorLog("Azure API version retrieval failed, instance detection not possible.")
			break
		}
		// retrieve instance type
		commParams = setupCSPMeta(csp, apiVersion, commParams)
		response = RetrievalRequest(commParams)
	case "aws":
		// request token
		commParams = setupCSPMeta(csp, "getToken", commParams)
		token := RetrievalRequest(commParams)
		if token == "" {
			ErrorLog("AWS metadata token retrieval failed, instance detection not possible.")
			break
		}
		// reset Header
		commParams.Header = make(map[string]string)
		// retrieve instance type
		commParams = setupCSPMeta(csp, token, commParams)
		response = RetrievalRequest(commParams)
	case "google":
		// retrieve instance type
		commParams = setupCSPMeta(csp, "", commParams)
		response = RetrievalRequest(commParams)
	default:
		response = "not supported"
	}
	DebugLog("retrieveCloudMetadata - return response as '%s'", response)
	return response
}

// retrieveCloudInstanceType retrieves the instance type from the metadata
// server
func retrieveCloudInstanceType(csp string) string {
	DebugLog("retrieveCloudInstanceType - cps is '%s'", csp)
	instanceType := ""
	response := retrieveCloudMetadata(csp)
	// The retrieval returns a string like
	// "projects/1234567890/machineTypes/n2-standard-2" or "t3.nano"
	// so split the response by '/' and grab the last element to get
	// the instance type name
	if response == "" {
		ErrorLog("instance detection response is empty, detection failed.")
		return ""
	}
	if response == "not supported" {
		InfoLog("instance detection for cloud service provider '%s' currently not supported.", csp)
		return ""
	}

	fields := strings.Split(response, "/")
	if len(fields) > 0 {
		instanceType = fields[len(fields)-1]
	}
	DebugLog("retrieveCloudInstanceType - return instanceType as '%s'", instanceType)
	return instanceType
}

// ResetCloudConfigFile resets content of file /var/lib/saptune/config/cloud.json
func ResetCloudConfigFile() error {
	DebugLog("ResetCloudConfigFile")
	cspInstance := CSPInstance{
		Provider:     GetCSP(),
		InstanceType: "",
		ManualSet:    false,
	}
	err := WriteCloudConfigFile(cspInstance)
	if err != nil {
		ErrorLog("Problems resetting cloud instance information in file '%s' - '%v'", CSPInstanceConfig, err)
	}
	DebugLog("ResetCloudConfigFile - return err as '%v'", err)
	return err
}

// WriteCloudConfigFile writes updated information to file
// /var/lib/saptune/config/cloud.json
func WriteCloudConfigFile(cspInstance CSPInstance) error {
	DebugLog("WriteCloudConfigFile - cspInstance is '%+v'", cspInstance)
	content, err := json.Marshal(cspInstance)
	if err == nil {
		err = os.WriteFile(CSPInstanceConfig, content, 0644)
	}
	DebugLog("WriteCloudConfigFile - return err as '%v'", err)
	return err
}

// readCloudConfigFile reads the cloud json file
// /var/lib/saptune/config/cloud.json
func readCloudConfigFile() (CSPInstance, error) {
	DebugLog("readCloudConfigFile")
	cspInstance := CSPInstance{
		Provider:     "",
		InstanceType: "",
		ManualSet:    false,
	}
	content, err := os.ReadFile(CSPInstanceConfig)
	if err != nil {
		ErrorLog("Failed to read file '%s' - '%v'.", CSPInstanceConfig, err)
		return cspInstance, err
	}
	if len(content) == 0 {
		err = ErrorLog("File '%s' is empty.", CSPInstanceConfig)
		return cspInstance, err
	}
	err = json.Unmarshal(content, &cspInstance)
	if err != nil {
		ErrorLog("Failed to read json content of file '%s' - '%v'.", CSPInstanceConfig, err)
		return cspInstance, err
	}
	DebugLog("readCloudConfigFile - returns cspInstance as '%+v', err as '%v'", cspInstance, err)
	return cspInstance, err
}

// GetCSPInstanceInfo reads the cloud instance type from file
// /var/lib/saptune/config/cloud.json
func GetCSPInstanceInfo(csp string) (CSPInstance, error) {
	DebugLog("GetCSPInstanceInfo - cps is '%s'", csp)
	instanceInfo, err := readCloudConfigFile()
	if err != nil {
		return instanceInfo, err
	}
	if instanceInfo.Provider != csp {
		InfoLog("Mishmash in file '%s' detected. The stored cloud provider '%s' does not match the current running one '%s'.", CSPInstanceConfig, instanceInfo.Provider, csp)
	}
	DebugLog("GetCSPInstanceInfo - returns instanceInfo as '%+v', err as '%v'", instanceInfo, err)
	return instanceInfo, err
}

// DetectCSPInstance is called early in main to check the instance type
// in /var/lib/saptune/config/cloud.json
func DetectCSPInstance(csp string) error {
	DebugLog("DetectCSPInstance - cps is '%s'", csp)
	// read /var/lib/saptune/config/cloud.json
	cspInstance, err := readCloudConfigFile()
	if err != nil {
		return err
	}
	if cspInstance.ManualSet {
		NoticeLog("cloud provider '%s' and instance type '%s' were set manually", cspInstance.Provider, cspInstance.InstanceType)
		return nil
	}
	// manual_set key is false
	if cspInstance.Provider == "" || cspInstance.InstanceType == "" {
		cspInstance.Provider = csp
		_, err = DetectAndStoreInstanceType(cspInstance)
		return err
	}

	switch CSPDetectOnBoot {
	case "always":
		err = updateCloudInstanceInfo(csp, cspInstance)
	case "first":
		_, err = os.Stat(cloudDetectMarker)
		if os.IsNotExist(err) && SystemIsRunning() {
			err = updateCloudInstanceInfo(csp, cspInstance)
			if err != nil {
				ErrorLog("Problems during update of cloud instance information - '%v'", err)
				break
			}
			// create the empty /run/.saptune.cloud_detected
			// as marker, that detection had already run
			marker, err := os.Create(cloudDetectMarker)
			if err != nil {
				ErrorLog("Problems during creation of cloud detection marker - '%v'", err)
			}
			defer marker.Close()
		}
	}
	DebugLog("DetectCSPInstance - return err as '%v'", err)
	return err
}

// DetectAndStoreInstanceType detects the instance type and writes the result to
// file /var/lib/saptune/config/cloud.json
func DetectAndStoreInstanceType(cspInstance CSPInstance) (CSPInstance, error) {
	DebugLog("DetectAndStoreInstanceType - cspInstance is '%+v'", cspInstance)
	cspInstance.InstanceType = retrieveCloudInstanceType(cspInstance.Provider)
	if cspInstance.InstanceType == "" {
		return cspInstance, ErrorLog("Instance detection failed")
	}
	err := WriteCloudConfigFile(cspInstance)
	if err != nil {
		ErrorLog("Problems writing cloud instance information to file '%s' - '%v'", CSPInstanceConfig, err)
	}
	DebugLog("DetectAndStoreInstanceType - return cspInstance as '%+v', err as '%v'", cspInstance, err)
	return cspInstance, err
}

// updateCloudInstanceInfo updates the instance informations in
// /var/lib/saptune/config/cloud.json
func updateCloudInstanceInfo(csp string, cspInstance CSPInstance) error {
	DebugLog("updateCloudInstanceInfo - csp is '%s', cspInstance is '%+v'", csp, cspInstance)
	instType := retrieveCloudInstanceType(csp)
	if instType == "" {
		return ErrorLog("Instance detection failed, no update of file '%s'", CSPInstanceConfig)
	}
	if csp != cspInstance.Provider {
		WarningLog("Stored cloud provider '%s' differs from the detected one '%s'.", cspInstance.Provider, csp)
	}
	if instType != cspInstance.InstanceType {
		WarningLog("Stored instance type '%s' differs from the detected one '%s'.", cspInstance.InstanceType, instType)
	}
	cspInstance.Provider = csp
	cspInstance.InstanceType = instType
	err := WriteCloudConfigFile(cspInstance)
	DebugLog("updateCloudInstanceInfo - returns err as '%v'", err)
	return err
}
