package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/privval"
	"github.com/hashicorp/logutils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tm_privval "github.com/tendermint/tendermint/privval"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts the SignCTRL node",
		Run: func(cmd *cobra.Command, args []string) {
			// Load the config into memory.
			cfg, err := config.Load()
			if err != nil {
				fmt.Printf("Couldn't load %v:\n%v", config.File, err)
				os.Exit(1)
			}

			// Set the logger and its mininum log level.
			logger := log.New(os.Stderr, "", 0)
			filter := &logutils.LevelFilter{
				Levels:   config.LogLevels,
				MinLevel: logutils.LogLevel(cfg.Init.LogLevel),
				Writer:   os.Stderr,
			}
			logger.SetOutput(filter)

			// Initialize a new SCFilePV.
			cfgDir := config.Dir()
			pv := privval.NewSCFilePV(
				logger,
				cfg,
				tm_privval.LoadOrGenFilePV(
					privval.KeyFilePath(cfgDir),
					privval.StateFilePath(cfgDir),
				),
			)
			if err := pv.CheckAndLoadLastRank(cfgDir, logger); err != nil {
				fmt.Printf("Couldn't load %v: %v\n", privval.LastRankFile, err)
				os.Exit(1)
			}

			// Start the SignCTRL service.
			if err := pv.Start(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Wait either for the service itself or a system call to quit the process.
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			select {
			case <-pv.Quit(): // Used for self-induced shutdown
				pv.Logger.Println("[INFO] signctrl: Shutting SignCTRL down... ⏻ (quit)")

				// The last_rank.json should only be created for user/os interrups, so delete it if
				// the node shut itself down on its own.
				// TODO: Later, when no more shutdowns are needed, this can be removed.
				os.Remove(privval.LastRankFilePath(cfgDir))
			case <-sigs: // The sigs channel is only used for OS interrupt signals
				pv.Logger.Println("[INFO] signctrl: Shutting SignCTRL down... ⏻ (user/os interrupt)")
				pv.Stop()
			}

			// TODO: The current logger is async which is why some of the last log messages before
			// shutdown aren't printed out. Make the logger sync. For now, wait a second for everything
			// to be printed out.
			time.Sleep(time.Second)

			// Terminate the process gracefully with exit code 0.
			os.Exit(0)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(startCmd)
}

func initConfig() {
	cfgParts := strings.Split(config.File, ".")

	viper.SetConfigName(cfgParts[0])
	viper.SetConfigType(cfgParts[1])

	viper.AddConfigPath("$SIGNCTRL_CONFIG_DIR")
	viper.AddConfigPath("$HOME/.signctrl")
	viper.AddConfigPath(".")
}
