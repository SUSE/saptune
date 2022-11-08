package note

import (
	"testing"
)

func TestGetLimitsVal(t *testing.T) {
	val, info, err := GetLimitsVal("@sdba soft nofile")
	if val != "@sdba soft nofile NA" || info != "" || err != nil {
		t.Error(val, info, err)
	}
	val, info, err = GetLimitsVal("@sdba soft")
	if val != "" || info != "" || err == nil {
		t.Error(val, info, err)
	}
}

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
