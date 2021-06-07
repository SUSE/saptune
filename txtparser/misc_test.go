package txtparser

import (
	"testing"
)

func TestCalculateOptimumValue(t *testing.T) {
	if val, err := CalculateOptimumValue(OperatorMoreThan, "21", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorMoreThan, "10", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorMoreThan, "", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(OperatorMoreThanEqual, "21", "20"); val != "21" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorMoreThanEqual, "20", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorMoreThanEqual, "10", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorMoreThanEqual, "", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(OperatorLessThan, "10", "20"); val != "10" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorLessThan, "10", "10"); val != "9" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorLessThan, "", "10"); val != "9" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(OperatorLessThanEqual, "10", "8"); val != "8" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorLessThanEqual, "10", "20"); val != "10" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorLessThanEqual, "10", "10"); val != "10" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorLessThanEqual, "", "10"); val != "10" || err != nil {
		t.Error(val, err)
	}

	if val, err := CalculateOptimumValue(OperatorEqual, "21", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorEqual, "10", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
	if val, err := CalculateOptimumValue(OperatorEqual, "", "20"); val != "20" || err != nil {
		t.Error(val, err)
	}
}
