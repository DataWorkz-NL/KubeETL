package commands

import (
	"fmt"
	"os"

	etldataworkznlv1alpha1 "github.com/dataworkz/kubeetl/api/v1alpha1"
	etlv1alpha1 "github.com/dataworkz/kubeetl/api/v1alpha1"
	etlhooks "github.com/dataworkz/kubeetl/api/v1alpha1/webhooks"
	"github.com/dataworkz/kubeetl/controllers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports

	"github.com/dataworkz/kubeetl/pkg/manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	ManagerCommand = "manager"
	// TODO make configurable
	DockerImage = "ghcr.io/dataworkz-nl/kubeetl:main"
)

type managerConfig struct {
	metricsAddr          string
	enableLeaderElection bool
	webhooksEnabled      bool
}

func NewManagerCommand() *cobra.Command {
	config := &managerConfig{}
	cmd := &cobra.Command{
		Use:   ManagerCommand,
		Short: fmt.Sprintf("%s is used to run the KubeETL operator", ManagerCommand),
		Run: func(cmd *cobra.Command, args []string) {
			config.run()
		},
	}

	cmd.Flags().StringVar(&config.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVar(&config.enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	cmd.Flags().BoolVar(&config.webhooksEnabled, "webhooks-enabled", false, "Enable validating webhooks for KubeETL.")

	return cmd
}

func (c *managerConfig) run() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	cm := manager.New(
		manager.WithMetricsAddress(c.metricsAddr),
		manager.WithLeaderElection(c.enableLeaderElection),
		manager.WithWebhooksEnabled(c.webhooksEnabled),
		manager.WithSchemas(
			clientgoscheme.AddToScheme,
			etlv1alpha1.AddToScheme,
			etldataworkznlv1alpha1.AddToScheme,
		),
		manager.WithWebhooks(
			etlhooks.SetupValidatingConnectionWebhookWithManager,
			etlhooks.SetupValidatingDataSetWebhookWithManager,
		),
		manager.WithReconcilers(
			(&controllers.DataSetReconciler{}).SetupWithManager,
			(&controllers.WorkflowReconciler{
				Log:                      ctrl.Log.WithName("controllers").WithName("Workflow"),
				ConnectionInjectionImage: DockerImage,
			}).SetupWithManager,
		),
	)
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := cm.Start(); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
