package platform

import (
	"context"
	"errors"
	"github.com/jjmaturino/bootstrapper/api"
	"log"
	"time"

	gzip "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func DefaultGinEngine(logger *zap.Logger) (*gin.Engine, error) {
	r := gin.Default()
	r.ContextWithFallback = true

	if logger == nil {
		logger = zap.L()
	}

	r.Use(
		gzip.Ginzap(logger, time.RFC3339, true),
		gzip.RecoveryWithZap(logger, true),
		corsMiddleware("*"),
	)
	r.NoRoute(func(c *gin.Context) {
		err := api.SendNotFoundResponse(c)
		if err != nil {
			zap.L().Error("failed to SendNotFoundResponse", zap.Error(err))
		}
	})
	return r, nil
}

// StartVM begins a gin server on a virtual machine
func StartVM(service ApiService, engine Engine, deps ...interface{}) error {
	ctx := context.TODO()

	if engine == nil {
		return errors.New("engine is nil")
	}

	err := service.ConstructService(ctx, deps)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return err
	}

	eng, err := service.SetupEngine(engine)
	if err != nil {
		log.Printf("Error: %s", err.Error())

		return err
	}

	// Start the Gin server on default port 8080
	return eng.Run() // Default listens on :8080
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
