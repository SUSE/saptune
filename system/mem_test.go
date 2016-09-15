package system

import (
	"os"
	"testing"
)

func TestParseMeminfo(t *testing.T) {
	infoMap := ParseMeminfo()
	if size, exists := infoMap[MEMINFO_MAIN_TOTAL_KEY]; !exists || size <= 0 {
		t.Fatal(size, MEMINFO_MAIN_TOTAL_KEY)
	}
	if _, exists := infoMap[MEMINFO_SWAP_TOTAL_KEY]; !exists {
		t.Fatal(MEMINFO_SWAP_TOTAL_KEY)
	}
}

func TestGetMemSize(t *testing.T) {
	if size := GetMainMemSizeMB(); size < 1 {
		t.Fatal(size)
	}
	if size := GetTotalMemSizeMB(); size < 1 {
		t.Fatal(size)
	}
}

func TestGetTotalMemSizePages(t *testing.T) {
	if pages := GetTotalMemSizePages(); pages != GetTotalMemSizeMB()*1024/uint64(os.Getpagesize()) {
		t.Fatal(pages)
	}
}

func TestGetSemaphoreLimits(t *testing.T) {
	msl, mns, opm, mni := GetSemaphoreLimits()
	if msl < 3 || mns < 3 || opm < 3 || mni < 3 {
		t.Fatal(msl, mns, opm, mni)
	}
}
