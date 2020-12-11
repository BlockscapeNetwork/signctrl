package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// The version command prints out version information.
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Pairmint",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pairmint v0.0.0-alpha")
		},
	}
)

func init() {
	// Add version to the root command.
	rootCmd.AddCommand(versionCmd)
}
