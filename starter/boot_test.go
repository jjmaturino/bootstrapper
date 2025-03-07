package starter

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jjmaturino/bootstrapper/platform"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"strings"
	"testing"
)

func TestServiceLauncher_Start(t *testing.T) {
	// Create context
	ctx := context.Background()

	// Create a test logger that writes to a buffer
	var logBuf bytes.Buffer
	testLogger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(
			c,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(&logBuf),
				zap.InfoLevel,
			),
		)
	})))

	// Create mock service starter
	mockStarter := &mockServiceStarter{
		startServiceFunc: func(ctx context.Context, service platform.Service, deps ...interface{}) error {
			return nil // Success by default
		},
	}

	// Create a service launcher with the test logger
	launcher := NewServiceLauncher(ctx, testLogger)

	// Register our mock starter for the VM platform
	launcher.RegisterPlatform(ctx, platform.VM, mockStarter)

	// Create mock service
	mockService := &mockService{}

	// Test Case 1: Service starter succeeds
	err := launcher.Start(ctx, mockService, platform.VM)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Verify the service starter was called
	if !mockStarter.startServiceCalled {
		t.Errorf("ServiceStarter.StartService was not called when platformType is 'VM'")
	}

	// Verify log contains expected info
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Starting service") {
		t.Errorf("Expected log to contain 'Starting service', but got: %s", logOutput)
	}

	// Reset
	mockStarter.startServiceCalled = false
	logBuf.Reset()

	// Test Case 2: Service starter fails
	expectedErr := fmt.Errorf("failed to start service")
	mockStarter.startServiceFunc = func(ctx context.Context, service platform.Service, deps ...interface{}) error {
		return expectedErr
	}

	err = launcher.Start(ctx, mockService, platform.VM)
	if err != expectedErr {
		t.Errorf("Expected error %v, but got: %v", expectedErr, err)
	}

	// Verify the service starter was called
	if !mockStarter.startServiceCalled {
		t.Errorf("ServiceStarter.StartService was not called when platformType is 'VM'")
	}

	// Reset
	mockStarter.startServiceCalled = false
	logBuf.Reset()

	// Test Case 3: Unsupported platform type
	unsupportedPlatformType := platform.Type("unsupported_type")
	err = launcher.Start(ctx, mockService, unsupportedPlatformType)

	if err == nil {
		t.Errorf("Expected error for unsupported platform type, but got nil")
	}

	expectedErrMsg := "unsupported platform type: unsupported_type"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrMsg, err.Error())
	}
}

func TestServiceLauncher_GetPlatformStarter(t *testing.T) {
	// Create context
	ctx := context.Background()

	// Create test logger
	testLogger := zaptest.NewLogger(t)

	// Create a service launcher
	launcher := NewServiceLauncher(ctx, testLogger)

	// Create mock service starter
	mockStarter := &mockServiceStarter{}

	// Register our mock starter
	launcher.RegisterPlatform(ctx, platform.VM, mockStarter)

	// Test Case 1: Get existing platform starter
	starter, err := launcher.GetPlatformStarter(platform.VM)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if starter != mockStarter {
		t.Errorf("Expected to get the registered starter")
	}

	// Test Case 2: Get non-existent platform starter
	nonExistentPlatform := platform.Type("non_existent")
	starter, err = launcher.GetPlatformStarter(nonExistentPlatform)

	if err == nil {
		t.Errorf("Expected error for non-existent platform type, but got nil")
	}

	if starter != nil {
		t.Errorf("Expected nil starter for non-existent platform, but got: %v", starter)
	}

	expectedErrMsg := "no starter registered for platform: non_existent"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrMsg, err.Error())
	}
}

func TestServiceLauncher_RegisterPlatform(t *testing.T) {
	// Create context
	ctx := context.Background()

	// Create a buffer to capture logs
	var logBuf bytes.Buffer
	testLogger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(
			c,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(&logBuf),
				zap.InfoLevel,
			),
		)
	})))

	// Create a service launcher
	launcher := NewServiceLauncher(ctx, testLogger)

	// Create mock service starters
	mockStarter1 := &mockServiceStarter{}
	mockStarter2 := &mockServiceStarter{}

	// Test Case 1: Register new platform starter
	launcher.RegisterPlatform(ctx, platform.VM, mockStarter1)

	// Verify the starter was registered
	starter, err := launcher.GetPlatformStarter(platform.VM)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if starter != mockStarter1 {
		t.Errorf("Expected to get the registered starter")
	}

	// Verify log contains expected info
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Registered platform starter") {
		t.Errorf("Expected log to contain 'Registered platform starter', but got: %s", logOutput)
	}

	logBuf.Reset()

	// Test Case 2: Override existing platform starter
	launcher.RegisterPlatform(ctx, platform.VM, mockStarter2)

	// Verify the starter was updated
	starter, err = launcher.GetPlatformStarter(platform.VM)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if starter != mockStarter2 {
		t.Errorf("Expected to get the updated starter")
	}

	// Verify log contains warning
	logOutput = logBuf.String()
	if !strings.Contains(logOutput, "Overriding existing platform starter") {
		t.Errorf("Expected log to contain 'Overriding existing platform starter', but got: %s", logOutput)
	}
}

// Mock implementations for testing

type mockServiceStarter struct {
	startServiceCalled bool
	startServiceFunc   func(ctx context.Context, service platform.Service, deps ...interface{}) error
}

func (m *mockServiceStarter) Start(ctx context.Context, service platform.Service, deps ...interface{}) error {
	m.startServiceCalled = true
	if m.startServiceFunc != nil {
		return m.startServiceFunc(ctx, service, deps...)
	}
	return nil
}

type mockService struct{}

func (m *mockService) Type() platform.ServiceType {
	return platform.ServiceType("mock-service")
}

func (m *mockService) Initialize(ctx context.Context, deps ...interface{}) error {
	return nil
}
