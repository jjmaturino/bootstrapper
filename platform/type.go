package platform

type Type string

// Platform environment constants
const (
	VM Type = "virtual_machine" // Only implementing VM for now

	// Future platform types (placeholders)
	// Docker      Type = "docker"
	// Lambda      Type = "lambda"
	// Kubernetes  Type = "kubernetes"
)

func (pT *Type) String() string {
	return string(*pT)
}

// ServiceType defines the type of service being run
type ServiceType string

// Service type constants
const (
	HTTPServiceType ServiceType = "http" // Only implementing HTTP service for now

	// Future service types (placeholders)
	// GRPCService	  ServiceType = "grcp"
	// QueueService   ServiceType = "queue"
	// WorkerService  ServiceType = "worker"
	// ScheduledTask  ServiceType = "scheduled"
)

func (sT *ServiceType) String() string {
	return string(*sT)
}
