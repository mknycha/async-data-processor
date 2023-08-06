package main

import (
	"log"

	_ "github.com/lib/pq"
	"github.com/mknycha/async-data-processor/cmd"
	"github.com/mknycha/async-data-processor/cmd/server"
	"github.com/mknycha/async-data-processor/cmd/workerpool"
)

func main() {
	rootCmd := cmd.RootCommand()
	rootCmd.AddCommand(server.ServerCommand())
	rootCmd.AddCommand(workerpool.WorkerPoolCommand())
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
