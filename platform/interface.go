package platform

import (
	"context"
	"github.com/gin-gonic/gin"
)

// Engine is an interface for HTTP engines like Gin
type Engine interface {
	Run(addr ...string) (err error)
	Handle(method, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes // TODO: Generalize this, Only allows for gin
}

// ApiService defines the interface that all services must adhere to
type ApiService interface {
	ConstructService(ctx context.Context, deps ...interface{}) error // TODO: Potentiall add ability to pass something like deps injector from do'samber
	// SetupEngine is called after ConstructServices. The Service is expected to bind any routes or other
	// necessary work to setup provided Engine for initial use. Usually this just means binding any appropriate
	// routes to handlers.
	SetupEngine(eng Engine) (Engine, error)
}

// RuntimeStarter defines how to start a service on a specific runtime
type RuntimeStarter interface {
	// Start begins the service on the specific runtime
	Start(service ApiService, engine Engine, deps ...interface{}) error
}
