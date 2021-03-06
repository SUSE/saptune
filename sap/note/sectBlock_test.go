package note

import (
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"testing"
)

var blockDev = system.CollectBlockDeviceInfo()

func TestGetBlkVal(t *testing.T) {
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	_, _, err := GetBlkVal("IO_SCHEDULER_sda", &tblck)
	if err != nil {
		t.Error(err)
	}
}

func TestOptBlkVal(t *testing.T) {
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info := OptBlkVal("IO_SCHEDULER_sda", "noop", &tblck, blckOK)
	if val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info := OptBlkVal("IO_SCHEDULER_sda", "none", &tblck, blckOK)
		if val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}

	val, info = OptBlkVal("IO_SCHEDULER_sda", "NoOP", &tblck, blckOK)
	if val != "NoOP" && val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal("IO_SCHEDULER_sda", "NoNE", &tblck, blckOK)
		if val != "NoNE" && val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "deadline", &tblck, blckOK)
	if val != "deadline" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal("IO_SCHEDULER_sda", "mq-deadline", &tblck, blckOK)
		if val != "mq-deadline" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "NoOp,NoNe", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", " noop , none ", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "hugo", &tblck, blckOK)
	if val != "hugo" && info != "NA" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
	}

	val, info = OptBlkVal("NRREQ_sda", "512", &tblck, blckOK)
	if val != "512" {
		t.Error(val)
	}
	val, info = OptBlkVal("NRREQ_sdb", "0", &tblck, blckOK)
	if val != "1024" {
		t.Error(val)
	}
	val, info = OptBlkVal("NRREQ_sdc", "128", &tblck, blckOK)
	if val != "128" {
		t.Error(val)
	}
}

func TestSetBlkVal(t *testing.T) {
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info, err := GetBlkVal("IO_SCHEDULER_sda", &tblck)
	oval := val
	if err != nil {
		t.Error(err)
	}
	val, info = OptBlkVal("IO_SCHEDULER_sda", "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" {
		t.Error(val, info)
	}
	// apply - value not used, but map changed above in optimise
	err = SetBlkVal("IO_SCHEDULER_sda", "notUsed", &tblck, false)
	// revert - value will be used to change map before applying
	err = SetBlkVal("IO_SCHEDULER_sda", oval, &tblck, true)
}
