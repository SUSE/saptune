package param

import (
	"github.com/SUSE/saptune/system"
	"io/ioutil"
	"path"
	"testing"
)

// ANGI TODO - check the results for optimised=applied and oldvals=reverted

var blockDev = system.CollectBlockDeviceInfo()

func TestIOElevators(t *testing.T) {
	bdev := "sda"
	scheduler := ""
	inspected, err := BlockDeviceSchedulers{}.Inspect()
	if err != nil {
		t.Error(err, inspected)
	}
	t.Logf("inspected - '%+v'\n", inspected)
	if len(inspected.(BlockDeviceSchedulers).SchedulerChoice) == 0 {
		t.Skip("the test case will not continue because inspection result turns out empty")
	}
	for name, elevator := range inspected.(BlockDeviceSchedulers).SchedulerChoice {
		if name == "" || elevator == "" {
			t.Error(inspected)
		}
	}
	oldvals := BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}
	t.Logf("oldvals - '%+v'\n", oldvals)
	for name, elevator := range inspected.(BlockDeviceSchedulers).SchedulerChoice {
		oldvals.SchedulerChoice[name] = elevator
	}
	t.Logf("oldvals - '%+v'\n", oldvals)

	// ANGI TODO - better solution
	_, err = ioutil.ReadDir("/sys/block/sda/mq")
	if err != nil {
		// single queue scheduler (values: noop deadline cfq)
		scheduler = "noop"
	} else {
		// multi queue scheduler (values: mq-deadline kyber bfq none)
		scheduler = "none"
	}

	optVal := bdev + " " + scheduler
	optimised, err := inspected.Optimise(optVal)
	if err != nil {
		t.Error(err)
	}
	t.Logf("optimised - '%+v'\n", optimised)
	if len(optimised.(BlockDeviceSchedulers).SchedulerChoice) == 0 {
		t.Error(optimised)
	}
	for name, elevator := range optimised.(BlockDeviceSchedulers).SchedulerChoice {
		if name == "" || (name == bdev && elevator != scheduler) {
			t.Error(optimised)
		}
	}
	// apply
	err = optimised.Apply(bdev)
	if err != nil {
		t.Error(err)
	}

	// refresh block device information
	_ = system.CollectBlockDeviceInfo()
	blkDev = &system.BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	// check
	applied, err := BlockDeviceSchedulers{}.Inspect()
	if err != nil {
		t.Error(err, applied)
	}
	t.Logf("applied - '%+v'\n", applied)
	if len(applied.(BlockDeviceSchedulers).SchedulerChoice) == 0 {
		t.Log("inspection result turns out empty")
	}
	for name, elevator := range applied.(BlockDeviceSchedulers).SchedulerChoice {
		if name == "" || (name == bdev && elevator != scheduler) {
			t.Error(applied)
		}
	}

	// reset original values
	t.Logf("oldvals - '%+v'\n", oldvals)
	err = oldvals.Apply(bdev)
	if err != nil {
		t.Error(err)
	}
	// refresh block device information
	_ = system.CollectBlockDeviceInfo()
	blkDev = &system.BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	rev, _ := BlockDeviceSchedulers{}.Inspect()
	t.Logf("reverted - '%+v'\n", rev)
}

func TestNrRequests(t *testing.T) {
	inspected, err := BlockDeviceNrRequests{}.Inspect()
	if err != nil {
		t.Error(err, inspected)
	}
	t.Logf("inspected - '%+v'\n", inspected)
	if len(inspected.(BlockDeviceNrRequests).NrRequests) == 0 {
		t.Skip("the test case will not continue because inspection result turns out empty")
	}
	for name, nrrequest := range inspected.(BlockDeviceNrRequests).NrRequests {
		if name == "" || nrrequest < 0 {
			t.Error(inspected)
		}
	}
	oldvals := BlockDeviceNrRequests{NrRequests: make(map[string]int)}
	t.Logf("oldvals - '%+v'\n", oldvals)
	for name, nrrequest := range inspected.(BlockDeviceNrRequests).NrRequests {
		oldvals.NrRequests[name] = nrrequest
	}
	t.Logf("oldvals - '%+v'\n", oldvals)
	optimised, err := inspected.Optimise(32)
	if err != nil {
		t.Error(err)
	}
	t.Logf("optimised - '%+v'\n", optimised)
	if len(optimised.(BlockDeviceNrRequests).NrRequests) == 0 {
		t.Error(optimised)
	}
	for name, nrrequest := range optimised.(BlockDeviceNrRequests).NrRequests {
		if name == "" || nrrequest < 0 {
			t.Error(optimised)
		}
	}
	// apply
	for _, bdev := range blockDev {
		err = optimised.Apply(bdev)
		if err != nil {
			t.Error(err)
		}
	}

	// refresh block device information
	_ = system.CollectBlockDeviceInfo()
	blkDev = &system.BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	// check
	applied, err := BlockDeviceNrRequests{}.Inspect()
	if err != nil {
		t.Error(err, applied)
	}
	t.Logf("applied - '%+v'\n", applied)
	if len(applied.(BlockDeviceNrRequests).NrRequests) == 0 {
		t.Log("inspection result turns out empty")
	}
	for name, nrrequest := range applied.(BlockDeviceNrRequests).NrRequests {
		if name == "" || nrrequest != 32 {
			t.Error(applied)
		}
	}

	// reset original values
	for _, bdev := range blockDev {
		err = oldvals.Apply(bdev)
		if err != nil {
			t.Error(err)
		}
	}

	// refresh block device information
	_ = system.CollectBlockDeviceInfo()
	blkDev = &system.BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	rev, _ := BlockDeviceNrRequests{}.Inspect()
	t.Logf("reverted - '%+v'\n", rev)
}

func TestReadAheadKB(t *testing.T) {
	inspected, err := BlockDeviceReadAheadKB{}.Inspect()
	if err != nil {
		t.Error(err, inspected)
	}
	t.Logf("inspected - '%+v'\n", inspected)
	if len(inspected.(BlockDeviceReadAheadKB).ReadAheadKB) == 0 {
		t.Skip("the test case will not continue because inspection result turns out empty")
	}
	for name, readaheadkb := range inspected.(BlockDeviceReadAheadKB).ReadAheadKB {
		if name == "" || readaheadkb < 0 {
			t.Error(inspected)
		}
	}
	oldvals := BlockDeviceReadAheadKB{ReadAheadKB: make(map[string]int)}
	t.Logf("oldvals - '%+v'\n", oldvals)
	for name, readaheadkb := range inspected.(BlockDeviceReadAheadKB).ReadAheadKB {
		oldvals.ReadAheadKB[name] = readaheadkb
	}
	t.Logf("oldvals - '%+v'\n", oldvals)
	optimised, err := inspected.Optimise(132)
	if err != nil {
		t.Error(err)
	}
	t.Logf("optimised - '%+v'\n", optimised)
	if len(optimised.(BlockDeviceReadAheadKB).ReadAheadKB) == 0 {
		t.Error(optimised)
	}
	for name, readaheadkb := range optimised.(BlockDeviceReadAheadKB).ReadAheadKB {
		if name == "" || readaheadkb < 0 {
			t.Error(optimised)
		}
	}
	// apply
	for _, bdev := range blockDev {
		err = optimised.Apply(bdev)
		if err != nil {
			t.Error(err)
		}
	}

	// refresh block device information
	_ = system.CollectBlockDeviceInfo()
	blkDev = &system.BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	// check
	applied, err := BlockDeviceReadAheadKB{}.Inspect()
	if err != nil {
		t.Error(err, applied)
	}
	t.Logf("applied - '%+v'\n", applied)
	if len(applied.(BlockDeviceReadAheadKB).ReadAheadKB) == 0 {
		t.Log("inspection result turns out empty")
	}
	for name, readaheadkb := range applied.(BlockDeviceReadAheadKB).ReadAheadKB {
		if name == "" || readaheadkb != 132 {
			t.Error(applied)
		}
	}

	// reset original values
	for _, bdev := range blockDev {
		err = oldvals.Apply(bdev)
		if err != nil {
			t.Error(err)
		}
	}

	// refresh block device information
	_ = system.CollectBlockDeviceInfo()
	blkDev = &system.BlockDev{
		AllBlockDevs:    make([]string, 0, 64),
		BlockAttributes: make(map[string]map[string]string),
	}

	rev, _ := BlockDeviceReadAheadKB{}.Inspect()
	t.Logf("reverted - '%+v'\n", rev)
}

func TestIsValidScheduler(t *testing.T) {
	scheduler := ""
	dirCont, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		t.Skip("no block files available. Skip test.")
	}
	for _, entry := range dirCont {
		_, err := ioutil.ReadDir(path.Join("/sys/block/", entry.Name(), "mq"))
		if err != nil {
			// single queue scheduler (values: noop deadline cfq)
			scheduler = "cfq"
		} else {
			// multi queue scheduler (values: mq-deadline kyber bfq none)
			scheduler = "none"
		}
		if entry.Name() == "sda" {
			if !IsValidScheduler("sda", scheduler) {
				t.Errorf("'%s' is not a valid scheduler for 'sda'\n", scheduler)
			}
			if IsValidScheduler("sda", "hugo") {
				t.Error("'hugo' is a valid scheduler for 'sda'")
			}
		}
		if entry.Name() == "vda" {
			if !IsValidScheduler("vda", scheduler) {
				t.Errorf("'%s' is not a valid scheduler for 'vda'\n", scheduler)
			}
			if IsValidScheduler("vda", "hugo") {
				t.Error("'hugo' is a valid scheduler for 'vda'")
			}
		}
	}
}

func TestIsValidforNrRequests(t *testing.T) {
	dirCont, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		t.Skip("no block files available. Skip test.")
	}
	for _, entry := range dirCont {
		if entry.Name() == "sda" {
			if !IsValidforNrRequests("sda", "1024") {
				t.Log("'1024' is not a valid number of requests for 'sda'")
			} else {
				t.Log("'1024' is a valid number of requests for 'sda'")
			}
			if !IsValidforNrRequests("sda", "32") {
				t.Log("'32' is not a valid number of requests for 'sda'")
			} else {
				t.Log("'32' is a valid number of requests for 'sda'")
			}
		}
		if entry.Name() == "vda" {
			if !IsValidforNrRequests("vda", "128") {
				t.Log("'128' is not a valid number of requests for 'vda'")
			} else {
				t.Log("'128' is a valid number of requests for 'vda'")
			}
		}
	}
}

func TestIsValidforReadAheadKB(t *testing.T) {
	dirCont, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		t.Skip("no block files available. Skip test.")
	}
	for _, entry := range dirCont {
		if entry.Name() == "sda" {
			if !IsValidforReadAheadKB("sda", "1024") {
				t.Log("'1024' is not a valid number of requests for 'sda'")
			} else {
				t.Log("'1024' is a valid number of requests for 'sda'")
			}
			if !IsValidforReadAheadKB("sda", "132") {
				t.Log("'132' is not a valid number of requests for 'sda'")
			} else {
				t.Log("'132' is a valid number of requests for 'sda'")
			}
			if !IsValidforReadAheadKB("sda", "133") {
				t.Log("'133' is not a valid number of requests for 'sda'")
			} else {
				t.Log("'133' is a valid number of requests for 'sda'")
			}
		}
		if entry.Name() == "vda" {
			if !IsValidforNrRequests("vda", "128") {
				t.Log("'128' is not a valid number of requests for 'vda'")
			} else {
				t.Log("'128' is a valid number of requests for 'vda'")
			}
		}
	}
}

// Apply für beide
