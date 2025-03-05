package platform

import (
	"context"
	"github.com/gin-gonic/gin"
)

// Engine is an interface for HTTP engines like Gin
type Engine interface {
	Run(addr ...string) (err error)
	Handle(method, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes // TODO: Generalize this, Only currently allows for gin
}

// HTTPService defines the interface that all services must adhere to
type HTTPService interface {
	Service

	// ConfigureRoutes sets ups the routes of the http service using an engine
	ConfigureRoutes(ctx context.Context, engine Engine) error
}

// Service is the base interface for all service types
type Service interface {
	// Initialize sets up the service with dependencies
	Initialize(ctx context.Context, deps ...interface{}) error // TODO: Potentially add ability to pass something like deps injector from do'samber

	// Type returns the service type
	Type() ServiceType
}

// ServiceStarter defines how to start a service on a specific platform runtime
type ServiceStarter interface {
	// Start begins the service on the specific runtime
	Start(ctx context.Context, service Service, deps ...interface{}) error
}
