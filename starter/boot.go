package starter

import (
	"context"
	"fmt"
	"github.com/jjmaturino/bootstrapper/platform"
	"go.uber.org/zap"
	"sync"
)

// ServiceLauncher manages the registration and launching of services on different platforms
type ServiceLauncher struct {
	// serviceStarterRegistry maps platform types to service starter implementations
	serviceStarterRegistry map[platform.Type]platform.ServiceStarter

	// registryMu protects the registry
	registryMu sync.RWMutex

	// logger for the launcher
	logger *zap.Logger
}

// NewServiceLauncher creates a new service launcher with the provided logger
func NewServiceLauncher(ctx context.Context, logger *zap.Logger) *ServiceLauncher {
	launcher := &ServiceLauncher{
		serviceStarterRegistry: make(map[platform.Type]platform.ServiceStarter),
		logger:                 logger,
	}

	// Register builtin platform starters
	launcher.RegisterPlatform(ctx, platform.VM, platform.NewVMServiceStarter(logger))
	// Other platforms would be registered here

	return launcher
}

// Start launches a service on the specified platform
func (l *ServiceLauncher) Start(
	ctx context.Context,
	service platform.Service,
	platformType platform.Type,
	deps ...interface{},
) error {
	// Get the appropriate service starter for the platform
	l.registryMu.RLock()
	starter, ok := l.serviceStarterRegistry[platformType]
	l.registryMu.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported platform type: %s", platformType)
	}

	// Start the service with the platform-specific starter
	l.logger.Info("Starting service",
		zap.String("platform", string(platformType)),
		zap.String("serviceType", string(service.Type())))

	return starter.Start(ctx, service, deps...)
}

// GetPlatformStarter retrieves a registered platform service starter
func (l *ServiceLauncher) GetPlatformStarter(platformType platform.Type) (platform.ServiceStarter, error) {
	l.registryMu.RLock()
	defer l.registryMu.RUnlock()

	starter, ok := l.serviceStarterRegistry[platformType]
	if !ok {
		return nil, fmt.Errorf("no starter registered for platform: %s", platformType)
	}

	return starter, nil
}

// RegisterPlatform allows registering custom platform service starters
func (l *ServiceLauncher) RegisterPlatform(ctx context.Context, platformType platform.Type, starter platform.ServiceStarter) {
	l.registryMu.Lock()
	defer l.registryMu.Unlock()

	if _, exists := l.serviceStarterRegistry[platformType]; exists {
		l.logger.Warn("Overriding existing platform starter",
			zap.String("platform", string(platformType)))
	}

	l.serviceStarterRegistry[platformType] = starter
	l.logger.Info("Registered platform starter",
		zap.String("platform", string(platformType)))
}
