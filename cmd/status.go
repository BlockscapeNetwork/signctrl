package cmd

import (
	"fmt"
	"os"

	"github.com/BlockscapeNetwork/signctrl/privval"
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Shows the node's status",
		Long:  "Prints out the current height, rank and missed block counter",
		Run: func(cmd *cobra.Command, args []string) {
			sr, err := privval.GetStatus()
			if err != nil {
				fmt.Printf("couldn't get status: %v", err)
				os.Exit(1)
			}

			fmt.Printf(`Status of SignCTRL validator:
  Height:  %v
  Rank:    %v/%v
  Counter: %v/%v
`, sr.Height, sr.Rank, sr.SetSize, sr.Counter, sr.Threshold)
		},
	}
)

func init() {
	rootCmd.AddCommand(statusCmd)
}
