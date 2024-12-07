package platform

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type Engine interface {
	Run(addr ...string) (err error)
	Handle(method, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
}

type Service interface {
	ConstructService(ctx context.Context, injector *do.Injector) error
	SetupEngine(engine Engine) (Engine, error)
}

type ApiService interface {
	ConstructService(ctx context.Context, deps ...interface{}) error
	// SetupEngine is called after ConstructServices. The Service is expected to bind any routes or other
	// necessary work to setup provided Engine for initial use. Usually this just means binding any appropriate
	// routes to handlers.
	SetupEngine(eng Engine) (Engine, error)
}
