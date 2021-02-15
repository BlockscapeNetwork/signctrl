package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "signctrl",
	Short: "High availability solution for Tendermint validators",
	Long:  "SignCTRL is a high availability solution for Tendermint validators written in Go, built by blockscape",
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(int(syscall.SIGHUP))
	}
}
