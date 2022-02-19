package manager

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// ControllerManager manages all of KubeETLs controllers + webhooks
type ControllerManager struct {
	bindPort int

	leaderElectionEnabled bool
	leaderElectionId string

	metricsAddr string
	
	scheme *runtime.Scheme
	schemasRegistration []SchemeRegistration
}

func New(opts ...ControllerManagerOpts) *ControllerManager {
	// TODO set defaults
	m := &ControllerManager{}
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Init initializes all components of the KubeETL operator:
// - It adds the KubeETL schemas to the runtime.Scheme
func (cm *ControllerManager) Init() error {
	for _, addToScheme := range cm.schemasRegistration {
		err := addToScheme(cm.scheme)
		if err != nil {
			return fmt.Errorf("could not initialize schemas: %w", err)
		}
	}

	return nil
}

type ControllerManagerOpts func(*ControllerManager)

func WithBindPort(port int) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.bindPort = port
	}
}

func WithLeaderElection(enabled bool) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.leaderElectionEnabled = enabled
		if enabled {
			cm.leaderElectionId = "1345a080.dataworkz.nl"
		}
	}
}

func WithMetricsAddress(address string) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.metricsAddr = address
	}
}

func WithScheme(scheme *runtime.Scheme) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.scheme = scheme
	}
}

func WithDefaultScheme() ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.scheme = runtime.NewScheme()
	}
}

type SchemeRegistration func(*runtime.Scheme) error

func WithSchemas(schemas ...SchemeRegistration) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.schemasRegistration = schemas
	}
}