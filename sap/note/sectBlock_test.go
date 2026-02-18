package note

import (
	"os"
	"github.com/SUSE/saptune/sap/param"
	"github.com/SUSE/saptune/system"
	"testing"
)

var blockDev = system.CollectBlockDeviceInfo()

func TestGetBlkVal(t *testing.T) {
	tstScheduler := ""
	if _, err := os.Stat("/sys/block/sda"); err == nil {
		tstScheduler = "IO_SCHEDULER_sda"
	}
	if _, err := os.Stat("/sys/block/vda"); err == nil {
		tstScheduler = "IO_SCHEDULER_vda"
	}
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	_, _, err := GetBlkVal(tstScheduler, &tblck)
	if err != nil {
		t.Error(err)
	}
}

func TestOptBlkVal(t *testing.T) {
	tstScheduler := ""
	if _, err := os.Stat("/sys/block/sda"); err == nil {
		tstScheduler = "IO_SCHEDULER_sda"
	}
	if _, err := os.Stat("/sys/block/vda"); err == nil {
		tstScheduler = "IO_SCHEDULER_vda"
	}
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	val, info := OptBlkVal(tstScheduler, "noop", &tblck, blckOK)
	if val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info := OptBlkVal(tstScheduler, "none", &tblck, blckOK)
		if val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}

	val, info = OptBlkVal(tstScheduler, "NoOP", &tblck, blckOK)
	if val != "NoOP" && val != "noop" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal(tstScheduler, "NoNE", &tblck, blckOK)
		if val != "NoNE" && val != "none" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal(tstScheduler, "deadline", &tblck, blckOK)
	if val != "deadline" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
		val, info = OptBlkVal(tstScheduler, "mq-deadline", &tblck, blckOK)
		if val != "mq-deadline" {
			t.Error(val, info)
		}
		if info == "NA" {
			t.Logf("scheduler '%s' is not supported\n", val)
		}
	}
	val, info = OptBlkVal(tstScheduler, "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal(tstScheduler, "NoOp,NoNe", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal(tstScheduler, " noop , none ", &tblck, blckOK)
	if val != "noop" && val != "none" && info != "NA" {
		t.Error(val, info)
	}
	val, info = OptBlkVal(tstScheduler, "hugo", &tblck, blckOK)
	if val != "hugo" && info != "NA" {
		t.Error(val, info)
	}
	if info == "NA" {
		t.Logf("scheduler '%s' is not supported\n", val)
	}
	val, info = OptBlkVal(tstScheduler, "", &tblck, blckOK)
	if val != "" || info != "" {
		t.Error(val, info)
	}

	val, _ = OptBlkVal("NRREQ_sda", "512", &tblck, blckOK)
	if val != "512" {
		t.Error(val)
	}
	val, _ = OptBlkVal("NRREQ_sdc", "128", &tblck, blckOK)
	if val != "128" {
		t.Error(val)
	}
}

func TestSetBlkVal(t *testing.T) {
	tstScheduler := ""
	if _, err := os.Stat("/sys/block/sda"); err == nil {
		tstScheduler = "IO_SCHEDULER_sda"
	}
	if _, err := os.Stat("/sys/block/vda"); err == nil {
		tstScheduler = "IO_SCHEDULER_vda"
	}
	blckOK := make(map[string][]string)
	tblck := param.BlockDeviceQueue{BlockDeviceSchedulers: param.BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}, BlockDeviceNrRequests: param.BlockDeviceNrRequests{NrRequests: make(map[string]int)}}
	
	val, info, err := GetBlkVal(tstScheduler, &tblck)
	oval := val
	if err != nil {
		t.Error(err, info)
	}
	val, info = OptBlkVal(tstScheduler, "noop, none", &tblck, blckOK)
	if val != "noop" && val != "none" {
		t.Error(val, info)
	}
	// apply - value not used, but map changed above in optimise
	_ = SetBlkVal(tstScheduler, "notUsed", &tblck, false)
	// revert - value will be used to change map before applying
	_ = SetBlkVal(tstScheduler, oval, &tblck, true)
}

func TestChkMaxHWsector(t *testing.T) {
	key := "MAX_SECTORS_KB_sda"
	val := "18446744073709551615"
	ival, sval, info := chkMaxHWsector(key, val)
	if info != "limited" {
		t.Errorf("expected info as 'limited', but got '%s' - '%+v' - '%+v'\n", info, ival, sval)
	}
}
