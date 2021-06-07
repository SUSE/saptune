package note

import (
	"testing"
)

//GetLimitsVal
func TestOptLimitsVal(t *testing.T) {
	val := OptLimitsVal("@sdba soft nofile NA", "@sdba soft nofile 32800")
	if val != "@sdba soft nofile 32800" {
		t.Error(val)
	}
	val = OptLimitsVal("@sdba soft nofile 75536", "@sdba soft nofile 32800")
	if val != "@sdba soft nofile 32800" {
		t.Error(val)
	}
}

//SetLimitsVal apply and revert
