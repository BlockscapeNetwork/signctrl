package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
			case <-pv.TermCh: // TermCh is used for self-terminating behavior
				pv.Logger.Println("\n[INFO] signctrl: Terminating SignCTRL... (stopped)")
			case <-sigs: // The sigs channel is only used for OS interrupt signals
				pv.Logger.Println("\n[INFO] signctrl: Terminating SignCTRL... (interrupt)")
			}

			// Save rank to last_rank.json file.
			if err := pv.Save(cfgDir, pv.Logger); err != nil {
				fmt.Printf("[ERR] signctrl: Couldn't save rank to %v: %v", privval.LastRankFile, err)
				os.Exit(1)
			}

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
