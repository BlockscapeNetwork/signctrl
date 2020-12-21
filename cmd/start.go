package cmd

import (
	"fmt"
	"os"

	"github.com/gogo/protobuf/proto"

	"github.com/BlockscapeNetwork/pairmint/privval"
	"github.com/BlockscapeNetwork/pairmint/utils"

	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Flags
	tmkms bool

	// The start command starts the pairmint application.
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the pairmint application",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			// If no custom configuration directory is set, always use the default one.
			if os.Getenv("PAIRMINT_CONFIG_DIR") == "" {
				os.Setenv("PAIRMINT_CONFIG_DIR", os.Getenv("HOME")+"/.pairmint")
			}

			// Create new PairmintFilePV instance.
			pm, err := privval.NewPairmintFilePV()
			if err != nil {
				pm.Logger.Printf("[ERR] pairmint: couldn't initialize pairminter: %v", err)
				os.Exit(1)
			}

			// Dial the Tendermint validator.
			conn := utils.RetryDial("tcp", pm.Config.Init.ValidatorAddr, pm.Logger)
			defer conn.Close()

			// Keep the application running.
			for {
				data := make([]byte, 16<<10)
				dataLen, err := conn.Read(data)
				if err != nil {
					pm.Logger.Printf("[ERR] pairmint: error while reading data: %v\n", err)
					continue
				}

				msg := privvalproto.Message{}
				if err = proto.Unmarshal(data[:dataLen], &msg); err != nil {
					pm.Logger.Printf("[ERR] pairmint: error while unmarshaling: %v\n", err)
					os.Exit(1) // TODO: continue

					// This currently returns an error because pairmint can't decrypt the message.
				}

				switch msg.GetSum().(type) {
				case *privvalproto.Message_PingRequest:
					req := msg.GetPingRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received PingRequest: %v\n", req)

				case *privvalproto.Message_PubKeyRequest:
					req := msg.GetPubKeyRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received PubKeyRequest: %v\n", req)

				case *privvalproto.Message_SignVoteRequest:
					req := msg.GetSignVoteRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received SignVoteRequest: %v\n", req)

				case *privvalproto.Message_SignProposalRequest:
					req := msg.GetSignProposalRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received SignProposalRequest: %v\n", req)

				default:
					panic(fmt.Sprintf("unknown sum type: %T", msg.GetSum()))
				}

			}

		},
	}
)

func init() {
	// Add start to the root command.
	rootCmd.AddCommand(startCmd)

	// Add tmkms flag.
	startCmd.Flags().BoolVar(&tmkms, "tmkms", false, "Use the TMKMS as an external PrivValidator process for Pairmint")
	viper.BindPFlag("tmkms", startCmd.Flags().Lookup("tmkms"))
}
