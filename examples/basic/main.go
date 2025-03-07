package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jjmaturino/bootstrapper/platform"
	"github.com/jjmaturino/bootstrapper/starter"
	"go.uber.org/zap"
)

// MyService is an example service implementation
type MyService struct {
	logger *zap.Logger
}

// NewService creates a new instance of MyService
func NewService() *MyService {
	logger, _ := zap.NewProduction()
	return &MyService{
		logger: logger,
	}
}

// ConstructService initializes the service
func (s *MyService) ConstructService(ctx context.Context, deps ...interface{}) error {
	s.logger.Info("Constructing service")

	// Process any dependencies
	for _, dep := range deps {
		switch d := dep.(type) {
		case *zap.Logger:
			s.logger = d
		default:
			// Ignore unknown dependencies
		}
	}

	s.logger.Info("Service constructed successfully")
	return nil
}

// SetupEngine configures the HTTP engine
func (s *MyService) SetupEngine(eng platform.Engine) (platform.Engine, error) {
	s.logger.Info("Setting up engine")

	// Add custom routes
	eng.Handle("GET", "/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	eng.Handle("GET", "/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"data": []string{"item1", "item2", "item3"},
		})
	})

	s.logger.Info("Engine setup complete")
	return eng, nil
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create Gin engine with default configuration
	engine, err := platform.DefaultGinEngine(logger)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// Create service
	service := NewService()

	// Start the service on VM platform
	starter.Start(service, platform.VM, engine, logger)
}
