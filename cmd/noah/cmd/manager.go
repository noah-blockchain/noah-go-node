package cmd

import (
	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	"github.com/noah-blockchain/noah-node-cli/service"
	"github.com/spf13/cobra"
)

var Manager = &cobra.Command{
	Use:   "manager",
	Short: "Noah CLI manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		service.RunCli(utils.GetNoahHome()+"/manager.sock", args)
		return nil
	},
}
