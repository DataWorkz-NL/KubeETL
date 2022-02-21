package manager

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ControllerManager manages all of KubeETLs controllers + webhooks
type ControllerManager struct {
	bindPort int

	leaderElectionEnabled bool
	leaderElectionId string

	metricsAddr string
	
	scheme *runtime.Scheme
	schemasRegistration []SchemeRegistration

	reconcilerRegstration []ReconcilerRegistration

	webhooksEnabled bool
	webhookRegistration []WebhookRegistration
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
	// TODO add proper logging
	for _, addToScheme := range cm.schemasRegistration {
		if err := addToScheme(cm.scheme); err != nil {
			return fmt.Errorf("could not initialize schemas: %w", err)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             cm.scheme,
		MetricsBindAddress: cm.metricsAddr,
		Port:               cm.bindPort,
		LeaderElection:     cm.leaderElectionEnabled,
		LeaderElectionID:   cm.leaderElectionId,
	})
	if err != nil {
		return fmt.Errorf("could not create runtime manager: %w", err)
	}

	for _, registerReconciler := range cm.reconcilerRegstration {
		if err := registerReconciler(mgr); err != nil {
			return fmt.Errorf("could not register reconciler: %v", err)
		}
	}

	if cm.webhooksEnabled {
		for _, registerWebhook := range cm.webhookRegistration {
			if err := registerWebhook(mgr); err != nil {
				return fmt.Errorf("unable to register webhook: %v", err)
			}
			
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

type ReconcilerRegistration func(ctrl.Manager) error

func WithReconcilers(reconcilers ...ReconcilerRegistration) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.reconcilerRegstration = reconcilers
	}
}

type WebhookRegistration func(ctrl.Manager) error

func WithWebhooks(webhooks ...WebhookRegistration) ControllerManagerOpts {
	return func(cm *ControllerManager) {
		cm.webhooksEnabled = true
		cm.webhookRegistration = webhooks
	}
}