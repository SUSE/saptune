package system

import (
	"testing"
)

func TestMax(t *testing.T) {
	if val := MaxU64(); val != 0 {
		t.Fatal(val)
	}
	if val := MaxU64(4, 3, 5, 2, 6); val != 6 {
		t.Fatal(val)
	}

	if val := MaxI64(); val != 0 {
		t.Fatal(val)
	}
	if val := MaxI64(4, 3, -5, -2, 6); val != 6 {
		t.Fatal(val)
	}

	if val := MaxI(); val != 0 {
		t.Fatal(val)
	}
	if val := MaxI(4, 3, -5, -2, 6); val != 6 {
		t.Fatal(val)
	}
}

func TestMin(t *testing.T) {
	if val := MinU64(0); val != 0 {
		t.Fatal(val)
	}
	if val := MinU64(4, 3, 5, 2, 6); val != 2 {
		t.Fatal(val)
	}
	if val := MinU64(); val != 0 {
		t.Fatal(val)
	}
}
