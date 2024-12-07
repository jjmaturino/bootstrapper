package platform

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
)

type MockGinService struct {
	ConstructServicesFunc func(ctx context.Context, deps ...interface{}) error
	SetupEngineFunc       func(eng Engine) (Engine, error)
}

func (m *MockGinService) ConstructService(ctx context.Context, deps ...interface{}) error {
	if m.ConstructServicesFunc != nil {
		return m.ConstructServicesFunc(ctx, deps...)
	}
	return nil
}

func (m *MockGinService) SetupEngine(engine Engine) (Engine, error) {
	if m.SetupEngineFunc != nil {
		return m.SetupEngineFunc(engine)
	}

	return engine, nil
}

type MockEngineSuccess struct{}

func (m *MockEngineSuccess) Run(addr ...string) error {
	return nil
}

func (m *MockEngineSuccess) Handle(method, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return nil
}

type MockEngineError struct{}

func (m *MockEngineError) Run(addr ...string) error {
	return errors.New("run error")
}

func (m *MockEngineError) Handle(method, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
