package launcher

import (
	"context"
	"fmt"
	"github.com/jjmaturino/bootstrapper/platform"
	"go.uber.org/zap"
	"sync"
)

// ServiceRegistry manages the registration and retrieval of platform service starters
type ServiceRegistry struct {
	// serviceStarterRegistry maps platform types to service starter implementations
	serviceStarterRegistry map[platform.Type]platform.ServiceStarter

	// registryMu protects the registry
	registryMu sync.RWMutex

	// logger for the registry
	logger *zap.Logger
}

var (
	// serviceStarterRegistry maps platform types to service starter implementations
	serviceStarterRegistry = make(map[platform.Type]platform.ServiceStarter)

	// registryMu protects the registry
	registryMu sync.RWMutex

	// logger for the launcher
	logger *zap.Logger
)

// Start launches a service on the specified platform
func Start(
	ctx context.Context,
	service platform.Service,
	platformType platform.Type,
	deps ...interface{},
) error {
	// Get the appropriate service starter for the platform
	registryMu.RLock()
	starter, ok := serviceStarterRegistry[platformType]
	registryMu.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported platform type: %s", platformType)
	}

	// Start the service with the platform-specific starter
	logger.Info("Starting service",
		zap.String("platform", string(platformType)),
		zap.String("serviceType", string(service.Type())))

	return starter.StartService(ctx, service, deps...)
}

// GetPlatformStarter retrieves a registered platform service starter
func GetPlatformStarter(platformType platform.Type) (platform.ServiceStarter, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	starter, ok := serviceStarterRegistry[platformType]
	if !ok {
		return nil, fmt.Errorf("no starter registered for platform: %s", platformType)
	}

	return starter, nil
}

// RegisterPlatform allows registering custom platform service starters
func RegisterPlatform(platformType platform.Type, starter platform.ServiceStarter) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := serviceStarterRegistry[platformType]; exists {
		logger.Warn("Overriding existing platform starter",
			zap.String("platform", string(platformType)))
	}

	serviceStarterRegistry[platformType] = starter
	logger.Info("Registered platform starter",
		zap.String("platform", string(platformType)))
}

// init initializes the launcher package
func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	// Register builtin platform starters
	RegisterPlatform(platform.VM, platform.NewVMServiceStarter(logger))
	// Other platforms would be registered here
}
