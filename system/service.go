package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetAvailServices returns a map of the available services of the system
func GetAvailServices() map[string]string {
	allServices := make(map[string]string)
	cmdArgs := []string{"--no-pager", "list-unit-files"}
	cmdOut, err := exec.Command(systemctlCmd, cmdArgs...).CombinedOutput()
	if err != nil {
		WarningLog("There was an error running external command %s %s: %v, output: %s", systemctlCmd, cmdArgs, err, cmdOut)
		return allServices
	}
	for _, line := range strings.Split(string(cmdOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		serv := strings.TrimSpace(fields[0])
		allServices[serv] = serv
	}
	return allServices
}

// GetServiceName returns the systemd service name for supported services
func GetServiceName(service string) string {
	serviceName := ""
	if services == nil || len(services) == 0 {
		services = GetAvailServices()
	}
	if _, ok := services[service]; ok {
		serviceName = service
	} else {
		serv := fmt.Sprintf("%s.service", service)
		if _, ok := services[serv]; ok {
			serviceName = serv
		}
	}
	if serviceName == "" {
		WarningLog("skipping unkown service '%s'", service)
	}
	return serviceName
}
