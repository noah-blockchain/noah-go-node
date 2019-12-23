package cmd

import (
	"fmt"

	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	"github.com/noah-blockchain/noah-go-node/config"
	"github.com/noah-blockchain/noah-go-node/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg *config.Config

var RootCmd = &cobra.Command{
	Use:   "noah",
	Short: "Noah Go Node",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg = config.GetConfig()

		v := viper.New()
		v.SetConfigFile(utils.GetNoahConfigPath())
		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}

		if err := v.Unmarshal(cfg); err != nil {
			panic(err)
		}

		validatorMode, err := cmd.Flags().GetBool("validator-mode")
		if err != nil {
			panic(err)
		}
		cfg.ValidatorMode = validatorMode

		if cfg.ValidatorMode {
			fmt.Println("This node working in validator mode.")
		} else {
			fmt.Println("This node working NOT in validator mode.")
		}

		fmt.Println("Current network id", config.NetworkId)
		log.InitLog(cfg)
	},
}
