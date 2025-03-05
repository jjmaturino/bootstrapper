package platform

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// startHTTPService starts an HTTP service on the VM runtime platform
func (v *VMServiceStarter) startHTTPService(ctx context.Context, service HTTPService, deps ...interface{}) error {
	v.logger.Info("Setting up HTTP service")

	// Find the engine in the dependencies
	var engine Engine
	for _, dep := range deps {
		if eng, ok := dep.(Engine); ok {
			engine = eng
			break
		}
	}

	if engine == nil {
		return errors.New("engine not found in dependencies for HTTP service")
	}

	// Configure routes
	v.logger.Info("Configuring HTTP routes")
	if err := service.ConfigureRoutes(ctx, engine); err != nil {
		v.logger.Error("Failed to configure routes", zap.Error(err))
		return fmt.Errorf("failed to configure routes: %w", err)
	}

	// Setup signal handling for graceful shutdown
	v.setupSignalHandling(ctx)

	// Start the HTTP server
	v.logger.Info("Starting HTTP server")

	// Run the engine (this is blocking)
	// Start the Gin server on default port 8080
	return engine.Run() // Default listens on :8080
}

// setupSignalHandling sets up OS signal handlers  for graceful shutdown
func (v *VMServiceStarter) setupSignalHandling(ctx context.Context) {
	// Create a cancellable context that we can pass to child goroutines
	ctx, cancel := context.WithCancel(ctx)

	// Create channel to listen for signals
	sigChan := make(chan os.Signal, 1)

	// Register for SIGINT and SIGTERM
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle signals in a separate goroutine
	go func() {
		select {
		case sig := <-sigChan:
			v.logger.Info("Received signal", zap.String("signal", sig.String()))
			cancel() // Cancel context to notify all parts of the application
		case <-ctx.Done():
			// Context was cancelled elsewhere
			v.logger.Info("Context done, exiting signal handler")
		}
	}()
}

// StartService starts a service on the VM platform based on service type
func (v *VMServiceStarter) Start(ctx context.Context, service Service, deps ...interface{}) error {
	v.logger.Info("Starting service on VM platform", zap.String("type", string(service.Type())))

	// Initialize the service first
	if err := service.Initialize(ctx, deps...); err != nil {
		v.logger.Error("Failed to initialize service", zap.Error(err))
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	// Handle based on service type
	switch service.Type() {
	case HTTPServiceType:
		httpService, ok := service.(HTTPService)
		if !ok {
			return errors.New("service claims to be HTTP but does not implement HTTPService interface")
		}
		return v.startHTTPService(ctx, httpService, deps...)

	default:
		return fmt.Errorf("unsupported service type for VM platform: %s", service.Type())
	}
}

// NewVMServiceStarter creates a new VM service starter
func NewVMServiceStarter(logger *zap.Logger) *VMServiceStarter {
	if logger == nil {
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			log.Printf("Failed to create logger: %v", err)
		}
	}

	return &VMServiceStarter{
		logger: logger,
	}
}

// VMServiceStarter starts services on VM platform
type VMServiceStarter struct {
	logger *zap.Logger
}

var _ ServiceStarter = (*VMServiceStarter)(nil)
