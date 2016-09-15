package system

import (
	"testing"
)

var procMountsSample = `
# SLES 11
rootfs / rootfs rw 0 0
udev /dev tmpfs rw,relatime,nr_inodes=0,mode=755 0 0
tmpfs /dev/shm tmpfs rw,relatime,size=8388608k 0 0
/dev/vda2 / ext3 rw,relatime,errors=continue,user_xattr,acl,barrier=1,data=ordered 0 0
proc /proc proc rw,relatime 0 0
sysfs /sys sysfs rw,relatime 0 0
devpts /dev/pts devpts rw,relatime,gid=5,mode=620,ptmxmode=000 0 0
debugfs /sys/kernel/debug debugfs rw,relatime 0 0
fusectl /sys/fs/fuse/connections fusectl rw,relatime 0 0
securityfs /sys/kernel/security securityfs rw,relatime 0 0
gvfs-fuse-daemon /root/.gvfs fuse.gvfs-fuse-daemon rw,nosuid,nodev,relatime,user_id=0,group_id=0 0 0

# SLES 12
rootfs / rootfs rw 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
devtmpfs /dev devtmpfs rw,nosuid,size=4086316k,nr_inodes=1021579,mode=755 0 0
securityfs /sys/kernel/security securityfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev,size=8388608k 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
tmpfs /sys/fs/cgroup tmpfs rw,nosuid,nodev,noexec,mode=755 0 0
cgroup /sys/fs/cgroup/systemd cgroup rw,nosuid,nodev,noexec,relatime,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd 0 0
pstore /sys/fs/pstore pstore rw,nosuid,nodev,noexec,relatime 0 0
cgroup /sys/fs/cgroup/cpuset cgroup rw,nosuid,nodev,noexec,relatime,cpuset 0 0
cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpuacct,cpu 0 0
cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
cgroup /sys/fs/cgroup/freezer cgroup rw,nosuid,nodev,noexec,relatime,freezer 0 0
cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
cgroup /sys/fs/cgroup/perf_event cgroup rw,nosuid,nodev,noexec,relatime,perf_event 0 0
cgroup /sys/fs/cgroup/hugetlb cgroup rw,nosuid,nodev,noexec,relatime,hugetlb 0 0
/dev/vda2 / ext4 rw,relatime,data=ordered 0 0
systemd-1 /proc/sys/fs/binfmt_misc autofs rw,relatime,fd=28,pgrp=1,timeout=300,minproto=5,maxproto=5,direct 0 0
debugfs /sys/kernel/debug debugfs rw,relatime 0 0
hugetlbfs /dev/hugepages hugetlbfs rw,relatime 0 0
mqueue /dev/mqueue mqueue rw,relatime 0 0
gvfsd-fuse /run/user/0/gvfs fuse.gvfsd-fuse rw,nosuid,nodev,relatime,user_id=0,group_id=0 0 0
fusectl /sys/fs/fuse/connections fusectl rw,relatime 0 0

# SLES 12 SAP
rootfs / rootfs rw 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
devtmpfs /dev devtmpfs rw,nosuid,size=4086164k,nr_inodes=1021541,mode=755 0 0
securityfs /sys/kernel/security securityfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev,size=8388608k 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
tmpfs /sys/fs/cgroup tmpfs rw,nosuid,nodev,noexec,mode=755 0 0
cgroup /sys/fs/cgroup/systemd cgroup rw,nosuid,nodev,noexec,relatime,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd 0 0
pstore /sys/fs/pstore pstore rw,nosuid,nodev,noexec,relatime 0 0
cgroup /sys/fs/cgroup/cpuset cgroup rw,nosuid,nodev,noexec,relatime,cpuset 0 0
cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpuacct,cpu 0 0
cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
cgroup /sys/fs/cgroup/freezer cgroup rw,nosuid,nodev,noexec,relatime,freezer 0 0
cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
cgroup /sys/fs/cgroup/perf_event cgroup rw,nosuid,nodev,noexec,relatime,perf_event 0 0
cgroup /sys/fs/cgroup/hugetlb cgroup rw,nosuid,nodev,noexec,relatime,hugetlb 0 0
/dev/vda2 / ext4 rw,relatime,data=ordered 0 0
systemd-1 /proc/sys/fs/binfmt_misc autofs rw,relatime,fd=31,pgrp=1,timeout=300,minproto=5,maxproto=5,direct 0 0
hugetlbfs /dev/hugepages hugetlbfs rw,relatime 0 0
debugfs /sys/kernel/debug debugfs rw,relatime 0 0
mqueue /dev/mqueue mqueue rw,relatime 0 0
gvfsd-fuse /run/user/0/gvfs fuse.gvfsd-fuse rw,nosuid,nodev,relatime,user_id=0,group_id=0 0 0
fusectl /sys/fs/fuse/connections fusectl rw,relatime 0 0

# Tumbleweed
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime,hidepid=2 0 0
devtmpfs /dev devtmpfs rw,nosuid,size=16427624k,nr_inodes=4106906,mode=755 0 0
securityfs /sys/kernel/security securityfs rw,nosuid,nodev,noexec,relatime 0 0
tmpfs /dev/shm tmpfs rw,nosuid,nodev,size=5120000k 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,nodev,mode=755 0 0
tmpfs /sys/fs/cgroup tmpfs ro,nosuid,nodev,noexec,mode=755 0 0
cgroup /sys/fs/cgroup/systemd cgroup rw,nosuid,nodev,noexec,relatime,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd 0 0
pstore /sys/fs/pstore pstore rw,nosuid,nodev,noexec,relatime 0 0
cgroup /sys/fs/cgroup/blkio cgroup rw,nosuid,nodev,noexec,relatime,blkio 0 0
cgroup /sys/fs/cgroup/hugetlb cgroup rw,nosuid,nodev,noexec,relatime,hugetlb 0 0
cgroup /sys/fs/cgroup/net_cls,net_prio cgroup rw,nosuid,nodev,noexec,relatime,net_cls,net_prio 0 0
cgroup /sys/fs/cgroup/perf_event cgroup rw,nosuid,nodev,noexec,relatime,perf_event 0 0
cgroup /sys/fs/cgroup/cpuset cgroup rw,nosuid,nodev,noexec,relatime,cpuset 0 0
cgroup /sys/fs/cgroup/memory cgroup rw,nosuid,nodev,noexec,relatime,memory 0 0
cgroup /sys/fs/cgroup/cpu,cpuacct cgroup rw,nosuid,nodev,noexec,relatime,cpu,cpuacct 0 0
cgroup /sys/fs/cgroup/pids cgroup rw,nosuid,nodev,noexec,relatime,pids 0 0
cgroup /sys/fs/cgroup/devices cgroup rw,nosuid,nodev,noexec,relatime,devices 0 0
cgroup /sys/fs/cgroup/freezer cgroup rw,nosuid,nodev,noexec,relatime,freezer 0 0
/dev/sda1 / ext4 rw,relatime,data=ordered 0 0
systemd-1 /proc/sys/fs/binfmt_misc autofs rw,relatime,fd=25,pgrp=1,timeout=0,minproto=5,maxproto=5,direct 0 0
mqueue /dev/mqueue mqueue rw,relatime 0 0
hugetlbfs /dev/hugepages hugetlbfs rw,relatime 0 0
debugfs /sys/kernel/debug debugfs rw,relatime 0 0
tmpfs /var/run tmpfs rw,nosuid,nodev,mode=755 0 0
/dev/sdb1 /mass ext4 rw,relatime,data=ordered 0 0
tmpfs /run/user/0 tmpfs rw,nosuid,nodev,relatime,size=3286976k,mode=700 0 0
tmpfs /var/run/user/0 tmpfs rw,nosuid,nodev,relatime,size=3286976k,mode=700 0 0
tmpfs /run/user/472 tmpfs rw,nosuid,nodev,relatime,size=3286976k,mode=700,uid=472,gid=474 0 0
tmpfs /var/run/user/472 tmpfs rw,nosuid,nodev,relatime,size=3286976k,mode=700,uid=472,gid=474 0 0
tmpfs /run/user/1000 tmpfs rw,nosuid,nodev,relatime,size=3286976k,mode=700,uid=1000,gid=100 0 0
tmpfs /var/run/user/1000 tmpfs rw,nosuid,nodev,relatime,size=3286976k,mode=700,uid=1000,gid=100 0 0
gvfsd-fuse /run/user/1000/gvfs fuse.gvfsd-fuse rw,nosuid,nodev,relatime,user_id=1000,group_id=100 0 0
gvfsd-fuse /var/run/user/1000/gvfs fuse.gvfsd-fuse rw,nosuid,nodev,relatime,user_id=1000,group_id=100 0 0
fusectl /sys/fs/fuse/connections fusectl rw,relatime 0 0
tracefs /sys/kernel/debug/tracing tracefs rw,relatime 0 0
binfmt_misc /proc/sys/fs/binfmt_misc binfmt_misc rw,relatime 0 0
`

var fstabSample = `
# SLES 11
/dev/vda1            swap                 swap       defaults              0 0
/dev/vda2            /                    ext3       acl,user_xattr        1 1
proc                 /proc                proc       defaults              0 0
sysfs                /sys                 sysfs      noauto                0 0
debugfs              /sys/kernel/debug    debugfs    noauto                0 0
usbfs                /proc/bus/usb        usbfs      noauto                0 0
devpts               /dev/pts             devpts     mode=0620,gid=5       0 0

# SLES 12
UUID=348fec89-d6bb-4683-a289-edb3401882aa swap                 swap       defaults              0 0
UUID=42ae7347-ce6b-49bb-b44c-a61f6f1d2c78 /                    ext4       acl,user_xattr        1 1

# SLES 12 SAP
UUID=d0d05f85-78e0-404b-89eb-146eeb770584 swap                 swap       defaults              0 0
UUID=12931751-bad1-4b49-992c-4eee57dda0a1 /                    ext4       acl,user_xattr        1 1

# Tumbleweed
UUID=069ea3e6-573e-48e4-87ff-00b15ab6f2db swap                 swap       defaults              0 0
UUID=c32aa786-b9c2-4212-8f5b-c4ab01f1ad91 /                    ext4       acl,user_xattr        1 1
UUID=92595693-aa49-45d6-9770-a767c498d40d /mass                ext4       defaults              1 2
`

func TestParseMounts(t *testing.T) {
	// source from /proc/mounts
	mountPoints := ParseMounts(procMountsSample)
	if len(mountPoints) != 101 {
		t.Fatal(len(mountPoints))
	}
	for _, mount := range mountPoints {
		if mount.Device == "" || mount.MountPoint == "" || len(mount.Options) < 1 || mount.Type == "" {
			t.Fatal(mount)
		}
	}
	shmMount := MountPoint{
		Device:     "tmpfs",
		MountPoint: "/dev/shm",
		Type:       "tmpfs",
		Options:    []string{"rw", "relatime", "size=8388608k"},
		Dump:       0,
		Fsck:       0,
	}
	if mount, found := mountPoints.GetByMountPoint("/dev/shm"); !found || !mount.Equals(shmMount) {
		t.Fatal(mount, found)
	}
	if mount, found := mountPoints.GetByMountPoint("/doesnotexist"); found || mount.MountPoint != "" {
		t.Fatal(mount, found)
	}

	// source from /etc/fstab
	mountPoints = ParseMounts(fstabSample)
	if len(mountPoints) != 14 {
		t.Fatal(len(mountPoints))
	}
	for _, mount := range mountPoints {
		if mount.Device == "" || mount.MountPoint == "" || len(mount.Options) < 1 || mount.Type == "" {
			t.Fatal(mount)
		}
	}

	vda2Mount := MountPoint{
		Device:     "/dev/vda2",
		MountPoint: "/",
		Type:       "ext3",
		Options:    []string{"acl", "user_xattr"},
		Dump:       1,
		Fsck:       1,
	}
	if mount, found := mountPoints.GetByMountPoint("/"); !found || !mount.Equals(vda2Mount) {
		t.Fatal(mount, found)
	}
	if mount, found := mountPoints.GetByMountPoint("/doesnotexist"); found || mount.MountPoint != "" {
		t.Fatal(mount, found)
	}
}

func TestMountPointGetFileSystemSizeMB(t *testing.T) {
	mountPoints := ParseMtabMounts()
	mount, found := mountPoints.GetByMountPoint("/")
	if !found {
		t.Fatal(mount, found)
	}
	if size := mount.GetFileSystemSizeMB(); size < 30 {
		t.Fatal(size)
	}
}
