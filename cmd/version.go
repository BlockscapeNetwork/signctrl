package cmd

import (
	"fmt"

	"github.com/BlockscapeNetwork/signctrl/cmd/version"
	"github.com/spf13/cobra"
)

const (
	// SemVer is the semantiv version of SignCTRL.
	SemVer = "0.1.0-RC1"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints out the version of SignCTRL",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`SignCTRL
  Version:    %v
  Git commit: %v`, SemVer, version.CommitHash())
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
