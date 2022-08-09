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

// IsXFSOption matches xfs options
var IsXFSOption = regexp.MustCompile(`^xfsopt_\w+$`)
var fstab = "/etc/fstab"
var mtab = "/etc/mtab"
var procMounts = "/proc/mounts"

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

// GetByMountOption find a mount point with special mount option.
// returns a list of mount points containing the option and a second list
// with mount points missing the option.
func (mounts MountPoints) GetByMountOption(fstype, mountOption, chkDflt string) ([]string, []string) {
	var found bool
	var dflt bool
	mntOK := []string{}
	mntNok := []string{}
	for _, mount := range mounts {
		if mount.Type == fstype {
			found = false
			dflt = false
			for _, opt := range mount.Options {
				if opt == mountOption {
					found = true
					mntOK = append(mntOK, mount.MountPoint)
					break
				}
				if opt == "defaults" {
					dflt = true
				}
			}
			if !found {
				if dflt && chkDflt == "chkOK" {
					mntOK = append(mntOK, mount.MountPoint)
				} else {
					mntNok = append(mntNok, mount.MountPoint)
				}
			}
		}
	}
	return mntOK, mntNok
}

// ParseMounts return all mount points defined in the input text.
// Skipping malformed entry.
func ParseMounts(txt string) (mounts MountPoints) {
	mounts = make([]MountPoint, 0)
	for _, line := range strings.Split(txt, "\n") {
		fields := consecutiveSpaces.Split(strings.TrimSpace(line), -1)
		if len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#' {
			continue // skip comments and empty lines
		}
		if len(fields) < 4 {
			// skip lines with wrong syntax
			ErrorLog("parsing mounts - incorrect number of fields in line '%s'. Skipping entry", line)
			continue
		}
		mountPoint := MountPoint{
			Device:     fields[0],
			MountPoint: fields[1],
			Type:       fields[2],
			Dump:       0,
			Fsck:       0,
		}
		// Split mount options
		mountPoint.Options = mountOptionSeparator.Split(fields[3], -1)
		var err error
		if len(fields) > 4 {
			if mountPoint.Dump, err = strconv.Atoi(fields[4]); err != nil {
				WarningLog("parsing mounts - not an integer for field 'dump' in '%s'. Working with default value '0'.", line)
			}
		}
		if len(fields) > 5 {
			if mountPoint.Fsck, err = strconv.Atoi(fields[5]); err != nil {
				WarningLog("parsing mounts - not an integer for field 'fsck' in '%s'. Working with default value '0'.", line)
			}
		}
		mounts = append(mounts, mountPoint)
	}
	return
}

// ParseMtab return all mount points defined in a given file.
// Returns empty list of mount points on error
func ParseMtab(file string) MountPoints {
	mounts := ""
	content, err := ioutil.ReadFile(file)
	if err != nil {
		ErrorLog("failed to read file '%s': %v", file, err)
	} else {
		mounts = string(content)
	}
	return ParseMounts(mounts)
}

// ParseFstab return all mount points defined in /etc/fstab.
// Returns empty list of mount points
func ParseFstab() MountPoints {
	return ParseMtab(fstab)
}

// ParseProcMounts return all mount points appearing in /proc/mounts.
// Returns empty list of mount points
func ParseProcMounts() MountPoints {
	return ParseMtab(procMounts)
}

// ParseMtabMounts return all mount points appearing in /etc/mtab.
// Returns empty list of mount points
func ParseMtabMounts() MountPoints {
	return ParseMtab(mtab)
}

// RemountSHM invoke mount command to resize /dev/shm to the specified value.
func RemountSHM(newSizeMB uint64) error {
	cmd := exec.Command("mount", "-o", fmt.Sprintf("remount,size=%dM", newSizeMB), "/dev/shm")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to invoke external command mount: %v, output: %s", err, out)
	}
	return nil
}

// GetMountOpts checks if mount points with the given type exists and contain
// the needed/not needed option.
// Returns a list of mount point containing the option and a list of mount
// point NOT containing the option
func GetMountOpts(mustExist bool, fstype, fsopt string) ([]string, []string) {
	// Find out mount options
	chkdflt := "noChk"
	// check the mounted FS
	mountProcOk, mountProcNok := ParseProcMounts().GetByMountOption(fstype, fsopt, chkdflt)
	if mustExist {
		chkdflt = "chkOK"
	} else {
		chkdflt = "chkNOK"
	}
	// check /etc/fstab to get the not mounted FS as well
	mountFSTOk, mountFSTNok := ParseFstab().GetByMountOption(fstype, fsopt, chkdflt)
	mntOk := getMounts(mountProcOk, mountFSTOk)
	mntNok := getMounts(mountProcNok, mountFSTNok)
	return mntOk, mntNok
}

// getMounts combines the mounted and not mounted FS
func getMounts(neededProcMnts, mntsFromFstab []string) []string {
	// initialize with the mounted FS
	mntRet := neededProcMnts
	for _, mnt := range mntsFromFstab {
		// search for not mounted FS, which are NOK/OK
		// and append them to the needed mounts
		found := isMntAvail(mnt, neededProcMnts)
		if !found {
			mntRet = append(mntRet, mnt)
		}
	}
	return mntRet
}

// isMntAvail checks, if a given mount point is available in a pool of
// mount points.
// returns true, if the mount point exists in the pool, otherwise false
func isMntAvail(mntPt string, poolOfMnts []string) bool {
	ret := false
	for _, mnt := range poolOfMnts {
		if mntPt == mnt {
			ret = true
			break
		}
	}
	return ret
}
