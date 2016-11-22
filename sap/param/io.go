package param

import (
	"github.com/HouzuoGuo/saptune/system"
	"io/ioutil"
	"path"
)

// Change IO elevators on all IO devices
type BlockDeviceSchedulers struct {
	SchedulerChoice map[string]string
}

func (ioe BlockDeviceSchedulers) Inspect() (Parameter, error) {
	newIOE := BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}
	// List /sys/block and inspect the IO elevator of each one
	dirContent, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return nil, err
	}
	for _, entry := range dirContent {
		// Remember, GetSysChoice does not accept the leading /sys/
		elev := system.GetSysChoice(path.Join("block", entry.Name(), "queue", "scheduler"))
		if elev != "" {
			newIOE.SchedulerChoice[entry.Name()] = elev
		}
	}
	return newIOE, nil
}
func (ioe BlockDeviceSchedulers) Optimise(newElevatorName interface{}) (Parameter, error) {
	newIOE := BlockDeviceSchedulers{SchedulerChoice: make(map[string]string)}
	for k := range ioe.SchedulerChoice {
		newIOE.SchedulerChoice[k] = newElevatorName.(string)
	}
	return newIOE, nil
}
func (ioe BlockDeviceSchedulers) Apply() error {
	for name, elevator := range ioe.SchedulerChoice {
		system.SetSysString(path.Join("block", name, "queue", "scheduler"), elevator)
	}
	return nil
}
