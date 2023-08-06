package cmd

import "github.com/spf13/cobra"

func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "async-data-processor",
		Short: "An application for processing asynchronously data",
	}
}
