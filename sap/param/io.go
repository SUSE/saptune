package param

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"path"
	"strconv"
	"strings"
)

// BlockDeviceQueue is the data structure for block devices
// for schedulers, IO nr_request and read_ahead_kb changes
type BlockDeviceQueue struct {
	BlockDeviceSchedulers
	BlockDeviceNrRequests
	BlockDeviceReadAheadKB
}

var blkDev *system.BlockDev

// BlockDeviceSchedulers changes IO elevators on all IO devices
type BlockDeviceSchedulers struct {
	SchedulerChoice map[string]string
}

// Inspect retrieves the current scheduler from the system
func (ioe BlockDeviceSchedulers) Inspect() (Parameter, error) {
	if len(ioe.SchedulerChoice) != 0 {
		// inspect needs to run only once per saptune call
		return ioe, nil
	}
	if blkDev == nil || (len(blkDev.AllBlockDevs) == 0 && len(blkDev.BlockAttributes) == 0) {
		blkDev, _ = system.GetBlockDeviceInfo()
	}
	newIOE := BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}
	for _, entry := range blkDev.AllBlockDevs {
		elev := blkDev.BlockAttributes[entry]["IO_SCHEDULER"]
		if elev != "" {
			newIOE.SchedulerChoice[entry] = elev
		}
	}
	return newIOE, nil
}

// Optimise gets the expected scheduler value from the configuration
func (ioe BlockDeviceSchedulers) Optimise(newElevatorName interface{}) (Parameter, error) {
	newIOE := ioe
	fields := strings.Fields(newElevatorName.(string))
	if len(fields) > 1 {
		bdev := fields[0]
		newSched := fields[1]
		newIOE.SchedulerChoice[bdev] = newSched
		/* Future
		if bdev == "all" {
			// all devices with same scheduler
			for k := range ioe.SchedulerChoice {
				if !IsValidScheduler(k, newSched) {
					continue
				}
				newIOE.SchedulerChoice[k] = newSched
			}
		} else {
			if IsValidScheduler(bdev, newSched) {
				newIOE.SchedulerChoice[bdev] = newSched
			}
		}
		*/
	}
	return newIOE, nil
}

// Apply sets the new scheduler value in the system
func (ioe BlockDeviceSchedulers) Apply(blkdev interface{}) error {
	//errs := make([]error, 0, 0)
	bdev := blkdev.(string)
	elevator := ioe.SchedulerChoice[bdev]
	err := system.SetSysString(path.Join("block", bdev, "queue", "scheduler"), elevator)

	/* reuse in future
	for name, elevator := range ioe.SchedulerChoice {
		errs = append(errs, system.SetSysString(path.Join("block", name, "queue", "scheduler"), elevator))
	}
	err := sap.PrintErrors(errs)
	*/
	return err
}

// BlockDeviceNrRequests changes IO nr_requests on all block devices
type BlockDeviceNrRequests struct {
	NrRequests map[string]int
}

// Inspect retrieves the current nr_requests from the system
func (ior BlockDeviceNrRequests) Inspect() (Parameter, error) {
	if len(ior.NrRequests) != 0 {
		// inspect needs to run only once per saptune call
		return ior, nil
	}
	if blkDev == nil || (len(blkDev.AllBlockDevs) == 0 && len(blkDev.BlockAttributes) == 0) {
		blkDev, _ = system.GetBlockDeviceInfo()
	}
	newIOR := BlockDeviceNrRequests{NrRequests: make(map[string]int)}
	for _, entry := range blkDev.AllBlockDevs {
		nrreq := blkDev.BlockAttributes[entry]["NRREQ"]
		if nrreq != "" {
			ival, _ := strconv.Atoi(nrreq)
			if ival >= 0 {
				newIOR.NrRequests[entry] = ival
			}
		}
	}
	return newIOR, nil
}

// Optimise gets the expected nr_requests value from the configuration
func (ior BlockDeviceNrRequests) Optimise(newNrRequestValue interface{}) (Parameter, error) {
	newIOR := BlockDeviceNrRequests{NrRequests: make(map[string]int)}
	for k := range ior.NrRequests {
		newIOR.NrRequests[k] = newNrRequestValue.(int)
	}
	return newIOR, nil
}

// Apply sets the new nr_requests value in the system
func (ior BlockDeviceNrRequests) Apply(blkdev interface{}) error {
	bdev := blkdev.(string)
	nrreq := ior.NrRequests[bdev]
	err := system.SetSysInt(path.Join("block", bdev, "queue", "nr_requests"), nrreq)
	if err != nil {
		system.WarningLog("skipping device '%s', not valid for setting 'number of requests' to '%v'", bdev, nrreq)
	}
	/* for future use
	errs := make([]error, 0, 0)
	for name, nrreq := range ior.NrRequests {
		if !IsValidforNrRequests(name, strconv.Itoa(nrreq)) {
			system.WarningLog("skipping device '%s', not valid for setting 'number of requests' to '%v'", name, nrreq)
			continue
		}
		errs = append(errs, system.SetSysInt(path.Join("block", name, "queue", "nr_requests"), nrreq))
	}
	err := sap.PrintErrors(errs)
	return err
	*/
	return nil
}

// BlockDeviceReadAheadKB changes the read_ahead_kb value on all block devices
type BlockDeviceReadAheadKB struct {
	ReadAheadKB map[string]int
}

// Inspect retrieves the current read_ahead_kb from the system
func (rakb BlockDeviceReadAheadKB) Inspect() (Parameter, error) {
	if len(rakb.ReadAheadKB) != 0 {
		// inspect needs to run only once per saptune call
		return rakb, nil
	}
	if blkDev == nil || (len(blkDev.AllBlockDevs) == 0 && len(blkDev.BlockAttributes) == 0) {
		blkDev, _ = system.GetBlockDeviceInfo()
	}
	newRAKB := BlockDeviceReadAheadKB{ReadAheadKB: make(map[string]int)}
	for _, entry := range blkDev.AllBlockDevs {
		readahead := blkDev.BlockAttributes[entry]["READ_AHEAD_KB"]
		if readahead != "" {
			ival, _ := strconv.Atoi(readahead)
			if ival >= 0 {
				newRAKB.ReadAheadKB[entry] = ival
			}
		}
	}
	return newRAKB, nil
}

// Optimise gets the expected read_ahead_kb value from the configuration
func (rakb BlockDeviceReadAheadKB) Optimise(newReadAheadKBValue interface{}) (Parameter, error) {
	newRAKB := BlockDeviceReadAheadKB{ReadAheadKB: make(map[string]int)}
	for k := range rakb.ReadAheadKB {
		newRAKB.ReadAheadKB[k] = newReadAheadKBValue.(int)
	}
	return newRAKB, nil
}

// Apply sets the new read_ahead_kb value in the system
func (rakb BlockDeviceReadAheadKB) Apply(blkdev interface{}) error {
	bdev := blkdev.(string)
	readahead := rakb.ReadAheadKB[bdev]
	err := system.SetSysInt(path.Join("block", bdev, "queue", "read_ahead_kb"), readahead)
	if err != nil {
		system.WarningLog("skipping device '%s', not valid for setting 'read_ahead_kb' to '%v'", bdev, readahead)
	}
	/* for future use
	errs := make([]error, 0, 0)
	for name, readahead := range rakb.ReadAheadKB {
		if !IsValidforReadAheadKB(name, strconv.Itoa(readahead)) {
			system.WarningLog("skipping device '%s', not valid for setting 'read_ahead_kb' to '%v'", name, readahead)
			continue
		}
		errs = append(errs, system.SetSysInt(path.Join("block", name, "queue", "read_ahead_kb"), readahead))
	}
	err := sap.PrintErrors(errs)
	return err
	*/
	return nil
}

// IsValidScheduler checks, if the scheduler value is supported by the system.
// only used during optimize
// During initialize, the scheduler is read from the system, so no check needed.
// Only needed during optimize, as apply is using the value from optimize and
// revert is using the stored valid old values from before apply.
// And a scheduler can only change during a system reboot
// (single-queued -> multi-queued)
func IsValidScheduler(blockdev, scheduler string) bool {
	if blkDev == nil || (len(blkDev.AllBlockDevs) == 0 && len(blkDev.BlockAttributes) == 0) {
		blkDev, _ = system.GetBlockDeviceInfo()
	}
	val := blkDev.BlockAttributes[blockdev]["VALID_SCHEDS"]
	actsched := fmt.Sprintf("[%s]", scheduler)
	if val != "" {
		for _, s := range strings.Split(string(val), " ") {
			s = strings.TrimSpace(s)
			if s == scheduler || s == actsched {
				return true
			}
		}
	}
	system.WarningLog("'%s' is not a valid scheduler for device '%s', skipping.", scheduler, blockdev)
	return false
}

// IsValidforNrRequests checks, if the nr_requests value is supported by the system
// it's not a good idea to use this during optimize, as it will write a new
// value to the device, so only used during apply, but this can be performed
// in a better way.
func IsValidforNrRequests(blockdev, nrreq string) bool {
	elev, _ := system.GetSysChoice(path.Join("block", blockdev, "queue", "scheduler"))
	if elev != "" {
		file := path.Join("block", blockdev, "queue", "nr_requests")
		if tstErr := system.TestSysString(file, nrreq); tstErr == nil {
			return true
		}
	}
	return false
}

// IsValidforReadAheadKB checks, if the read_ahead_kb value is supported by the system
// it's not a good idea to use this during optimize, as it will write a new
// value to the device, so only used during apply, but this can be performed
// in a better way.
func IsValidforReadAheadKB(blockdev, readahead string) bool {
	elev, _ := system.GetSysChoice(path.Join("block", blockdev, "queue", "scheduler"))
	if elev != "" {
		file := path.Join("block", blockdev, "queue", "read_ahead_kb")
		if tstErr := system.TestSysString(file, readahead); tstErr == nil {
			return true
		}
	}
	return false
}
