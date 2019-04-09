package sap

import (
	"fmt"
	//"github.com/SUSE/saptune/system"
	"log"
)

// PrintErrors prints out non-nil errors among the array. Returns an error if the array does not have nil element.
func PrintErrors(errors []error) error {
	hasNil := len(errors) == 0
	for _, err := range errors {
		if err == nil {
			hasNil = true
		} else {
			log.Printf("%v", err)
			//txt := fmt.Sprintf("%v", err)
			//system.ErrorLog(txt)
		}
	}
	if hasNil {
		return nil
	}
	return fmt.Errorf("the tuning procedure failed entirely")
}
