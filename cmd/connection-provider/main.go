package main

import (
	// load authentication plugin for obtaining credentials from cloud providers.
	"github.com/dataworkz/kubeetl/cmd/connection-provider/commands"
	"github.com/sirupsen/logrus"
)

func main() {
	err := commands.NewRootCommand().Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
