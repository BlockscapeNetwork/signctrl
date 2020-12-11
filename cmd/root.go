package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pairmint",
	Short: "Pairmint is a high availability solution for Tendermint validators",
	Long:  "A high availability solution for Tendermint validators written in Go, built by blockscape",
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
