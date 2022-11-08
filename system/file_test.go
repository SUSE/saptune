package system

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var readFileMatchText = `Only a test for read file
`

func TestReadConfigFile(t *testing.T) {
	content, err := ReadConfigFile("/file_does_not_exist", true)
	if string(content) != "" {
		t.Error(content, err)
	}
	os.Remove("/file_does_not_exist")
	content, err = ReadConfigFile("/file_does_not_exist", false)
	if string(content) != "" || err == nil {
		t.Error(content, err)
	}
	//content, err = ReadConfigFile("/app/testdata/tstfile", false)
	content, err = ReadConfigFile(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile"), false)
	if string(content) != readFileMatchText || err != nil {
		t.Error(string(content), err)
	}
}

func TestFileIsEmpty(t *testing.T) {
	empty := FileIsEmpty(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/extra/wrongFileNamesyntax.conf"))
	if !empty {
		t.Errorf("file should be reported as empty, but returns 'true'")
	}
	empty = FileIsEmpty("/file_does_not_exist")
	if !empty {
		t.Errorf("file should be reported as empty (not existing), but returns 'true'")
	}
	empty = FileIsEmpty(path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/product_name"))
	if empty {
		t.Errorf("file should be reported as non empty, but returns 'false'")
	}
}

func TestCopyFile(t *testing.T) {
	//src := "/app/testdata/tstfile"
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	dst := "/tmp/saptune_tstfile"
	err := CopyFile(src, dst)
	if err != nil {
		t.Error(err)
	}
	content, err := ReadConfigFile(dst, false)
	if string(content) != readFileMatchText || err != nil {
		t.Error(string(content), err)
	}
	err = CopyFile("/file_does_not_exist", dst)
	if err == nil {
		t.Errorf("copied from non existing file")
	}
	err = CopyFile(src, "/tmp/saptune_test/saptune_tstfile")
	if err == nil {
		t.Errorf("copied to non existing file")
	}
	os.Remove(dst)
}

func TestListDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := os.Mkdir(path.Join(tmpDir, "aDir"), 0700); err != nil {
		t.Fatal(err)
	}
	if file, err := os.OpenFile(path.Join(tmpDir, "aFile"), os.O_CREATE, 0600); err != nil {
		t.Fatal(err)
	} else if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	dirs, files := ListDir(tmpDir, "")
	if !reflect.DeepEqual(dirs, []string{"aDir"}) {
		t.Fatal(dirs)
	}
	if !reflect.DeepEqual(files, []string{"aFile"}) {
		t.Fatal(files)
	}
}

func TestEditFile(t *testing.T) {
	oldEditor := os.Getenv("EDITOR")
	defer func() { os.Setenv("EDITOR", oldEditor) }()
	os.Setenv("EDITOR", "/usr/bin/cat")
	//src := "/app/testdata/tstfile"
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	dst := "/tmp/saptune_tstfile"
	err := EditFile(src, dst)
	if err != nil {
		t.Error(err)
	}
	err = EditFile("/file_does_not_exist", dst)
	if err == nil {
		t.Errorf("copied from non existing file")
	}
	err = EditFile(src, "/tmp/saptune_test/saptune_tstfile")
	if err == nil {
		t.Errorf("copied to non existing file")
	}
	os.Remove(dst)
}

func TestEditAndCheckFile(t *testing.T) {
	oldEditor := os.Getenv("EDITOR")
	defer func() { os.Setenv("EDITOR", oldEditor) }()
	os.Setenv("EDITOR", "/usr/bin/cat")
	//src := "/app/testdata/tstfile"
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	dst := "/tmp/saptune_tstfile"
	changed, err := EditAndCheckFile(src, dst, "ANGI", "note")
	if err != nil {
		t.Error(err)
	}
	if changed {
		t.Error("got 'true', but expected 'false'")
	}

	changed, err = EditAndCheckFile("/file_does_not_exist", dst, "ANGI", "note")
	if err == nil {
		t.Errorf("copied from non existing file")
	}
	os.Remove(dst)
}

func TestMD5(t *testing.T) {
	//src := "/app/testdata/tstfile"
	match1 := "1fd006c2c4a9c3bebb749b43889339f6"
	match2 := "28407312133599fac8e5d22dc16f2726"
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	src2 := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/sysconfig-sample")
	dst := "/tmp/saptune_tstfile"
	err := CopyFile(src, dst)
	if err != nil {
		t.Error(err)
	}
	if !ChkMD5Pair(src, dst) {
		t.Error("checksum should be equal, but is not.")
	}
	sum1, err := GetMD5Hash(src)
	if err != nil {
		t.Error(sum1, err)
	}
	dum1, err := GetMD5Hash(dst)
	if err != nil {
		t.Error(dum1, err)
	}
	if sum1 != dum1 {
		t.Errorf("checksum should be equal, but is not. %s - %s.\n", sum1, dum1)
	}
	if sum1 != match1 {
		t.Errorf("wrong checksum. got: %s, expected: %s\n", sum1, match1)
	}
	if ChkMD5Pair(src, "/file_does_not_exist") {
		t.Error("checksum should differ, but is equal.")
	}
	if ChkMD5Pair("/file_does_not_exist", dst) {
		t.Error("checksum should differ, but is equal.")
	}
	os.Remove(dst)

	sum2, err := GetMD5Hash(src2)
	if err != nil {
		t.Error(sum2, err)
	}
	if sum2 != match2 {
		t.Errorf("got: %s, expected: %s\n", sum2, match2)
	}
	sum3, err := GetMD5Hash("/file_does_not_exist")
	if err == nil {
		t.Error(sum2, err)
	}
	if sum3 != "" {
		t.Errorf("got: %s, expected: \n", sum3)
	}
}

func TestCleanUpRun(t *testing.T) {
	src := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/tstfile")
	dst := "/run/saptune/sections/tstfile.run"
	dst2 := "/run/saptune/sections/tstfile.sections"
	if err := os.MkdirAll(SaptuneSectionDir, 0755); err != nil {
		t.Error(err)
	}
	err := CopyFile(src, dst)
	if err != nil {
		t.Error(err)
	}
	err = CopyFile(src, dst2)
	if err != nil {
		t.Error(err)
	}
	CleanUpRun()
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(dst2); os.IsNotExist(err) {
		t.Errorf("file '%s' does not exist\n", dst2)
	}
	os.Remove(dst)
	os.Remove(dst2)
	os.RemoveAll(SaptuneSectionDir)
}

func TestBackupValue(t *testing.T) {
	start := "12488"
	file := "/tmp/tst_backup"
	WriteBackupValue(start, file)
	val := GetBackupValue(file)
	if val != start {
		t.Errorf("got: %s, expected: %s\n", val, start)
	}
	start = ""
	file = "/tmp/tst_backup"
	WriteBackupValue(start, file)
	val = GetBackupValue(file)
	if val != "NA" {
		t.Errorf("got: %s, expected: NA\n", val)
	}
}

func TestAddGap(t *testing.T) {
	os.Args = []string{"saptune", "--format=json"}
	RereadArgs()
	buffer := bytes.Buffer{}
	AddGap(&buffer)
	txt := buffer.String()
	if txt != "" {
		t.Errorf("got: %s, expected: empty\n", txt)
	}
	os.Args = []string{"saptune", "status"}
	RereadArgs()
	buffer2 := bytes.Buffer{}
	AddGap(&buffer2)
	txt2 := buffer2.String()
	if txt2 != "\n" {
		t.Errorf("got: %s, expected: '\n'\n", txt2)
	}
	os.Args = []string{"saptune", "status", "--format="}
	RereadArgs()
	buffer3 := bytes.Buffer{}
	AddGap(&buffer3)
	txt3 := buffer3.String()
	if txt3 != "\n" {
		t.Errorf("got: %s, expected: '\n'\n", txt3)
	}
}
