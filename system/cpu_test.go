package system

import "testing"

func TestReadPerfBias(t *testing.T) {
	if !IsUserRoot() {
		t.Skip("the test requires root access")
	}
	if value := GetPerfBias(); len(value) < 3 {
		t.Fatal(value)
	}
}
