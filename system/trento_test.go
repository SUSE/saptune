package system

import (
	"os"
	"path"
	"testing"
)

func TestCheckAndSetTrento(t *testing.T) {
	configVal := "600"
	err := CheckAndSetTrento("TrentoASDP", configVal, true)
	if err == nil {
		t.Errorf("Test reports success, but should fail")
	}

	oldTrentoAgentFile := trentoAgentFile
	defer func() { trentoAgentFile = oldTrentoAgentFile }()
	trentoAgentFile = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/trento_agent.yaml")
	configVal = "600"
	err = CheckAndSetTrento("TrentoASDP", configVal, true)
	if err != nil {
		t.Errorf("Test reports error, but should succeed")
	}
	configVal = "600"
	err = CheckAndSetTrento("TrentoASDP", configVal, false)
	if err != nil {
		t.Errorf("Test reports error, but should succeed")
	}
	configVal = "300"
	err = CheckAndSetTrento("TrentoASDP", configVal, false)
	if err != nil {
		t.Errorf("Test reports error, but should succeed")
	}
	configVal = "300"
	err = CheckAndSetTrento("TrentoASDP", configVal, true)
	if err != nil {
		t.Errorf("Test reports error, but should succeed")
	}
	configVal = "off"
	err = CheckAndSetTrento("TrentoASDP", configVal, false)
	if err != nil {
		t.Errorf("Test reports error, but should succeed")
	}
	configVal = "700"
	err = CheckAndSetTrento("TrentoASDP", configVal, false)
	if err == nil {
		t.Errorf("Test reports success, but should fail")
	}
	trentoAgentFile = oldTrentoAgentFile
}
