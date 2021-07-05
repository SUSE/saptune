package note

import (
	"github.com/SUSE/saptune/system"
	"strconv"
	"testing"
)

func TestGetMemVal(t *testing.T) {
	val := GetMemVal("VSZ_TMPFS_PERCENT")
	if val == "-1" {
		t.Log("/dev/shm not found")
	}
	val = GetMemVal("ShmFileSystemSizeMB")
	if val == "-1" {
		t.Log("/dev/shm not found")
	}
	val = GetMemVal("UNKOWN_PARAMETER")
	if val != "" {
		t.Error(val)
	}
}

func TestOptMemVal(t *testing.T) {
	val := OptMemVal("VSZ_TMPFS_PERCENT", "47", "80", "80")
	if val != "80" {
		t.Error(val)
	}
	val = OptMemVal("VSZ_TMPFS_PERCENT", "-1", "75", "75")
	if val != "75" {
		t.Error(val)
	}

	size75 := uint64(system.GetTotalMemSizeMB()) * 75 / 100
	size80 := uint64(system.GetTotalMemSizeMB()) * 80 / 100

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "80")
	if val != strconv.FormatUint(size80, 10) {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "80")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "0")
	if val != strconv.FormatUint(size75, 10) {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "0")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "80")
	if val != "25605" {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "80")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "0")
	if val != "25605" {
		t.Error(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "0")
	if val != "-1" {
		t.Error(val)
	}

	val = OptMemVal("UNKOWN_PARAMETER", "16043", "0", "0")
	if val != "" {
		t.Error(val)
	}
	val = OptMemVal("UNKOWN_PARAMETER", "-1", "0", "0")
	if val != "" {
		t.Error(val)
	}
}

//SetMemVal
