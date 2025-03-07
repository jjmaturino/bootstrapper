package platform

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

func (m *MockService) Initialize(ctx context.Context, deps ...interface{}) error {
	args := m.Called(ctx, deps)
	return args.Error(0)
}

func (m *MockService) Type() ServiceType {
	args := m.Called()
	return args.Get(0).(ServiceType)
}

// MockHTTPService is a mock implementation of the HTTPService interface
type MockHTTPService struct {
	MockService
}

func (m *MockHTTPService) ConfigureRoutes(ctx context.Context, engine Engine) error {
	args := m.Called(ctx, engine)
	return args.Error(0)
}

// MockEngine is a mock implementation of the Engine interface
type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) Run(addr ...string) error {
	args := m.Called(addr)
	return args.Error(0)
}

func (m *MockEngine) Handle(method, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(method, relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func TestNewVMServiceStarter(t *testing.T) {
	tests := []struct {
		name   string
		logger *zap.Logger
		want   *VMServiceStarter
	}{
		{
			name:   "with logger",
			logger: zaptest.NewLogger(t),
			want:   &VMServiceStarter{logger: zaptest.NewLogger(t)},
		},
		{
			name:   "with nil logger",
			logger: nil,
			want:   &VMServiceStarter{}, // logger will be created inside
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Remove this if it exists
			got := NewVMServiceStarter(tt.logger)
			assert.NotNil(t, got)
			assert.NotNil(t, got.logger)
		})
	}
}

func TestVMServiceStarter_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		service     func() Service
		deps        []interface{}
		wantErr     bool
		expectedErr string
	}{
		{
			name: "HTTP service success",
			service: func() Service {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("Initialize", mock.Anything, mock.Anything).Return(nil)
				mockHTTP.On("Type").Return(HTTPServiceType)
				mockHTTP.On("ConfigureRoutes", mock.Anything, mock.Anything).Return(nil)
				return mockHTTP
			},
			deps: []interface{}{
				func() Engine {
					mockEngine := new(MockEngine)
					mockEngine.On("Run", mock.Anything).Return(nil)
					return mockEngine
				}(),
			},
			wantErr: false,
		},
		{
			name: "HTTP service without engine",
			service: func() Service {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("Initialize", mock.Anything, mock.Anything).Return(nil)
				mockHTTP.On("Type").Return(HTTPServiceType)
				return mockHTTP
			},
			deps:        []interface{}{},
			wantErr:     true,
			expectedErr: "engine not found in dependencies for HTTP service",
		},
		{
			name: "HTTP service with initialize error",
			service: func() Service {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("Initialize", mock.Anything, mock.Anything).Return(errors.New("initialize error"))
				return mockHTTP
			},
			deps:        []interface{}{},
			wantErr:     true,
			expectedErr: "failed to initialize service: initialize error",
		},
		{
			name: "HTTP service with configure routes error",
			service: func() Service {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("Initialize", mock.Anything, mock.Anything).Return(nil)
				mockHTTP.On("Type").Return(HTTPServiceType)
				mockHTTP.On("ConfigureRoutes", mock.Anything, mock.Anything).Return(errors.New("configure error"))
				return mockHTTP
			},
			deps: []interface{}{
				func() Engine {
					mockEngine := new(MockEngine)
					return mockEngine
				}(),
			},
			wantErr:     true,
			expectedErr: "failed to configure routes: configure error",
		},
		{
			name: "unsupported service type",
			service: func() Service {
				mockService := new(MockService)
				mockService.On("Initialize", mock.Anything, mock.Anything).Return(nil)
				mockService.On("Type").Return(ServiceType("unknown"))
				return mockService
			},
			deps:        []interface{}{},
			wantErr:     true,
			expectedErr: "unsupported service type for VM platform: unknown",
		},
		{
			name: "HTTP service type with wrong interface",
			service: func() Service {
				// Service says it's HTTP but doesn't implement HTTPService
				mockService := new(MockService)
				mockService.On("Initialize", mock.Anything, mock.Anything).Return(nil)
				mockService.On("Type").Return(HTTPServiceType)
				return mockService
			},
			deps:        []interface{}{},
			wantErr:     true,
			expectedErr: "service claims to be HTTP but does not implement HTTPService interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Parallel()
			starter := NewVMServiceStarter(logger)
			service := tt.service()

			// Use a context with timeout to avoid test hanging
			ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()

			// For tests with engine.Run(), which normally blocks, we need to cancel the context
			// to allow the test to complete
			if mockHTTP, ok := service.(*MockHTTPService); ok {
				if mockHTTP.AssertExpectations(t) {
					for _, dep := range tt.deps {
						if mockEngine, ok := dep.(Engine); ok {
							// Mock the Run method to return after a short delay or when context is done
							mockEngine.(*MockEngine).On("Run", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
								// Wait for context to be done or a short timeout
								select {
								case <-ctx.Done():
									return
								case <-time.After(50 * time.Millisecond):
									cancel() // Cancel context to allow test to complete
									return
								}
							})
						}
					}
				}
			}

			err := starter.Start(ctx, service, tt.deps...)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify that all expectations were met
			if mockHTTP, ok := service.(*MockHTTPService); ok {
				mockHTTP.AssertExpectations(t)
				for _, dep := range tt.deps {
					if mockEngine, ok := dep.(*MockEngine); ok {
						mockEngine.AssertExpectations(t)
					}
				}
			}
		})
	}
}

func TestVMServiceStarter_startHTTPService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		service     func() HTTPService
		deps        []interface{}
		wantErr     bool
		expectedErr string
	}{
		{
			name: "success with engine",
			service: func() HTTPService {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("ConfigureRoutes", mock.Anything, mock.Anything).Return(nil)
				return mockHTTP
			},
			deps: []interface{}{
				func() Engine {
					mockEngine := new(MockEngine)
					mockEngine.On("Run", mock.Anything).Return(nil)
					return mockEngine
				}(),
			},
			wantErr: false,
		},
		{
			name: "no engine provided",
			service: func() HTTPService {
				mockHTTP := new(MockHTTPService)
				return mockHTTP
			},
			deps:        []interface{}{},
			wantErr:     true,
			expectedErr: "engine not found in dependencies for HTTP service",
		},
		{
			name: "configure routes error",
			service: func() HTTPService {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("ConfigureRoutes", mock.Anything, mock.Anything).Return(errors.New("configure error"))
				return mockHTTP
			},
			deps: []interface{}{
				func() Engine {
					mockEngine := new(MockEngine)
					return mockEngine
				}(),
			},
			wantErr:     true,
			expectedErr: "failed to configure routes: configure error",
		},
		{
			name: "engine run error",
			service: func() HTTPService {
				mockHTTP := new(MockHTTPService)
				mockHTTP.On("ConfigureRoutes", mock.Anything, mock.Anything).Return(nil)
				return mockHTTP
			},
			deps: []interface{}{
				func() Engine {
					mockEngine := new(MockEngine)
					mockEngine.On("Run", mock.Anything).Return(errors.New("run error"))
					return mockEngine
				}(),
			},
			wantErr:     true,
			expectedErr: "run error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			starter := NewVMServiceStarter(logger)
			service := tt.service()

			// Use a context with timeout to avoid test hanging
			ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()

			// For tests with engine.Run(), which normally blocks, we need to cancel the context
			// to allow the test to complete
			for _, dep := range tt.deps {
				if mockEngine, ok := dep.(*MockEngine); ok {
					if !tt.wantErr {
						// For success cases, Run will be called and should wait for context or timeout
						mockEngine.On("Run", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
							select {
							case <-ctx.Done():
								return
							case <-time.After(50 * time.Millisecond):
								cancel() // Cancel context to allow test to complete
								return
							}
						})
					}
				}
			}

			err := starter.startHTTPService(ctx, service, tt.deps...)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify that all expectations were met
			service.(*MockHTTPService).AssertExpectations(t)
			for _, dep := range tt.deps {
				if mockEngine, ok := dep.(*MockEngine); ok {
					mockEngine.AssertExpectations(t)
				}
			}
		})
	}
}

func TestVMServiceStarter_setupSignalHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	starter := NewVMServiceStarter(logger)

	// Create a context that will be canceled explicitly before the test ends
	ctx, cancel := context.WithCancel(context.Background())

	// This shouldn't panic
	starter.setupSignalHandling(ctx)

	// Create a timeout to ensure the test doesn't hang
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	// Wait for a short time to ensure signal handling is set up
	<-timer.C

	// Cancel the context explicitly before the test ends
	// This will signal the goroutine to exit cleanly
	cancel()

	// Give the goroutine time to process the cancellation and exit
	time.Sleep(50 * time.Millisecond)
}
