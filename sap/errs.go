package sap

import (
	"fmt"
	"log"
)

// PrintErrors prints out non-nil errors among the array. Returns an error if the array does not have nil element.
func PrintErrors(errors []error) error {
	hasNil := false
	for _, err := range errors {
		if err == nil {
			hasNil = true
		} else {
			log.Printf("%v", err)
		}
	}
	if hasNil {
		return nil
	}
	return fmt.Errorf("The tuning procedure failed entirely.")
}
