package system

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetHostname(t *testing.T) {
	origCmd := HostnameCmd
	defer func() { HostnameCmd = origCmd }()

	// Test with actual system command
	host, err := GetHostname()
	if err != nil {
		t.Logf("GetHostname returned error on current system: %v", err)
	} else {
		if host == "" {
			t.Errorf("GetHostname returned empty hostname with nil error")
		}
	}

	// Test with non-zero exit code (e.g. /usr/bin/false)
	HostnameCmd = "/usr/bin/false"
	_, err = GetHostname()
	if err == nil {
		t.Errorf("GetHostname should have returned an error when command exits non-zero")
	}

	// Test with mock script returning exit code 0
	tmpDir := t.TempDir()
	mockScript := filepath.Join(tmpDir, "mock_hostname")
	scriptContent := "#!/bin/sh\necho 'my-test-host.local'\nexit 0\n"
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}

	HostnameCmd = mockScript
	host, err = GetHostname()
	if err != nil {
		t.Errorf("GetHostname unexpected error: %v", err)
	}
	if host != "my-test-host.local" {
		t.Errorf("GetHostname expected 'my-test-host.local', got '%s'", host)
	}
}

func TestGetDNSDomainname(t *testing.T) {
	origCmd := DnsdomainnameCmd
	defer func() { DnsdomainnameCmd = origCmd }()

	// Test with non-zero exit code (e.g. /usr/bin/false)
	DnsdomainnameCmd = "/usr/bin/false"
	_, err := GetDNSDomainname()
	if err == nil {
		t.Errorf("GetDNSDomainname should have returned an error when command exits non-zero")
	}

	// Test with mock script returning exit code 0
	tmpDir := t.TempDir()
	mockScript := filepath.Join(tmpDir, "mock_dnsdomainname")
	scriptContent := "#!/bin/sh\necho 'corp.example.com'\nexit 0\n"
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}

	DnsdomainnameCmd = mockScript
	domain, err := GetDNSDomainname()
	if err != nil {
		t.Errorf("GetDNSDomainname unexpected error: %v", err)
	}
	if domain != "corp.example.com" {
		t.Errorf("GetDNSDomainname expected 'corp.example.com', got '%s'", domain)
	}
}

func TestGetFQDN(t *testing.T) {
	origCmd := HostnameCmd
	defer func() { HostnameCmd = origCmd }()

	// Test with non-zero exit code (e.g. /usr/bin/false)
	HostnameCmd = "/usr/bin/false"
	_, err := GetFQDN()
	if err == nil {
		t.Errorf("GetFQDN should have returned an error when command exits non-zero")
	}

	// Test with mock script checking for -f argument
	tmpDir := t.TempDir()
	mockScript := filepath.Join(tmpDir, "mock_hostname_fqdn")
	scriptContent := "#!/bin/sh\nif [ \"$1\" = \"-f\" ]; then\n  echo 'node01.corp.example.com'\n  exit 0\nelse\n  exit 1\nfi\n"
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}

	HostnameCmd = mockScript
	fqdn, err := GetFQDN()
	if err != nil {
		t.Errorf("GetFQDN unexpected error: %v", err)
	}
	if fqdn != "node01.corp.example.com" {
		t.Errorf("GetFQDN expected 'node01.corp.example.com', got '%s'", fqdn)
	}
}
