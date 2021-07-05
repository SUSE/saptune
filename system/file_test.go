package system

import (
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
