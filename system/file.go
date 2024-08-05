package system

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
)

// ReadConfigFile read content of config file
func ReadConfigFile(fileName string, autoCreate bool) ([]byte, error) {
	content, err := os.ReadFile(fileName)
	if os.IsNotExist(err) && autoCreate {
		content = []byte{}
		err = os.MkdirAll(path.Dir(fileName), 0755)
		if err == nil {
			err = os.WriteFile(fileName, []byte{}, 0644)
		}
	}
	return content, err
}

// FileIsEmpty returns true, if the given file is empty or does not exist
// or false, if exist, but not empty
func FileIsEmpty(fileName string) bool {
	f, err := os.Stat(fileName)
	if err == nil && f.Size() != 0 {
		DebugLog("FileIsEmpty - file '%s' exists and is NOT empty(%+v)", fileName, f.Size())
		return false
	}
	return true
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

// EditAndCheckFile creates or modify note or solution definition files
// using an editor
func EditAndCheckFile(srcFileName, destFileName, defName, defType string) (bool, error) {
	var err error
	changed := false
	tmpFile := fmt.Sprintf("/tmp/%s.sttemp", defName)
	if err = EditFile(srcFileName, tmpFile); err != nil {
		// clean up before exit
		os.Remove(tmpFile)
		ErrorLog("Problems while editing %s definition file '%s' - %v", defType, destFileName, err)
		return changed, err
	}
	// check if something was changed in the file
	if !ChkMD5Pair(srcFileName, tmpFile) {
		// template and temporary file differ, so something was
		// written/changed during the editor session
		// copy temporary file to extra location
		if err = CopyFile(tmpFile, destFileName); err != nil {
			// clean up before exit
			os.Remove(tmpFile)
			ErrorLog("Problems writing %s definition file '%s' - %v", defType, destFileName, err)
			return changed, err
		}
		changed = true
	}
	// remove no longer needed temporary file
	os.Remove(tmpFile)
	return changed, err
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
	entries, err := os.ReadDir(dir)
	if err != nil {
		DebugLog("failed to read %s, called from '%v' - %v", dir, CalledFrom(), err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files[entry.Name()] = entry.Name()
		}
	}
	return files
}

// ListDir list directory content and returns a slice for the directory names
// and a slice for the file names.
func ListDir(dirPath, logMsg string) (dirNames, fileNames []string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil && logMsg != "" {
		// Not a fatal error
		WarningLog("failed to read %s - %v", logMsg, err)
	}
	dirNames = make([]string, 0)
	fileNames = make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		} else {
			fileNames = append(fileNames, entry.Name())
		}
	}
	return
}

// CleanUpRun cleans up runtime files
func CleanUpRun() {
	var runfile = regexp.MustCompile(`.*\.run$`)
	content, _ := os.ReadDir(SaptuneSectionDir)
	for _, entry := range content {
		if runfile.MatchString(entry.Name()) {
			// remove runtime file
			_ = os.Remove(path.Join(SaptuneSectionDir, entry.Name()))
		}
	}
}

// GetBackupValue reads the value from the backup file
// currently used for the former start TasksMax value
func GetBackupValue(fileName string) string {
	value := ""
	content, err := os.ReadFile(fileName)
	if err != nil {
		DebugLog("reading backup file '%s' failed - '%v'", fileName, err)
		return "NA"
	}
	value = string(content)
	if value == "" {
		value = "NA"
	}
	return value
}

// WriteBackupValue writes a value into the backup file
// currently used for the former start TasksMax value
func WriteBackupValue(value, fileName string) {
	err := os.WriteFile(fileName, []byte(value), 0600)
	if err != nil {
		DebugLog("writing backup file '%s' for value '%s' failed - '%v'", fileName, value, err)
	}
}

// AddGap adds an empty line to improve readability of the screen output
func AddGap(writer io.Writer) {
	if GetFlagVal("format") == "" || GetFlagVal("format") == "flag_value" {
		fmt.Fprintf(writer, "\n")
	}
}
