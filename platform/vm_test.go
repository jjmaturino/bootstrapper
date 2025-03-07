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

	service := &MockGinService{
		ConstructServicesFunc: func(ctx context.Context, deps ...interface{}) error {
			return serviceError
		},
	}
	eng := MockEngineSuccess{}

	err := StartVM(service, &eng)
	require.Error(t, err)
	assert.Equal(t, serviceError, err)
}

func TestStartVM_SetupEngineError(t *testing.T) {
	serviceError := errors.New("setup engine error")

	service := &MockGinService{
		SetupEngineFunc: func(eng Engine) (Engine, error) {
			return nil, serviceError
		},
	}
	eng := MockEngineSuccess{}
	r := eng.Handle("GET", "/", func(c *gin.Context) {})

	err := StartVM(service, &eng)

	require.Nil(t, r)
	require.Error(t, err)
	assert.Equal(t, serviceError, err)
}

func TestStartVM_EngineNil(t *testing.T) {
	err := StartVM(&MockGinService{}, nil)
	require.Error(t, err)
	assert.Equal(t, "engine is nil", err.Error())
}

func TestStartVM_RunError(t *testing.T) {
	// Create a mock engine that returns an error when running
	// Create a mock gin service that returns nil when constructing services
	service := &MockGinService{}
	eng := MockEngineError{}
	r := eng.Handle("GET", "/", func(c *gin.Context) {})

	err := StartVM(service, &eng)

	require.Nil(t, r)
	require.Error(t, err)
	assert.Equal(t, "run error", err.Error())
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

	// Testing signal handling is tricky, so we'll just ensure it doesn't panic
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This shouldn't panic
	starter.setupSignalHandling(ctx)

	// Wait for context to be done
	<-ctx.Done()
}
