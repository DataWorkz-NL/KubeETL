package manager

import "k8s.io/apimachinery/pkg/runtime"

// ControllerManager manages all of KubeETLs controllers + webhooks
type ControllerManager struct {
	bindPort int

	leaderElectionEnabled bool
	leaderElectionId string

	metricsAddr string
	
	scheme *runtime.Scheme
}

func New(opts ...ControllerManagerOpts) *ControllerManager {
	m := &ControllerManager{}
	for _, opt := range opts {
		opt(m)
	}

	return m
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