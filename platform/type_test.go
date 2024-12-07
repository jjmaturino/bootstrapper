package platform

import "testing"

func TestType(t *testing.T) {
	pT := VM
	expected := "virtual_machine"

	if pT.String() != expected {
		t.Errorf("Expected %s, got %s", "virtual_machine", pT.String())
	}
}
