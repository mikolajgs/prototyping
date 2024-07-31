package struct2db

import (
	"log"
	"testing"
)

// TestGetModelIDInterface tests if GetModelIDInterface return pointer to ID field
func TestGetModelIDInterface(t *testing.T) {
	ts := &TestStruct{}
	ts.ID = 123
	i := testController.GetModelIDInterface(ts)
	if *(i.(*int64)) != int64(123) {
		log.Fatalf("GetModelIDInterface failed to get ID")
	}
}

// TestGetModelIDValue tests if GetModelIDValue returns values of the ID field
func TestGetModelIDValue(t *testing.T) {
	ts := &TestStruct{}
	ts.ID = 123
	v := testController.GetModelIDValue(ts)
	if v != 123 {
		log.Fatalf("GetModelIDValue failed to get ID")
	}
}

// TestGetModelFieldInterfaces tests if GetModelFieldInterfaces returns interfaces to object fields
func TestGetModelFieldInterfaces(t *testing.T) {
	// TODO
}

// TestResetFields tests if ResetFields zeroes object fields
func TestResetFields(t *testing.T) {
	// TODO
}
