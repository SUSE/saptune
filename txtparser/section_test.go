package txtparser

import (
	"github.com/SUSE/saptune/system"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestStoreSectionInfo(t *testing.T) {
	saptuneSectionDir = "/tmp/saptune_sections"
	runFile := path.Join(saptuneSectionDir, "1234567.run")
	sectionFile := path.Join(saptuneSectionDir, "1234567.sections")
	if err := os.Mkdir(saptuneSectionDir, 0755); err != nil {
		t.Error(err)
	}
	// parse configuration file
	iniPath := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/ini_test.ini")
	ini, err := ParseINIFile(iniPath, false)
	if err != nil {
		t.Error(err)
	}
	// write section data
	err = StoreSectionInfo(ini, "run", "1234567", true)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(runFile); err != nil {
		t.Error(err)
	}

	readIni, err := GetSectionInfo("sns", "1234567", false)
	if !reflect.DeepEqual(ini, readIni) {
		t.Errorf("got: %+v, expected: %+v\n", readIni, ini)
	}

	// looking for override file
	override, ow := GetOverrides("ovw", "1234567")
	if override {
		t.Errorf("override file found, but should not exist - '%+v'\n", ow)
	}

	_ = system.CopyFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/etc/saptune/override/1234567"), "/etc/saptune/override/1234567")
	override, ow = GetOverrides("ovw", "1234567")
	if !override {
		t.Errorf("override file not found, but should exist - '%+v'\n", ow)
	}
	os.Remove("/etc/saptune/override/1234567")

	// write section data
	err = StoreSectionInfo(ini, "section", "1234567", true)
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(sectionFile); err != nil {
		t.Error(err)
	}
	readIni, err = GetSectionInfo("sns", "1234567", true)
	if !reflect.DeepEqual(ini, readIni) {
		t.Errorf("got: %+v, expected: %+v\n", readIni, ini)
	}

	defer os.RemoveAll(saptuneSectionDir)
}
