package cmd

import (
	"github.com/noah-blockchain/noah-go-node/cli/service"
	"github.com/noah-blockchain/noah-go-node/cmd/utils"
	"github.com/spf13/cobra"
	"strings"
)

var ManagerCommand = &cobra.Command{
	Use:                "manager",
	Short:              "Noah manager execute command",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		newArgs := setParentFlags(cmd, args)
		console, err := service.NewCLI(utils.GetNoahHome() + "/manager.sock")
		if err != nil {
			return nil
		}
		return console.Execute(newArgs)
	},
}

var ManagerConsole = &cobra.Command{
	Use:                "console",
	Short:              "Noah CLI manager",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = setParentFlags(cmd, args)
		console, err := service.NewCLI(utils.GetNoahHome() + "/manager.sock")
		if err != nil {
			return nil
		}
		console.Cli(cmd.Context())
		return nil
	},
}

func setParentFlags(cmd *cobra.Command, args []string) (newArgs []string) {
	for _, arg := range args {
		split := strings.Split(arg, "=")
		if len(split) == 2 {
			err := cmd.Parent().PersistentFlags().Set(strings.TrimLeft(split[0], "-"), split[1])
			if err == nil {
				continue
			}
		}
		newArgs = append(newArgs, arg)
	}
	return newArgs
}
