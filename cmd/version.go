package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// GitCommit is the current commit hash.
	GitCommit = ""

	// SemVer is the semantiv version of SignCTRL.
	SemVer = ""

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints out the version of SignCTRL",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`SignCTRL
  Version:    %v
  Git commit: %v
`, SemVer, GitCommit)
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
