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

func TestMax(t *testing.T) {
	if val := MaxU64(); val != 0 {
		t.Fatal(val)
	}
	if val := MaxU64(4, 3, 5, 2, 6); val != 6 {
		t.Fatal(val)
	}

	if val := MaxI64(); val != 0 {
		t.Fatal(val)
	}
	if val := MaxI64(4, 3, -5, -2, 6); val != 6 {
		t.Fatal(val)
	}
}

func TestMin(t *testing.T) {
	if val := MinU64(0); val != 0 {
		t.Fatal(val)
	}
	if val := MinU64(4, 3, 5, 2, 6); val != 2 {
		t.Fatal(val)
	}
}
