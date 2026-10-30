package txtparser

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"testing"
)

func TestChkOsTags(t *testing.T) {
	tag := "15-*"
	secFields := []string{"rpm", "os=15-*", "arch=amd64"}

	ret := chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	_ = system.CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr12"), "/etc/os-release")
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	_ = system.CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/osr16.2"), "/etc/os-release")
	tag = "16.2"
	secFields = []string{"rpm", "os=16.2", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "19.0"
	secFields = []string{"rpm", "os=19.0", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[0,2,4-5,7-]"
	secFields = []string{"rpm", "os=16.[0,2,4-5,7-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[-2,4-5,7-]"
	secFields = []string{"rpm", "os=16.[-2,4-5,7-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[0-1,2-]"
	secFields = []string{"rpm", "os=16.[0-1,2-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[0-2,4-5,7-]"
	secFields = []string{"rpm", "os=16.[0-2,4-5,7-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[-6]"
	secFields = []string{"rpm", "os=16.[-6]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if !ret {
		t.Error("not matching os version")
	}
	tag = "16.[02]"
	secFields = []string{"rpm", "os=16.[02]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "19.[2]"
	secFields = []string{"rpm", "os=19.[2]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[-]"
	secFields = []string{"rpm", "os=16.[-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[4-]"
	secFields = []string{"rpm", "os=16.[4-]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	tag = "16.[0+2]"
	secFields = []string{"rpm", "os=16.[0+2]", "arch=amd64"}
	ret = chkOsTags(tag, secFields)
	if ret {
		t.Error("matching os version, but shouldn't")
	}
	_ = system.CopyFile("/etc/os-release_OrG", "/etc/os-release")
}

func TestChkHWTags(t *testing.T) {
	system.DmiID = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")
	secFields := []string{"sysctl", "vendor=SUSE", "arch=amd64"}
	info := "vendor"
	tag := "SUSE"
	ret := chkHWTags(info, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content of %s file", tag, info)
	}

	secFields = []string{"sysctl", "model=SUSE saptune", "arch=amd64"}
	info = "model"
	tag = "SUSE saptune"
	ret = chkHWTags(info, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content of %s file", tag, info)
	}

	tag = "SE sap"
	ret = chkHWTags(info, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content of %s file", tag, info)
	}

	tag = "hugo"
	ret = chkHWTags(info, tag, secFields)
	if ret {
		t.Errorf("tag '%s' matches content of %s file, but shouldn't", tag, info)
	}

	os.Rename(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name"), path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name_OrG"))
	tag = "SUSE saptune"
	ret = chkHWTags(info, tag, secFields)
	if ret {
		t.Errorf("tag '%s' matches content of %s file, but shouldn't", tag, info)
	}
	os.Rename(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name_OrG"), path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name"))

	system.DmiID = "/sys/class/dmi/id"
}

func TestChkOtherTags(t *testing.T) {
	system.DmiID = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata")
	secFields := []string{"sysctl", "product_name=SUSE", "arch=amd64"}
	file := "product_name"
	tag := "SUSE"

	ret := chkOtherTags(file, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content in '%+s'", tag, file)
	}

	tag = "saptune"
	ret = chkOtherTags(file, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content in '%+s'", tag, file)
	}

	tag = "SE sap"
	ret = chkOtherTags(file, tag, secFields)
	if !ret {
		t.Errorf("tag '%s' does not match content in '%+s'", tag, file)
	}

	tag = "hugo"
	ret = chkOtherTags(file, tag, secFields)
	if ret {
		t.Errorf("tag '%s' does match content in '%+s', but shouldn't", tag, file)
	}
	system.DmiID = "/sys/class/dmi/id"
}

func TestIsTagAvail(t *testing.T) {
	secFields := []string{"block", "blkvendor=HUGO", "blkmodel=EGON"}
	tag := "blkvendor"
	if !isTagAvail(tag, secFields) {
		t.Errorf("tag '%s' is expected to be available, but is repoted as not available\n", tag)
	}
	tag = "blkmodel"
	if !isTagAvail(tag, secFields) {
		t.Errorf("tag '%s' is expected to be available, but is repoted as not available\n", tag)
	}
	tag = "blkpat"
	if isTagAvail(tag, secFields) {
		t.Errorf("tag '%s' is expected to be NOT available, but is repoted as available\n", tag)
	}
	tag = "blkvendor"
	secFields = []string{"block", "blkvendor="}
	if !isTagAvail(tag, secFields) {
		t.Error("expected 'true', because of correct syntax, but got 'false'")
	}
	secFields = []string{"block", "blkvendor"}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	secFields = []string{"block", "=EGON"}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	secFields = []string{"block", "="}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	secFields = []string{"block", ""}
	if isTagAvail(tag, secFields) {
		t.Error("expected 'false', because of wrong syntax, but got 'true'")
	}
	// meminfo tag in isTagAvail
	secFields = []string{"sysctl", "meminfo%MemTotal>=2147483648"}
	if !isTagAvail("meminfo", secFields) {
		t.Error("expected 'true' for isTagAvail('meminfo'), but got 'false'")
	}
	secFields = []string{"block", "meminfo%MemTotal>=2147483648", "blkvendor=HUGO"}
	if !isTagAvail("blkvendor", secFields) {
		t.Error("expected 'true' for isTagAvail('blkvendor') following meminfo tag, but got 'false'")
	}
}

func TestChkMeminfoTags(t *testing.T) {
	origProc := system.ProcMeminfo
	defer func() { system.ProcMeminfo = origProc }()

	tmpFile, err := os.CreateTemp("", "meminfo_test_tag")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `MemTotal:       2147483648 kB
MemFree:         1048576 kB
MemAvailable:   1500000 kB
Active(anon):      50000 kB
HugePages_Total:       4
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	system.ProcMeminfo = tmpFile.Name()

	secFields := []string{"sysctl"}

	// Valid matching cases for all operators
	validMatches := []string{
		"meminfo%MemTotal==2147483648",
		"meminfo%MemTotal!=1000",
		"meminfo%MemTotal>=2147483648",
		"meminfo%MemTotal>=2000000000",
		"meminfo%MemTotal>2000000000",
		"meminfo%MemTotal<=2147483648",
		"meminfo%MemTotal<=3000000000",
		"meminfo%MemTotal<3000000000",
		"meminfo%MemFree<=1048576",
		"meminfo%MemFree<2000000",
		"meminfo%Active(anon)==50000",
		"meminfo%HugePages_Total==4",
		"meminfo%HugePages_Total>0",
		"meminfo%MemTotal >= 2147483648",
		"meminfo%MemTotal == 2147483648",
		"meminfo%MemFree < 2000000",
	}
	for _, tag := range validMatches {
		if !chkMeminfoTags(tag, append(secFields, tag)) {
			t.Errorf("expected tag '%s' to match, but got false", tag)
		}
	}

	// Valid non-matching cases
	nonMatches := []string{
		"meminfo%MemTotal==123",
		"meminfo%MemTotal!=2147483648",
		"meminfo%MemTotal>2147483648",
		"meminfo%MemTotal<2147483648",
		"meminfo%MemFree>=2000000",
		"meminfo%MemFree<=500000",
		"meminfo%HugePages_Total==0",
		"meminfo%HugePages_Total<2",
		"meminfo%MemTotal <= 100",
	}
	for _, tag := range nonMatches {
		if chkMeminfoTags(tag, append(secFields, tag)) {
			t.Errorf("expected tag '%s' to NOT match, but got true", tag)
		}
	}

	// Unknown key
	if chkMeminfoTags("meminfo%NoSuchKey>=100", append(secFields, "meminfo%NoSuchKey>=100")) {
		t.Errorf("expected tag with non-existent key to return false")
	}

	// Invalid syntax cases
	invalidSyntax := []string{
		"meminfo%MemTotal",
		"meminfo%MemTotal=2147483648",
		"meminfo%MemTotal=~2147483648",
		"meminfo%MemTotal>=abc",
		"meminfo%>=100",
		"meminfo%",
		"meminfo%MemTotal>>100",
	}
	for _, tag := range invalidSyntax {
		if chkMeminfoTags(tag, append(secFields, tag)) {
			t.Errorf("expected invalid tag '%s' to return false, but got true", tag)
		}
	}
}

func TestChkSecTagsWithMeminfo(t *testing.T) {
	origProc := system.ProcMeminfo
	defer func() { system.ProcMeminfo = origProc }()

	tmpFile, err := os.CreateTemp("", "meminfo_test_sec")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `MemTotal:       2147483648 kB
MemFree:         1048576 kB
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	system.ProcMeminfo = tmpFile.Name()

	// Section matching meminfo tag
	ok, _ := chkSecTags([]string{"sysctl", "meminfo%MemTotal>=2147483648"}, nil)
	if !ok {
		t.Errorf("expected chkSecTags with matching meminfo tag to be true")
	}

	// Section failing meminfo tag
	ok, _ = chkSecTags([]string{"sysctl", "meminfo%MemTotal<1000"}, nil)
	if ok {
		t.Errorf("expected chkSecTags with non-matching meminfo tag to be false")
	}

	// Section with multiple meminfo tags: one true, one false
	ok, _ = chkSecTags([]string{"sysctl", "meminfo%MemTotal>=2147483648", "meminfo%MemFree<500"}, nil)
	if ok {
		t.Errorf("expected chkSecTags with one failing tag to be false")
	}

	// Section with multiple meminfo tags: both true
	ok, _ = chkSecTags([]string{"sysctl", "meminfo%MemTotal>=2147483648", "meminfo%MemFree>500"}, nil)
	if !ok {
		t.Errorf("expected chkSecTags with both matching tags to be true")
	}
}

func TestParseINIWithMeminfo(t *testing.T) {
	origProc := system.ProcMeminfo
	defer func() { system.ProcMeminfo = origProc }()

	tmpMem, err := os.CreateTemp("", "meminfo_ini_test")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpMem.Name())

	contentMem := `MemTotal:       2147483648 kB
MemFree:         1048576 kB
`
	if _, err := tmpMem.WriteString(contentMem); err != nil {
		t.Errorf("failed to write temp file: %v", err)
	}
	tmpMem.Close()

	system.ProcMeminfo = tmpMem.Name()

	iniContent := `
[sysctl:meminfo%MemTotal>=2147483648]
kernel.shmmax = 123456789

[sysctl:meminfo%MemTotal<1000000]
kernel.shmall = 999999
`
	iniData := ParseINI(iniContent)

	sysctlMap, exists := iniData.KeyValue["sysctl"]
	if !exists {
		t.Errorf("expected 'sysctl' section to exist in parsed INI")
	}

	if entry, ok := sysctlMap["kernel.shmmax"]; !ok || entry.Value != "123456789" {
		t.Errorf("expected kernel.shmmax to be '123456789', got '%+v'", entry)
	}

	if _, ok := sysctlMap["kernel.shmall"]; ok {
		t.Errorf("kernel.shmall should have been skipped due to non-matching meminfo tag")
	}
}

func TestChkHostTags(t *testing.T) {
	// chkHostnameTags checks
	origCmd := system.HostnameCmd
	defer func() { system.HostnameCmd = origCmd }()

	tmpDir := t.TempDir()
	mockScript := path.Join(tmpDir, "mock_hostname")
	scriptContent := "#!/bin/sh\necho 'sap-hana-db01'\nexit 0\n"
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}
	system.HostnameCmd = mockScript

	secFields := []string{"sysctl", "hostname=sap-.*"}

	// Matching regex
	if !chkHostTags("hostname", "sap-.*", secFields) {
		t.Errorf("expected 'sap-.*' to match 'sap-hana-db01'")
	}
	if !chkHostTags("hostname", "^sap-hana-db01$", secFields) {
		t.Errorf("expected '^sap-hana-db01$' to match 'sap-hana-db01'")
	}
	if !chkHostTags("hostname", "db01", secFields) {
		t.Errorf("expected 'db01' to match 'sap-hana-db01'")
	}

	// Non-matching regex
	if chkHostTags("hostname", "^db02$", secFields) {
		t.Errorf("expected '^db02$' not to match 'sap-hana-db01'")
	}

	// Invalid regex
	if chkHostTags("hostname", "[a-z", secFields) {
		t.Errorf("expected invalid regex to return false")
	}

	// Non-zero exit code
	system.HostnameCmd = "/usr/bin/false"
	if chkHostTags("hostname", ".*", secFields) {
		t.Errorf("expected non-zero exit code to return false")
	}

	// chkDNSDomainnameTags checks
	origCmd = system.DnsdomainnameCmd
	defer func() { system.DnsdomainnameCmd = origCmd }()

	tmpDir = t.TempDir()
	mockScript = path.Join(tmpDir, "mock_dnsdomainname")
	scriptContent = "#!/bin/sh\necho 'corp.sap.internal'\nexit 0\n"
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}
	system.DnsdomainnameCmd = mockScript

	secFields = []string{"sysctl", "dnsdomainname=.*sap\\.internal"}

	// Matching regex
	if !chkHostTags("dnsdomainname", ".*\\.sap\\.internal$", secFields) {
		t.Errorf("expected '.*\\.sap\\.internal$' to match 'corp.sap.internal'")
	}
	if !chkHostTags("dnsdomainname", "^corp\\.", secFields) {
		t.Errorf("expected '^corp\\.' to match 'corp.sap.internal'")
	}

	// Non-matching regex
	if chkHostTags("dnsdomainname", "example\\.com", secFields) {
		t.Errorf("expected 'example\\.com' not to match 'corp.sap.internal'")
	}

	// Invalid regex
	if chkHostTags("dnsdomainname", "(", secFields) {
		t.Errorf("expected invalid regex to return false")
	}

	// Non-zero exit code
	system.DnsdomainnameCmd = "/usr/bin/false"
	if chkHostTags("dnsdomainname", ".*", secFields) {
		t.Errorf("expected non-zero exit code to return false")
	}

	// chkFqdnTags checks
	origCmd = system.HostnameCmd
	defer func() { system.HostnameCmd = origCmd }()

	tmpDir = t.TempDir()
	mockScript = path.Join(tmpDir, "mock_fqdn")
	scriptContent = "#!/bin/sh\nif [ \"$1\" = \"-f\" ]; then\n  echo 'sap-app01.prod.acme.com'\n  exit 0\nelse\n  exit 1\nfi\n"
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}
	system.HostnameCmd = mockScript

	secFields = []string{"sysctl", "fqdn=.*\\.prod\\.acme\\.com"}

	// Matching regex
	if !chkHostTags("fqdn", "sap-app\\d+\\.prod\\.acme\\.com", secFields) {
		t.Errorf("expected 'sap-app\\d+\\.prod\\.acme\\.com' to match 'sap-app01.prod.acme.com'")
	}
	if !chkHostTags("fqdn", ".*\\.acme\\.com$", secFields) {
		t.Errorf("expected '.*\\.acme\\.com$' to match 'sap-app01.prod.acme.com'")
	}

	// Non-matching regex
	if chkHostTags("fqdn", "dev\\.acme\\.com", secFields) {
		t.Errorf("expected 'dev\\.acme\\.com' not to match 'sap-app01.prod.acme.com'")
	}

	// Invalid regex
	if chkHostTags("fqdn", "[", secFields) {
		t.Errorf("expected invalid regex to return false")
	}

	// Non-zero exit code
	system.HostnameCmd = "/usr/bin/false"
	if chkHostTags("fqdn", ".*", secFields) {
		t.Errorf("expected non-zero exit code to return false")
	}
}

func TestChkSecTagsHost(t *testing.T) {
	origHostnameCmd := system.HostnameCmd
	origDnsdomainnameCmd := system.DnsdomainnameCmd
	defer func() {
		system.HostnameCmd = origHostnameCmd
		system.DnsdomainnameCmd = origDnsdomainnameCmd
	}()

	tmpDir := t.TempDir()
	mockHostname := path.Join(tmpDir, "mock_hostname")
	scriptHostname := "#!/bin/sh\nif [ \"$1\" = \"-f\" ]; then\n  echo 'hana01.mycorp.internal'\n  exit 0\nfi\necho 'hana01'\nexit 0\n"
	if err := os.WriteFile(mockHostname, []byte(scriptHostname), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}
	mockDomain := path.Join(tmpDir, "mock_domain")
	scriptDomain := "#!/bin/sh\necho 'mycorp.internal'\nexit 0\n"
	if err := os.WriteFile(mockDomain, []byte(scriptDomain), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}

	system.HostnameCmd = mockHostname
	system.DnsdomainnameCmd = mockDomain

	// Valid hostname tag
	ok, _ := chkSecTags([]string{"sysctl", "hostname=^hana\\d+$"}, []string{})
	if !ok {
		t.Errorf("expected chkSecTags with hostname tag to be true")
	}

	// Valid dnsdomainname tag
	ok, _ = chkSecTags([]string{"sysctl", "dnsdomainname=^mycorp\\.internal$"}, []string{})
	if !ok {
		t.Errorf("expected chkSecTags with dnsdomainname tag to be true")
	}

	// Valid fqdn tag
	ok, _ = chkSecTags([]string{"sysctl", "fqdn=^hana01\\.mycorp\\.internal$"}, []string{})
	if !ok {
		t.Errorf("expected chkSecTags with fqdn tag to be true")
	}

	// Combined all three tags
	ok, _ = chkSecTags([]string{"sysctl", "hostname=hana.*", "dnsdomainname=mycorp.*", "fqdn=.*\\.internal"}, []string{})
	if !ok {
		t.Errorf("expected chkSecTags with all three host tags to be true")
	}

	// Combined with one non-matching tag
	ok, _ = chkSecTags([]string{"sysctl", "hostname=hana.*", "dnsdomainname=wrongdomain"}, []string{})
	if ok {
		t.Errorf("expected chkSecTags to fail when one tag does not match")
	}
}

func TestParseINIHostTags(t *testing.T) {
	origHostnameCmd := system.HostnameCmd
	origDnsdomainnameCmd := system.DnsdomainnameCmd
	defer func() {
		system.HostnameCmd = origHostnameCmd
		system.DnsdomainnameCmd = origDnsdomainnameCmd
	}()

	tmpDir := t.TempDir()
	mockHostname := path.Join(tmpDir, "mock_hostname")
	scriptHostname := "#!/bin/sh\nif [ \"$1\" = \"-f\" ]; then\n  echo 'hana01.mycorp.internal'\n  exit 0\nfi\necho 'hana01'\nexit 0\n"
	if err := os.WriteFile(mockHostname, []byte(scriptHostname), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}
	mockDomain := path.Join(tmpDir, "mock_domain")
	scriptDomain := "#!/bin/sh\necho 'mycorp.internal'\nexit 0\n"
	if err := os.WriteFile(mockDomain, []byte(scriptDomain), 0755); err != nil {
		t.Errorf("failed to write mock script: %v", err)
	}

	system.HostnameCmd = mockHostname
	system.DnsdomainnameCmd = mockDomain

	iniContent := `
[sysctl:hostname=^hana\d+$]
kernel.sem = 1250 256000 100 8192

[sysctl:hostname=^wronghost$]
fs.file-max = 6815744

[sys:dnsdomainname=^mycorp\.internal$]
net.ipv4.tcp_slow_start_after_idle = 1

[sys:dnsdomainname=^othercorp\.internal$]
net.ipv4.tcp_slow_start_after_idle = 0

[vm:fqdn=^hana01\.mycorp\.internal$]
vm.max_map_count = 2147483647

[vm:fqdn=^wrong\.fqdn$]
net.ipv4.tcp_tw_reuse = 1
`

	parsed := ParseINI(iniContent)
	if parsed == nil {
		t.Errorf("ParseINI returned nil")
	}

	sysctlEntries, exists := parsed.KeyValue["sysctl"]
	if !exists {
		t.Errorf("expected 'sysctl' section in parsed INI")
	}
	sysEntries, exists := parsed.KeyValue["sys"]
	if !exists {
		t.Errorf("expected 'sys' section in parsed INI")
	}
	vmEntries, exists := parsed.KeyValue["vm"]
	if !exists {
		t.Errorf("expected 'vm' section in parsed INI")
	}

	// Should match and exist
	if _, ok := sysctlEntries["kernel.sem"]; !ok {
		t.Errorf("expected 'kernel.sem' to be parsed for matching hostname")
	}
	if _, ok := sysEntries["sys:net.ipv4.tcp_slow_start_after_idle"]; !ok {
		t.Errorf("expected 'sys:net.ipv4.tcp_slow_start_after_idle' to be parsed for matching dnsdomainname")
	}
	if _, ok := vmEntries["vm.max_map_count"]; !ok {
		t.Errorf("expected 'vm.max_map_count' to be parsed for matching fqdn")
	}

	// Should not match and not exist
	if _, ok := sysctlEntries["fs.file-max"]; ok {
		t.Errorf("did not expect 'fs.file-max' for non-matching hostname")
	}
	if sysEntries["sys:net.ipv4.tcp_slow_start_after_idle"].Value != "1" {
		t.Errorf("expected net.ipv4.tcp_slow_start_after_idle value '1', got '%s'", sysEntries["sys:net.ipv4.tcp_slow_start_after_idle"].Value)
	}
	if _, ok := vmEntries["net.ipv4.tcp_tw_reuse"]; ok {
		t.Errorf("did not expect 'net.ipv4.tcp_tw_reuse' for non-matching fqdn")
	}

	// Test ParseINIFile with the same content
	confFilePath := path.Join(tmpDir, "test_note.conf")
	if err := os.WriteFile(confFilePath, []byte(iniContent), 0644); err != nil {
		t.Errorf("failed to write test note conf: %v", err)
	}

	parsedFile, err := ParseINIFile(confFilePath, false)
	if err != nil {
		t.Errorf("ParseINIFile failed: %v", err)
	}
	fileSysctl, exists := parsedFile.KeyValue["sysctl"]
	if !exists {
		t.Errorf("expected 'sysctl' section in parsed file INI")
	}
	if _, ok := fileSysctl["kernel.sem"]; !ok {
		t.Errorf("expected 'kernel.sem' to be parsed in file INI")
	}
	if _, ok := fileSysctl["fs.file-max"]; ok {
		t.Errorf("did not expect 'fs.file-max' in file INI")
	}
}
