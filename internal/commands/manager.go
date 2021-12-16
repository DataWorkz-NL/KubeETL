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

// TODO move init to seperate manager package
func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = etlv1alpha1.AddToScheme(scheme)
	_ = etldataworkznlv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type managerConfig struct {
	metricsAddr          string
	enableLeaderElection bool
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

	return cmd
}

func (c *managerConfig) run() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: c.metricsAddr,
		Port:               9443,
		LeaderElection:     c.enableLeaderElection,
		LeaderElectionID:   "1345a080.dataworkz.nl",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.DataSetReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataSet")
		os.Exit(1)
	}

	err = etlhooks.SetupValidatingConnectionWebhookWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to start webhook")
		os.Exit(1)
	}
	if err = (&controllers.WorkflowReconciler{
		Client:                   mgr.GetClient(),
		Log:                      ctrl.Log.WithName("controllers").WithName("Workflow"),
		Scheme:                   mgr.GetScheme(),
		ConnectionInjectionImage: DockerImage,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Workflow")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
