package commands

import (
	// load authentication plugin for obtaining credentials from cloud providers.
	"fmt"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dataworkz/kubeetl/api/v1alpha1"
	"github.com/dataworkz/kubeetl/internal/provider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(initConfig)
}

const (
	// CLIName is the name of the CLI
	WorkflowResourceName = "Workflows"
	CLIName              = "connectionprovider"
)

var (
	workflow  string
	namespace string
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
		FullTimestamp:   true,
	})
}

func NewInjectionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   fmt.Sprintf("%s --workflow <workflow-name> --namespace <workflow-namespace>", CLIName),
		Short: fmt.Sprintf("%s provides injectable secrets for a workflow", CLIName),
		Run: func(cmd *cobra.Command, args []string) {
			// creates the in-cluster config
			config, err := rest.InClusterConfig()
			er(err)

			scheme, err := v1alpha1.SchemeBuilder.Build()
			er(err)

			client, err := client.New(config, client.Options{Scheme: scheme})
			er(err)

			p := provider.NewSecretProvider(client)

			err = p.ProvideWorkflowSecret(workflow, namespace)
			er(err)
		},
	}

	return command
}

func er(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
