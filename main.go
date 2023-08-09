package main

import (
	"log"

	"github.com/mknycha/async-data-processor/cmd"
	"github.com/mknycha/async-data-processor/cmd/messagesconsumer"
	"github.com/mknycha/async-data-processor/cmd/server"
)

func main() {
	rootCmd := cmd.RootCommand()
	rootCmd.AddCommand(server.ServerCommand())
	rootCmd.AddCommand(messagesconsumer.MessageConsumerCommand())
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
