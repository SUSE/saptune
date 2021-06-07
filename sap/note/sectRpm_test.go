package note

import (
	"testing"
)

func TestGetRpmVal(t *testing.T) {
	val := GetRpmVal("rpm:glibc")
	if val == "" {
		t.Log("rpm 'glibc' not found")
	}
}

func TestOptRpmVal(t *testing.T) {
	val := OptRpmVal("rpm:glibc", "NO_OPT")
	if val != "NO_OPT" {
		t.Error(val)
	}
}

func TestSetRpmVal(t *testing.T) {
	val := SetRpmVal("NO_OPT")
	if val != nil {
		t.Error(val)
	}
}
