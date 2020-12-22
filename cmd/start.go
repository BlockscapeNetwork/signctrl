package cmd

import (
	"fmt"
	"os"

	"github.com/gogo/protobuf/proto"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/privval"
	"github.com/BlockscapeNetwork/pairmint/utils"

	"github.com/tendermint/tendermint/crypto/ed25519"
	p2pconn "github.com/tendermint/tendermint/p2p/conn"
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
			// Get the configuration directory.
			configDir := config.GetConfigDir()

			// Create new PairmintFilePV instance.
			pm, err := privval.NewPairmintFilePV()
			if err != nil {
				pm.Logger.Printf("[ERR] pairmint: couldn't initialize pairminter: %v", err)
				os.Exit(1)
			}
			pm.Logger.Println("[DEBUG] pairmint: Created new PairmintFilePV successfully. ✓")

			// Load the keypair from the pm-identity.key file. The keypair is necessary for
			// establishing a secret connection to Tendermint.
			priv, _, err := utils.LoadKeypair(configDir + "/pm-identity.key")
			if err != nil {
				pm.Logger.Printf("[ERR] pairmint: error while loading keypair: %v\n", err)
				os.Exit(1)
			}
			pm.Logger.Println("[DEBUG] pairmint: Loaded keypair successfully. ✓")

			// Establish a connection to the Tendermint validator.
			conn := utils.RetryDial("tcp", pm.Config.Init.ValidatorAddr, pm.Logger)
			defer conn.Close()
			pm.Logger.Println("[DEBUG] pairmint: Dialed Tendermint validator successfully. ✓")

			// Make the connection to the Tendermint validator secret.
			secretConn, err := p2pconn.MakeSecretConnection(conn, ed25519.PrivKey(priv))
			if err != nil {
				pm.Logger.Printf("[ERR] pairmint: error while establishing secret connection: %v\n", err)
				os.Exit(1)
			}
			defer secretConn.Close()
			pm.Logger.Println("[DEBUG] pairmint: Established secret connection with Tendermint validator successfully. ✓")

			// Keep the application running.
			for {
				data := make([]byte, 16<<10)
				dataLen, err := secretConn.Read(data)
				if err != nil {
					pm.Logger.Printf("[ERR] pairmint: error while reading data: %v\n", err)
					continue
				}

				msg := privvalproto.Message{}
				if err = proto.Unmarshal(data[:dataLen], &msg); err != nil {
					pm.Logger.Printf("[ERR] pairmint: error while unmarshaling: %v\n", err)
					os.Exit(1) // TODO: continue

					// TODO: Find out why it throws this error.
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
