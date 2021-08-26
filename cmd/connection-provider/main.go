package main

import (
	"os"
	"os/exec"

	// load authentication plugin for obtaining credentials from cloud providers.
	"github.com/dataworkz/kubeetl/cmd/connection-provider/commands"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	err := commands.NewRootCommand().Execute()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() >= 0 {
				os.Exit(exitError.ExitCode())
			} else {
				os.Exit(137) // probably SIGTERM or SIGKILL
			}
		} else {
			// util.WriteTeriminateMessage(err.Error()) // we don't want to overwrite any other message
			println(err.Error())
			os.Exit(64)
		}
	}
}
