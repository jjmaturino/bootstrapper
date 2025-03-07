package platform

import "testing"

func TestType(t *testing.T) {
	pT := VM
	expected := "virtual_machine"

	if pT.String() != expected {
		t.Errorf("Expected %s, got %s", expected, sT.String())
	}
}

func TestServiceType(t *testing.T) {
	sT := HTTPServiceType
	expected := "http"

	if sT.String() != expected {
		t.Errorf("Expected %s, got %s", expected, sT.String())
	}
}
