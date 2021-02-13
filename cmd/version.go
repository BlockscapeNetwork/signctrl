package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// The version command prints out version information.
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of SignCTRL",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pairmint v0.1.0-RC1")
		},
	}
)

func init() {
	// Add version to the root command.
	rootCmd.AddCommand(versionCmd)
}
