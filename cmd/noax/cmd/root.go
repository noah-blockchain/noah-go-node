package cmd

import (
	"github.com/PillarDevelopment/noax-go-node/cmd/utils"
	"github.com/PillarDevelopment/noax-go-node/config"
	"github.com/PillarDevelopment/noax-go-node/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg *config.Config

var RootCmd = &cobra.Command{
	Use:   "noah",
	Short: "Noax Go Node",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		v := viper.New()
		v.SetConfigFile(utils.GetNoaxConfigPath())
		cfg = config.GetConfig()

		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}

		if err := v.Unmarshal(cfg); err != nil {
			panic(err)
		}

		log.InitLog(cfg)
	},
}
