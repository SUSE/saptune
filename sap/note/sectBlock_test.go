package note

import (
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"testing"
)

var blockDev = system.CollectBlockDeviceInfo()
var tstDisk = "sda"

var setUp = func(t *testing.T) {
	t.Helper()
	_, bdevs := system.ListDir("/sys/block", "the available block devices of the system")
	for _, bdev := range bdevs {
		if bdev == "sda" {
			tstDisk = "sda"
			break
		}
		if bdev == "vda" {
			tstDisk = "vda"
			break
		}
	}
}

func TestGetBlkVal(t *testing.T) {
	setUp(t)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	_, _, err := GetBlkVal("IO_SCHEDULER_"+tstDisk, &tblck)
	if err != nil {
		t.Error(err)
	}
}

func TestOptBlkVal(t *testing.T) {
	setUp(t)
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info := OptBlkVal("IO_SCHEDULER_"+tstDisk, "noop", &tblck, blckOK)
	if val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info := OptBlkVal("IO_SCHEDULER_"+tstDisk, "none", &tblck, blckOK)
		if val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}

	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "NoOP", &tblck, blckOK)
	if val != "NoOP" && val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "NoNE", &tblck, blckOK)
		if val != "NoNE" && val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "deadline", &tblck, blckOK)
	if val != "deadline" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "mq-deadline", &tblck, blckOK)
		if val != "mq-deadline" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "NoOp,NoNe", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, " noop , none ", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "hugo", &tblck, blckOK)
	if val != "hugo" && info != "NA" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "", &tblck, blckOK)
	if val != "" || info != "" {
		t.Error(val, info)
	}

	val, _ = OptBlkVal("NRREQ_"+tstDisk, "512", &tblck, blckOK)
	if val != "512" {
		t.Error(val)
	}
	val, _ = OptBlkVal("NRREQ_sdc", "128", &tblck, blckOK)
	if val != "128" {
		t.Error(val)
	}
}

func TestSetBlkVal(t *testing.T) {
	setUp(t)
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info, err := GetBlkVal("IO_SCHEDULER_"+tstDisk, &tblck)
	oval := val
	if err != nil {
		t.Error(err, info)
	}
	val, info = OptBlkVal("IO_SCHEDULER_"+tstDisk, "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" {
		t.Error(val, info)
	}
	// apply - value not used, but map changed above in optimise
	_ = SetBlkVal("IO_SCHEDULER_"+tstDisk, "notUsed", &tblck, false)
	// revert - value will be used to change map before applying
	_ = SetBlkVal("IO_SCHEDULER_"+tstDisk, oval, &tblck, true)
}

func TestChkMaxHWsector(t *testing.T) {
	setUp(t)
	key := "MAX_SECTORS_KB_" + tstDisk
	val := "18446744073709551615"
	ival, sval, info := chkMaxHWsector(key, val)
	if info != "limited" {
		t.Errorf("expected info as 'limited', but got '%s' - '%+v' - '%+v'\n", info, ival, sval)
	}
}
