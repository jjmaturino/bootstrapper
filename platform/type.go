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
