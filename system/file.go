package system

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
)

// ReadConfigFile read content of config file
func ReadConfigFile(fileName string, autoCreate bool) ([]byte, error) {
	content, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) && autoCreate {
		content = []byte{}
		err = os.MkdirAll(path.Dir(fileName), 0755)
		if err == nil {
			err = ioutil.WriteFile(fileName, []byte{}, 0644)
		}
	}
	return content, err
}

// EditFile copies a source file to another name and opens this copy in an
// editor defined by environment variable "EDITOR" or in 'vim'
func EditFile(srcFile, destFile string) error {
	editor := os.Getenv("EDITOR")
	// copy source to destintion
	if err := CopyFile(srcFile, destFile); err != nil {
		ErrorLog("Problems while copying '%s' to '%s' - %v", srcFile, destFile, err)
		return err
	}
	if editor == "" {
		editor = "/usr/bin/vim" // launch vim by default
	}
	cmd := exec.Command(editor, destFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ErrorLog("Failed to launch editor %s: %v", editor, err)
	}
	return err
}

// ChkMD5Pair checks, if the md5sum of 2 files are equal
func ChkMD5Pair(srcFile, destFile string) bool {
	ret := false
	chkSumSrc, err := GetMD5Hash(srcFile)
	if err != nil {
		ErrorLog("Failed to get md5 checksum of file '%s': %v", srcFile, err)
	}
	chkSumDest, err := GetMD5Hash(destFile)
	if err != nil {
		ErrorLog("Failed to get md5 checksum of file '%s': %v", destFile, err)
	}
	if chkSumSrc == chkSumDest && chkSumSrc != "" {
		ret = true
	}
	return ret
}

// GetMD5Hash generate the md5sum of a file
func GetMD5Hash(file string) (string, error) {
	md5Sum := ""
	// open file for reading
	f, err := os.Open(file)
	if err != nil {
		return md5Sum, err
	}
	defer f.Close()

	// create a new hash, which is a writer interface
	hash := md5.New()

	// copy the file in the hash interface
	if _, err := io.Copy(hash, f); err != nil {
		return md5Sum, err
	}
	// hash and print as string. Pass nil since the data is not coming
	// in as a slice argument but is coming through the writer interface
	md5Sum = fmt.Sprintf("%x", hash.Sum(nil))
	return md5Sum, nil
}

// CopyFile from source to destination
func CopyFile(srcFile, destFile string) error {
	var src, dst *os.File
	var err error
	if src, err = os.Open(srcFile); err == nil {
		defer src.Close()
		if dst, err = os.OpenFile(destFile, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644); err == nil {
			defer dst.Close()
			if _, err = io.Copy(dst, src); err == nil {
				// flush file content from  memory to disk
				err = dst.Sync()
			}
		}
	}
	return err
}

// GetFiles returns the files from a directory as map
// skip directories
func GetFiles(dir string) map[string]string {
	files := make(map[string]string)
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		WarningLog("failed to read %s - %v", dir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files[entry.Name()] = entry.Name()
		}
	}
	return files
}

// CleanUpRun cleans up runtime files
func CleanUpRun() {
	var runfile = regexp.MustCompile(`.*\.run$`)
	content, _ := ioutil.ReadDir(SaptuneSectionDir)
	for _, entry := range content {
		if runfile.MatchString(entry.Name()) {
			// remove runtime file
			_ = os.Remove(path.Join(SaptuneSectionDir, entry.Name()))
		}
	}
}
