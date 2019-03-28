package note

import (
	"github.com/SUSE/saptune/sap/param"
)

// 2161991 - VMware vSphere (guest) configuration guidelines
type VmwareGuestIOElevator struct {
	BlockDeviceSchedulers param.BlockDeviceSchedulers
}

func (vmio VmwareGuestIOElevator) Name() string {
	return "VMware vSphere (guest) configuration guidelines"
}
func (vmio VmwareGuestIOElevator) Initialise() (Note, error) {
	inspectedParam, err := vmio.BlockDeviceSchedulers.Inspect()
	return VmwareGuestIOElevator{
		BlockDeviceSchedulers: inspectedParam.(param.BlockDeviceSchedulers),
	}, err
}
func (vmio VmwareGuestIOElevator) Optimise() (Note, error) {
	// SAP recommends noop for Vmware guests
	optimisedParam, err := vmio.BlockDeviceSchedulers.Optimise("noop")
	return VmwareGuestIOElevator{
		BlockDeviceSchedulers: optimisedParam.(param.BlockDeviceSchedulers),
	}, err
}
func (vmio VmwareGuestIOElevator) Apply() error {
	return vmio.BlockDeviceSchedulers.Apply()
}
