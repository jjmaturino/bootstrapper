package platform

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultGinEngine(t *testing.T) {
	// Initialize a test logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create the gin engine using the DefaultGinEngine function
	engine, err := DefaultGinEngine(logger)
	require.NoError(t, err)
	require.NotNil(t, engine)

	// Set gin mode to test mode
	gin.SetMode(gin.TestMode)

	// Create a test server using the gin engine
	ts := httptest.NewServer(engine)
	defer ts.Close()

	// Test the NoRoute handler by making a request to a nonexistent route
	resp, err := http.Get(ts.URL + "/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert that the status code is 404 Not Found
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Assert that the CORS header is set correctly
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestDefaultGinEngineWithoutLogger(t *testing.T) {
	// Create the gin engine using the DefaultGinEngine function
	engine, err := DefaultGinEngine(nil)
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestStartVM_ConstructServicesError(t *testing.T) {
	serviceError := errors.New("construct services error")

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

func TestStartVM_Success(t *testing.T) {
	// Create a mock engine that returns nil when running
	// Create a mock gin service that returns nil when constructing services
	service := &MockGinService{}
	eng := MockEngineSuccess{}

	err := StartVM(service, &eng)

	if err != nil {
		t.Errorf("Expected nil error but got '%s'", err.Error())
	}
	require.NoError(t, err)
}
