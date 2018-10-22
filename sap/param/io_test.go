package param

import (
	"github.com/SUSE/saptune/system"
	"testing"
)

func TestIOElevators(t *testing.T) {
	if !system.IsUserRoot() {
		t.Skip("the test requires root access")
	}
	inspected, err := BlockDeviceSchedulers{}.Inspect()
	if err != nil {
		t.Fatal(err, inspected)
	}
	if len(inspected.(BlockDeviceSchedulers).SchedulerChoice) == 0 {
		t.Skip("the test case will not continue because inspection result turns out empty")
	}
	for name, elevator := range inspected.(BlockDeviceSchedulers).SchedulerChoice {
		if name == "" || elevator == "" {
			t.Fatal(inspected)
		}
	}
	optimised, err := inspected.Optimise("noop")
	if err != nil {
		t.Fatal(err)
	}
	if len(optimised.(BlockDeviceSchedulers).SchedulerChoice) == 0 {
		t.Fatal(optimised)
	}
	for name, elevator := range optimised.(BlockDeviceSchedulers).SchedulerChoice {
		if name == "" || elevator != "noop" {
			t.Fatal(optimised)
		}
	}
}

func TestNrRequests(t *testing.T) {
	if !system.IsUserRoot() {
		t.Skip("the test requires root access")
	}
	inspected, err := BlockDeviceNrRequests{}.Inspect()
	if err != nil {
		t.Fatal(err, inspected)
	}
	if len(inspected.(BlockDeviceNrRequests).NrRequests) == 0 {
		t.Skip("the test case will not continue because inspection result turns out empty")
	}
	for name, nrrequest := range inspected.(BlockDeviceNrRequests).NrRequests {
		if name == "" || nrrequest < 0 {
			t.Fatal(inspected)
		}
	}
	optimised, err := inspected.Optimise(128)
	if err != nil {
		t.Fatal(err)
	}
	if len(optimised.(BlockDeviceNrRequests).NrRequests) == 0 {
		t.Fatal(optimised)
	}
	for name, nrrequest := range optimised.(BlockDeviceNrRequests).NrRequests {
		if name == "" || nrrequest < 0 {
			t.Fatal(optimised)
		}
	}
}
