package main

import (
	"github.com/noah-blockchain/noah-go-node/cmd/noah/cmd"
	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	"github.com/noah-blockchain/noah-go-node/config"
)

func main() {
	rootCmd := cmd.RootCmd

	rootCmd.AddCommand(
		cmd.RunNode,
		cmd.ShowNodeId,
		cmd.ShowValidator,
		cmd.Version,
	)

	rootCmd.PersistentFlags().StringVar(&utils.NoahHome, "home-dir", "", "base dir (default is $HOME/noah)")
	rootCmd.PersistentFlags().StringVar(&utils.NoahConfig, "config", "", "path to config (default is $(home-dir)/config/config.toml)")
	rootCmd.PersistentFlags().StringVar(&config.NetworkId, "network-id", config.DefaultNetworkId, "network id")

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
