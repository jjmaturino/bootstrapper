package launcher

import (
	"bytes"
	"fmt"
	"github.com/jjmaturino/bootstrapper/platform"
	"log"
	"strings"
	"testing"
)

func TestStart(t *testing.T) {
	// Save the original platform.StartVM
	originalStartVM := platform.StartVM
	defer func() {
		StartVM = originalStartVM
	}()

	// Mock implementation of StartVM
	var startVMCalled bool
	StartVM = func(service platform.HTTPService, engine platform.Engine, deps ...interface{}) error {
		startVMCalled = true
		return nil // Simulate success
	}

	// Create mock implementations
	mockService := &platform.MockGinService{}
	mockEngine := &platform.MockEngineSuccess{}

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(nil) // Restore default output
	}()

	// Test Case 1: StartVM succeeds
	platformType := platform.VM
	Start(mockService, platformType, mockEngine)

	// Verify that StartVM was called
	if !startVMCalled {
		t.Errorf("platform.StartVM was not called when platformType is 'virtual_machine'")
	}

	// Verify that no error was logged
	if buf.Len() != 0 {
		t.Errorf("Expected no logs, but got: %s", buf.String())
	}

	// Reset variables
	startVMCalled = false
	buf.Reset()

	// Test Case 2: StartVM fails
	StartVM = func(service platform.HTTPService, engine platform.Engine, deps ...interface{}) error {
		startVMCalled = true
		return fmt.Errorf("failed to start VM")
	}

	Start(mockService, platformType, mockEngine)

	// Verify that platform.StartVM was called
	if !startVMCalled {
		t.Errorf("platform.StartVM was not called when platformType is 'virtual_machine'")
	}

	// Verify that the error was logged
	logged := buf.String()
	expectedLogMessage := "failed to start vm: failed to start VM"
	if !strings.Contains(logged, expectedLogMessage) {
		t.Errorf("Expected log message '%s', but got '%s'", expectedLogMessage, logged)
	}

	// Reset variables
	startVMCalled = false
	buf.Reset()

	// Test Case 3: Unsupported platform type
	unsupportedPlatformType := platform.Type("unsupported_type")
	Start(mockService, unsupportedPlatformType, mockEngine)

	// Verify that platform.StartVM was not called
	if startVMCalled {
		t.Errorf("platform.StartVM was called when platformType is unsupported")
	}

	// Verify that the unsupported platform type was logged
	logged = buf.String()
	expectedLogMessage = "unsupported platform type unsupported_type"
	if !strings.Contains(logged, expectedLogMessage) {
		t.Errorf("Expected log message '%s', but got '%s'", expectedLogMessage, logged)
	}
}
