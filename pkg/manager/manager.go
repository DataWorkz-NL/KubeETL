package manager

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ControllerManager manages all of KubeETLs controllers + webhooks
type ControllerManager struct {
	bindPort int

	mgr ctrl.Manager

	leaderElectionEnabled bool
	leaderElectionId      string

	metricsAddr string

	scheme              *runtime.Scheme
	schemasRegistration []SchemeRegistration

	reconcilerRegstration []ReconcilerRegistration

	webhooksEnabled     bool
	webhookRegistration []WebhookRegistration

	initOnce sync.Once
	opts     []ControllerManagerOpts
}

func New(opts ...ControllerManagerOpts) *ControllerManager {
	return &ControllerManager{
		opts: opts,
	}
}

// setDefaults on the ControllerManager
func (cm *ControllerManager) setDefaults() {
	cm.bindPort = 9443
	cm.scheme = runtime.NewScheme()
}

// Init initializes all components of the KubeETL operator:
// - It configures defaults for the CM and applies all options
// - It adds the KubeETL schemas to the runtime.Scheme
// - It registers all reconcilers & webhooks
func (cm *ControllerManager) init() error {
	cm.setDefaults()
	for _, opt := range cm.opts {
		opt(cm)
	}

	// TODO add proper logging
	for _, addToScheme := range cm.schemasRegistration {
		if err := addToScheme(cm.scheme); err != nil {
			return fmt.Errorf("could not initialize schemas: %w", err)
		}
	}

	// TODO move out and make sure this is mockable
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
	cm.mgr = mgr

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

// Start starts the controller manager and the underlying reconcilers + webhooks
func (cm *ControllerManager) Start() error {
	cm.initOnce.Do(func() {
		err := cm.init()
		if err != nil {
			panic(err)
		}
	})
	if err := cm.mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("manager run failed: %v", err)
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
