package commands

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubetl <subcommand>",
		Short:        "KubeETL CLI",
		SilenceUsage: true,
	}

	cmd.AddCommand(NewManagerCommand())
	cmd.AddCommand(NewInjectionCommand())
	return cmd
}
