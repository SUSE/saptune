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
}
