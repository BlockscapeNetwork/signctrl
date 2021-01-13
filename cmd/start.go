package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/tendermint/tendermint/libs/protoio"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/privval"
	"github.com/BlockscapeNetwork/pairmint/utils"

	"github.com/tendermint/tendermint/crypto/ed25519"
	p2pconn "github.com/tendermint/tendermint/p2p/conn"
	cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
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

			protoReader := protoio.NewDelimitedReader(secretConn, 64<<10)
			protoWriter := protoio.NewDelimitedWriter(secretConn)

			// Keep the application running.
			for {
				msg := privvalproto.Message{}
				if _, err = protoReader.ReadMsg(&msg); err != nil {
					if err == io.EOF {
						// Prevent the console log from being spammed with EOF errors.
						continue
					}

					pm.Logger.Printf("[ERR] pairmint: error while reading message: %v\n", err)
					continue
				}

				switch msg.GetSum().(type) {
				case *privvalproto.Message_PingRequest:
					pm.Logger.Println("[DEBUG] pairmint: Received PingRequest")

					// Construct proto message for PingResponse.
					res := &privvalproto.Message{
						Sum: &privvalproto.Message_PingResponse{
							PingResponse: &privvalproto.PingResponse{},
						},
					}

					if _, err := protoWriter.WriteMsg(res); err != nil {
						pm.Logger.Printf("[ERR] pairmint: error while writing PingResponse: %v\n", err)
						continue
					}

				case *privvalproto.Message_PubKeyRequest:
					req := msg.GetPubKeyRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received PubKeyRequest: %v\n", req)

					// Get the pubkey from the priv_validator_key.json file.
					pubkey, err := pm.GetPubKey()
					if err != nil {
						pm.Logger.Printf("[ERR] pairmint: couldn't get pubkey: %v\n", err)
						os.Exit(1)
					}

					// Construct proto message for PubKeyResponse.
					res := &privvalproto.Message{
						Sum: &privvalproto.Message_PubKeyResponse{
							PubKeyResponse: &privvalproto.PubKeyResponse{
								PubKey: cryptoproto.PublicKey{
									Sum: &cryptoproto.PublicKey_Ed25519{
										Ed25519: pubkey.Bytes(),
									},
								},
							},
						},
					}

					if _, err = protoWriter.WriteMsg(res); err != nil {
						pm.Logger.Printf("[ERR] pairmint: error while writing PubKeyResponse: %v\n", err)
						continue
					}

				case *privvalproto.Message_SignVoteRequest:
					req := msg.GetSignVoteRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received SignVoteRequest: %v\n", req)

					// TODO: Sign vote if node is primary, and reply with signed vote.
					// TODO: Else, reply with an error.

					// Sign vote.
					if err := pm.FilePV.SignVote(pm.Config.FilePV.ChainID, req.Vote); err != nil {
						pm.Logger.Printf("[ERR] pairmint: couldn't sign vote for height %v: %v\n", req.Vote.Height, err)
						continue
					}

					// Construct proto message for SignedVoteResponse.
					res := &privvalproto.Message{
						Sum: &privvalproto.Message_SignedVoteResponse{
							SignedVoteResponse: &privvalproto.SignedVoteResponse{
								Vote: *req.Vote,
							},
						},
					}

					if _, err = protoWriter.WriteMsg(res); err != nil {
						pm.Logger.Printf("[ERR] pairmint: error while writing SignedVoteResponse: %v\n", err)
						continue
					}

				case *privvalproto.Message_SignProposalRequest:
					req := msg.GetSignProposalRequest()
					pm.Logger.Printf("[DEBUG] pairmint: Received SignProposalRequest: %v\n", req)

					// TODO: Sign proposal if node is primary, and reply with signed vote.
					// TODO: Else, reply with an error.

					// Sign proposal.
					if err := pm.FilePV.SignProposal(pm.Config.FilePV.ChainID, req.Proposal); err != nil {
						pm.Logger.Printf("[ERR] pairmint: couldn't sign proposal for height %v: %v\n", req.Proposal.Height, err)
						continue
					}

					// Construct proto message for SignProposalResponse.
					res := &privvalproto.Message{
						Sum: &privvalproto.Message_SignedProposalResponse{
							SignedProposalResponse: &privvalproto.SignedProposalResponse{
								Proposal: *req.Proposal,
							},
						},
					}

					if _, err = protoWriter.WriteMsg(res); err != nil {
						pm.Logger.Printf("[ERR] pairmint: error while writing SignedProposalResponse: %v\n", err)
						continue
					}

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
