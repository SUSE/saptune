package system

import (
	"os/exec"
	"strings"
)

// HostnameCmd is the path to the hostname executable
var HostnameCmd = "/usr/bin/hostname"

// DnsdomainnameCmd is the path to the dnsdomainname executable
var DnsdomainnameCmd = "/usr/bin/dnsdomainname"

// GetHostname returns the output of 'hostname' if the exit code is 0
func GetHostname() (string, error) {
	out, err := exec.Command(HostnameCmd).CombinedOutput()
	if err != nil {
		ErrorLogNoStdErr("GetHostname - external command '%s' failed: %v, output: %s", HostnameCmd, err, strings.TrimSpace(string(out)))
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDNSDomainname returns the output of 'dnsdomainname' if the exit code is 0
func GetDNSDomainname() (string, error) {
	out, err := exec.Command(DnsdomainnameCmd).CombinedOutput()
	if err != nil {
		ErrorLogNoStdErr("GetDNSDomainname - external command '%s' failed: %v, output: %s", DnsdomainnameCmd, err, strings.TrimSpace(string(out)))
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetFQDN returns the output of 'hostname -f' if the exit code is 0
func GetFQDN() (string, error) {
	out, err := exec.Command(HostnameCmd, "-f").CombinedOutput()
	if err != nil {
		ErrorLogNoStdErr("GetFQDN - external command '%s -f' failed: %v, output: %s", HostnameCmd, err, strings.TrimSpace(string(out)))
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
