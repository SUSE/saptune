package sap

import (
	"errors"
	"testing"
)

func TestPrintErrors(t *testing.T) {
	if err := PrintErrors([]error{nil, nil, errors.New("1")}); err != nil {
		t.Fatal("should not return failure", err)
	}
	if err := PrintErrors([]error{nil, nil, nil}); err != nil {
		t.Fatal("should not return failure", err)
	}
	if err := PrintErrors([]error{}); err != nil {
		t.Fatal("should not return failure", err)
	}
	if err := PrintErrors([]error{errors.New("2"), errors.New("3")}); err == nil {
		t.Fatal("did not fail")
	}
	if err := PrintErrors([]error{errors.New("4")}); err == nil {
		t.Fatal("did not fail")
	}
}
