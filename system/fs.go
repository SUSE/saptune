package system

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var mountOptionSeparator = regexp.MustCompile("[[:space:]]*,[[:space:]]*")

// MountPoint Represent a mount point entry in /proc/mounts or /etc/fstab
type MountPoint struct {
	Device     string
	MountPoint string
	Type       string
	Options    []string
	Dump       int
	Fsck       int
}

// Equals return true only if two mount points are identical in all attributes.
func (mount1 MountPoint) Equals(mount2 MountPoint) bool {
	return reflect.DeepEqual(mount1, mount2)
}

// GetFileSystemSizeMB return the total size of the file system in MegaBytes.
// Panic on error.
func (mount MountPoint) GetFileSystemSizeMB() uint64 {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(mount.MountPoint, &fs)
	if err != nil {
		panic(fmt.Errorf("failed to stat file system on mount point %s - %v", mount.MountPoint, err))
	}
	return uint64(fs.Bsize) * fs.Blocks / 1048576
}

// MountPoints contains a list of mount points.
type MountPoints []MountPoint

// GetByMountPoint find a mount point by its path.
func (mounts MountPoints) GetByMountPoint(mountPoint string) (MountPoint, bool) {
	for _, mount := range mounts {
		if mount.MountPoint == mountPoint {
			return mount, true
		}
	}
	return MountPoint{}, false
}

// ParseMounts return all mount points defined in the input text.
// Panic on malformed entry.
func ParseMounts(txt string) (mounts MountPoints) {
	mounts = make([]MountPoint, 0, 0)
	for _, line := range strings.Split(txt, "\n") {
		fields := consecutiveSpaces.Split(strings.TrimSpace(line), -1)
		if len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#' {
			continue // skip comments and empty lines
		}
		if len(fields) != 6 {
			panic(fmt.Sprintf("parsing mounts - incorrect number of fields in '%s'", line))
		}
		mountPoint := MountPoint{
			Device:     fields[0],
			MountPoint: fields[1],
			Type:       fields[2],
		}
		// Split mount options
		mountPoint.Options = mountOptionSeparator.Split(fields[3], -1)
		var err error
		if mountPoint.Dump, err = strconv.Atoi(fields[4]); err != nil {
			panic(fmt.Sprintf("parsing mounts - not an integer in '%s'", line))
		}
		if mountPoint.Fsck, err = strconv.Atoi(fields[4]); err != nil {
			panic(fmt.Sprintf("parsing mounts - not an integer in '%s'", line))
		}
		mounts = append(mounts, mountPoint)
	}
	return
}

// ParseFstab return all mount points defined in /etc/fstab. Panic on error.
func ParseFstab() MountPoints {
	fstab, err := ioutil.ReadFile("/etc/fstab")
	if err != nil {
		panic(fmt.Errorf("failed to read /etc/fstab: %v", err))
	}
	return ParseMounts(string(fstab))
}

// ParseProcMounts return all mount points appearing in /proc/mounts.
// Panic on error.
func ParseProcMounts() MountPoints {
	mounts, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		panic(fmt.Errorf("failed to open /proc/mounts: %v", err))
	}
	return ParseMounts(string(mounts))
}

// ParseMtabMounts return all mount points appearing in /proc/mounts.
// Panic on error.
func ParseMtabMounts() MountPoints {
	mounts, err := ioutil.ReadFile("/etc/mtab")
	if err != nil {
		panic(fmt.Errorf("failed to open /etc/mtab: %v", err))
	}
	return ParseMounts(string(mounts))
}

// RemountSHM invoke mount command to resize /dev/shm to the specified value.
func RemountSHM(newSizeMB uint64) error {
	cmd := exec.Command("mount", "-o", fmt.Sprintf("remount,size=%dM", newSizeMB), "/dev/shm")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to invoke external command mount: %v, output: %s", err, out)
	}
	return nil
}

// ListDir list directory content.
func ListDir(dirPath string) (dirNames, fileNames []string, err error) {
	entries, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}
	dirNames = make([]string, 0, 0)
	fileNames = make([]string, 0, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		} else {
			fileNames = append(fileNames, entry.Name())
		}
	}
	return
}
