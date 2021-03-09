package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/BlockscapeNetwork/signctrl/privval"
	"github.com/spf13/cobra"
	tm_json "github.com/tendermint/tendermint/libs/json"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Shows the node's status",
		Long:  "Prints out the current height, rank and missed block counter",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := http.DefaultClient.Get("http://127.0.0.1:8080/status")
			if err != nil {
				fmt.Printf("couldn't get status: %v", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var sr privval.StatusResponse
			if err := tm_json.Unmarshal(bytes, &sr); err != nil {
				fmt.Printf("couldn't unmarshal status response: %v", err)
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
