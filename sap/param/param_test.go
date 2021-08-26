package param

import (
	"encoding/json"
	"reflect"
	"testing"
)

func jsonMarshalAndBack(original interface{}, receiver interface{}, t *testing.T) {
	serialised, err := json.Marshal(original)
	if err != nil {
		t.Fatal(original, err)
	}
	json.Unmarshal(serialised, &receiver)
}

func TestParamSerialisation(t *testing.T) {
	// All parameters must be tested here
	ioel := BlockDeviceSchedulers{SchedulerChoice: map[string]string{"a": "noop", "b": "deadline"}}
	newIoel := BlockDeviceSchedulers{}
	jsonMarshalAndBack(ioel, &newIoel, t)
	if !reflect.DeepEqual(newIoel, ioel) {
		t.Fatal(newIoel, ioel)
	}
}
